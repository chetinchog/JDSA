package backend

import (
	"context"
	"encoding/csv"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// isValidJobURL validates the target URL before scraping.
func isValidJobURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("URL inválida: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("solo se permiten URLs HTTP/HTTPS")
	}
	if u.Host == "" {
		return fmt.Errorf("la URL no tiene host")
	}
	// Block localhost and private IPs
	host := u.Hostname()
	blockedHosts := []string{"localhost", "127.0.0.1", "0.0.0.0", "[::1]"}
	for _, b := range blockedHosts {
		if strings.EqualFold(host, b) {
			return fmt.Errorf("no se permiten URLs locales")
		}
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return fmt.Errorf("no se permiten URLs a redes privadas")
		}
	}
	return nil
}

// App struct
type App struct {
	ctx          context.Context
	registry     *ScraperRegistry
	cancelSearch context.CancelFunc
	cancelExport context.CancelFunc
	cancelMu     sync.Mutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		registry: NewScraperRegistry(
			NewIndeedScraper(),
		),
	}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}



// ScrapeJob extracts job information from a given URL using the appropriate scraper.
func (a *App) ScrapeJob(targetURL string) (JobData, error) {
	var job JobData

	// Validate URL before proceeding
	if err := isValidJobURL(targetURL); err != nil {
		return job, err
	}

	// Parse host to select the right scraper
	u, err := url.Parse(targetURL)
	if err != nil {
		return job, fmt.Errorf("error parsing URL: %v", err)
	}

	scraper, err := a.registry.GetScraper(u.Hostname())
	if err != nil {
		return job, err
	}

	return scraper.Scrape(targetURL)
}

// BulkScrape searches for jobs on a specific platform.
func (a *App) BulkScrape(query string, platform string, start int) (SearchResponse, error) {
	scraper, err := a.registry.GetScraper(platform)
	if err != nil {
		return SearchResponse{}, err
	}

	PreventSleep()
	defer AllowSleep()

	// Create a cancellable context so the user can stop the search
	searchCtx, cancel := context.WithCancel(a.ctx)
	a.cancelMu.Lock()
	a.cancelSearch = cancel
	a.cancelMu.Unlock()

	result, err := scraper.ScrapeSearch(searchCtx, query, start)

	a.cancelMu.Lock()
	a.cancelSearch = nil
	a.cancelMu.Unlock()

	return result, err
}

// CancelSearch cancels the currently running search.
func (a *App) CancelSearch() {
	a.cancelMu.Lock()
	defer a.cancelMu.Unlock()
	if a.cancelSearch != nil {
		a.cancelSearch()
	}
}

// CancelExport cancels the currently running export.
func (a *App) CancelExport() {
	a.cancelMu.Lock()
	defer a.cancelMu.Unlock()
	if a.cancelExport != nil {
		a.cancelExport()
	}
}

// SaveConfig saves the configuration for a platform.
func (a *App) SaveConfig(platform string, config ScraperConfig) error {
	scraper, err := a.registry.GetScraper(platform)
	if err != nil {
		return err
	}
	scraper.SetConfig(config)
	// Optionally emit an event to the frontend to save this to Firebase
	runtime.EventsEmit(a.ctx, "config-updated", map[string]interface{}{
		"platform": platform,
		"config":   config,
	})
	return nil
}

// CheckConfig returns true if the given platform has a valid config set.
func (a *App) CheckConfig(platform string) bool {
	scraper, err := a.registry.GetScraper(platform)
	if err != nil {
		return false
	}
	return scraper.HasValidConfig()
}

// GetConfig returns the current config for a platform.
func (a *App) GetConfig(platform string) (ScraperConfig, error) {
	scraper, err := a.registry.GetScraper(platform)
	if err != nil {
		return ScraperConfig{}, err
	}
	return scraper.GetConfig(), nil
}

// GetClipboardText returns the content of the system clipboard.
func (a *App) GetClipboardText() (string, error) {
	return runtime.ClipboardGetText(a.ctx)
}

// cleanScrapedText removes CSS blocks {...}, CSS classes, and HTML tags from the text
func cleanScrapedText(input string, preserveNewlines bool) string {
	// 1. Remove CSS blocks like {color: red; ...}
	reCSS := regexp.MustCompile(`\{[^\}]*\}`)
	input = reCSS.ReplaceAllString(input, "")

	// 2. Remove Indeed's dynamic CSS classes (e.g., .css-1jtd2m7 or just css-1jtd2m7)
	reCSSClass := regexp.MustCompile(`\.?css-[a-zA-Z0-9]+`)
	input = reCSSClass.ReplaceAllString(input, "")

	// 3. Remove HTML tags if any were captured
	reHTML := regexp.MustCompile(`<[^>]*>`)
	input = reHTML.ReplaceAllString(input, "")

	// 4. Fix common entities
	input = strings.ReplaceAll(input, "&nbsp;", " ")
	input = strings.ReplaceAll(input, "&amp;", "&")
	input = strings.ReplaceAll(input, "&lt;", "<")
	input = strings.ReplaceAll(input, "&gt;", ">")

	// 5. Normalizar espacios sin romper saltos de línea
	if !preserveNewlines {
		reAllSpaces := regexp.MustCompile(`\s+`)
		input = reAllSpaces.ReplaceAllString(input, " ")
	} else {
		// Solo colapsar múltiples espacios horizontales, mantener saltos de línea
		reHorizontalSpaces := regexp.MustCompile(`[ \t\r\f]+`)
		input = reHorizontalSpaces.ReplaceAllString(input, " ")
	}

	return strings.TrimSpace(input)
}

// cleanCompanyName applies specialized cleanup to a company name scraped from Indeed.
// Indeed injects CSS pseudo-selectors, media queries, and attribute selectors into the
// text content of the company name node (e.g. `:visited:hover@media(...)[dir="rtl"] svg`).
// This function strips those artifacts on top of the generic cleanScrapedText pass.
func cleanCompanyName(input string) string {
	// First apply generic cleanup (CSS blocks, CSS classes, HTML tags, entities, whitespace)
	input = cleanScrapedText(input, false)

	// Remove CSS pseudo-selectors chains (e.g. :visited, :hover, :focus-visible, :active, :focus)
	rePseudo := regexp.MustCompile(`:[a-zA-Z-]+`)
	input = rePseudo.ReplaceAllString(input, "")

	// Remove CSS attribute selectors (e.g. [dir="rtl"] or [dir=""rtl""])
	reAttr := regexp.MustCompile(`\[[^\]]*\]`)
	input = reAttr.ReplaceAllString(input, "")

	// Remove media queries (e.g. @media (prefers-reduced-motion: reduce))
	reMedia := regexp.MustCompile(`@media[^)]*\)?`)
	input = reMedia.ReplaceAllString(input, "")

	// Remove any remaining lone @ characters
	input = strings.ReplaceAll(input, "@", "")

	// Remove stray unbalanced braces that cleanScrapedText leaves behind
	// (cleanScrapedText only removes matched {…} pairs)
	input = strings.ReplaceAll(input, "{", "")
	input = strings.ReplaceAll(input, "}", "")

	// Remove SVG/CSS keyword prefixes that can appear immediately glued to the company name
	// (e.g. "svgVerisure" → "Verisure"). Must run AFTER whitespace collapse so the prefix
	// is at the start of a word boundary with the real name or a whitespace boundary.
	reLeadingSVG := regexp.MustCompile(`(?i)\bsvg([A-Z])`)
	input = reLeadingSVG.ReplaceAllString(input, "$1")

	// Also remove standalone svg/rtl/ltr tokens surrounded by whitespace
	reKeywords := regexp.MustCompile(`(?i)\b(svg|rtl|ltr)\b`)
	input = reKeywords.ReplaceAllString(input, "")

	// Collapse any extra whitespace introduced by the removals above
	reSpaces := regexp.MustCompile(`\s+`)
	input = reSpaces.ReplaceAllString(input, " ")

	return strings.TrimSpace(input)
}

func parseLocation(loc string) string {
	text := strings.ToLower(loc)
	
	switch {
	// Check explicit remote variants that map to US first
	case strings.Contains(text, "remote-us"), strings.Contains(text, "remote us"), strings.Contains(text, "remote - us"):
		return "US"

	// General remote
	case strings.Contains(text, "anyware"), strings.Contains(text, "anywhere"),
		strings.Contains(text, "homeoffice"), strings.Contains(text, "home office"),
		strings.Contains(text, "remote"), strings.Contains(text, "remoto"),
		strings.Contains(text, "desde casa"):
		return "AR"

	// User mapped cases
	case strings.Contains(text, "alabama"), strings.Contains(text, "alaska"), strings.Contains(text, "arizona"),
		strings.Contains(text, "arkansas"), strings.Contains(text, "california"), strings.Contains(text, "colorado"),
		strings.Contains(text, "connecticut"), strings.Contains(text, "delaware"), strings.Contains(text, "florida"),
		strings.Contains(text, "georgia"), strings.Contains(text, "hawaii"), strings.Contains(text, "idaho"),
		strings.Contains(text, "illinois"), strings.Contains(text, "indiana"), strings.Contains(text, "iowa"),
		strings.Contains(text, "kansas"), strings.Contains(text, "kentucky"), strings.Contains(text, "louisiana"),
		strings.Contains(text, "maine"), strings.Contains(text, "maryland"), strings.Contains(text, "massachusetts"),
		strings.Contains(text, "michigan"), strings.Contains(text, "minnesota"), strings.Contains(text, "mississippi"),
		strings.Contains(text, "missouri"), strings.Contains(text, "montana"), strings.Contains(text, "nebraska"),
		strings.Contains(text, "nevada"), strings.Contains(text, "new hampshire"), strings.Contains(text, "new jersey"),
		strings.Contains(text, "new mexico"), strings.Contains(text, "new york"), strings.Contains(text, "nueva york"),
		strings.Contains(text, "north carolina"), strings.Contains(text, "north dakota"), strings.Contains(text, "ohio"),
		strings.Contains(text, "oklahoma"), strings.Contains(text, "oregon"), strings.Contains(text, "pennsylvania"),
		strings.Contains(text, "rhode island"), strings.Contains(text, "south carolina"), strings.Contains(text, "south dakota"),
		strings.Contains(text, "tennessee"), strings.Contains(text, "texas"), strings.Contains(text, "utah"),
		strings.Contains(text, "vermont"), strings.Contains(text, "virginia"), strings.Contains(text, "washington"),
		strings.Contains(text, "west virginia"), strings.Contains(text, "wisconsin"), strings.Contains(text, "wyoming"),
		strings.Contains(text, "estados unidos"), strings.Contains(text, "ee. uu."), strings.Contains(text, "ee uu"),
		strings.Contains(text, "usa"), strings.Contains(text, "austin"), strings.Contains(text, "los ángeles"),
		strings.Contains(text, "los angeles"), strings.Contains(text, "eden prairie"), strings.Contains(text, "jersey"),
		strings.Contains(text, " dallas"), strings.Contains(text, " houston"), strings.Contains(text, " chicago"),
		strings.Contains(text, " atlanta"), strings.Contains(text, " phoenix"), strings.Contains(text, " boston"),
		strings.Contains(text, " denver"), strings.Contains(text, " seattle"), strings.Contains(text, " san francisco"),
		strings.Contains(text, ", tx"), strings.Contains(text, ", ca"), strings.Contains(text, ", ny"), strings.Contains(text, ", fl"),
		strings.Contains(text, ", ga"), strings.Contains(text, ", il"), strings.Contains(text, ", pa"), strings.Contains(text, ", oh"),
		strings.Contains(text, ", mi"), strings.Contains(text, ", nc"), strings.Contains(text, ", nj"), strings.Contains(text, ", va"),
		strings.Contains(text, ", co"), strings.Contains(text, ", wa"), strings.Contains(text, ", ma"), strings.Contains(text, ", md"),
		strings.Contains(text, ", or"), strings.Contains(text, ", az"), strings.Contains(text, ", ut"),
		strings.Contains(text, ", mn"), strings.Contains(text, ", wi"), strings.Contains(text, ", tn"), strings.Contains(text, ", sc"),
		strings.Contains(text, ", ky"), strings.Contains(text, ", nv"), strings.Contains(text, ", ok"):
		return "US"
	case strings.Contains(text, "ontario"), strings.Contains(text, "columbia británica"), strings.Contains(text, "british columbia"),
		strings.Contains(text, "vancouver"), strings.Contains(text, "calgary"), strings.Contains(text, "montreal"),
		strings.Contains(text, "canadá"), strings.Contains(text, "canada"):
		return "CA"
	case strings.Contains(text, "londres"), strings.Contains(text, "london"), strings.Contains(text, "reino unido"), strings.Contains(text, "uk"),
		strings.Contains(text, "manchester"), strings.Contains(text, "birmingham"), strings.Contains(text, "edinburgh"):
		return "GB"
	case strings.Contains(text, "ámsterdam"), strings.Contains(text, "amsterdam"), strings.Contains(text, "países bajos"), strings.Contains(text, "netherlands"), strings.Contains(text, "la haya"), strings.Contains(text, "the hague"), strings.Contains(text, "neerlandés"), strings.Contains(text, "dutch"):
		return "NL"
	case strings.Contains(text, "brno"), strings.Contains(text, "pilsenský kraj"), strings.Contains(text, "república checa"), strings.Contains(text, "czech"):
		return "CZ"
	case strings.Contains(text, "zug"), strings.Contains(text, "suiza"), strings.Contains(text, "switzerland"):
		return "CH"
	case strings.Contains(text, "irlanda"), strings.Contains(text, "ireland"), strings.Contains(text, "dublín"), strings.Contains(text, "dublin"):
		return "IE"
	case strings.Contains(text, "malta"), strings.Contains(text, "santa venera"):
		return "MT"
	case strings.Contains(text, "austria"), strings.Contains(text, "klagenfurt"), strings.Contains(text, "graz"), strings.Contains(text, "viena"), strings.Contains(text, "vienna"):
		return "AT"
	case strings.Contains(text, "ucrania"), strings.Contains(text, "ukraine"), strings.Contains(text, "kiev"), strings.Contains(text, "kyiv"):
		return "UA"
	case strings.Contains(text, "andorra"):
		return "AD"
	case strings.Contains(text, "alemania"), strings.Contains(text, "germany"), strings.Contains(text, "berlín"), strings.Contains(text, "berlin"), strings.Contains(text, "múnich"), strings.Contains(text, "munich"),
		strings.Contains(text, "gmbh"), strings.Contains(text, "analogue insight"):
		return "DE"
	case strings.Contains(text, "francia"), strings.Contains(text, "france"), strings.Contains(text, "parís"), strings.Contains(text, "paris"),
		strings.Contains(text, "eviden"):
		return "FR"
	case strings.Contains(text, "españa"), strings.Contains(text, "spain"), strings.Contains(text, "madrid"), strings.Contains(text, "barcelona"):
		return "ES"
	case strings.Contains(text, "italia"), strings.Contains(text, "italy"), strings.Contains(text, "roma"), strings.Contains(text, "rome"):
		return "IT"
	case strings.Contains(text, "mexico"), strings.Contains(text, "méxico"):
		return "MX"
	case strings.Contains(text, "argentina"), strings.Contains(text, "buenos aires"):
		return "AR"
	case strings.Contains(text, "brasil"), strings.Contains(text, "brazil"), strings.Contains(text, "são paulo"),
		strings.Contains(text, " br "), strings.Contains(text, ": br"), strings.Contains(text, ", br"):
		return "BR"
	case strings.Contains(text, "filipinas"), strings.Contains(text, "philippines"), strings.Contains(text, "manila"),
		strings.Contains(text, "cagayan de oro"):
		return "PH"
	case strings.Contains(text, "israel"), strings.Contains(text, "tel aviv"), strings.Contains(text, "jerusalén"), strings.Contains(text, "jerusalem"):
		return "IL"
	case strings.Contains(text, "india"), strings.Contains(text, "bangalore"), strings.Contains(text, "bengaluru"), strings.Contains(text, "mumbai"), strings.Contains(text, "delhi"),
		strings.Contains(text, "odisha"), strings.Contains(text, "uttar pradesh"):
		return "IN"
	case strings.Contains(text, "honduras"), strings.Contains(text, "tegucigalpa"), strings.Contains(text, "francisco morazán"):
		return "HN"
	case strings.Contains(text, "guatemala"):
		return "GT"
	case strings.Contains(text, "nicaragua"):
		return "NI"
	case strings.Contains(text, "costa rica"):
		return "CR"
	case strings.Contains(text, "panamá"), strings.Contains(text, "panama"):
		return "PA"
	case strings.Contains(text, "el salvador"):
		return "SV"
	case strings.Contains(text, "república dominicana"), strings.Contains(text, "dominican republic"):
		return "DO"
	case strings.Contains(text, "ecuador"), strings.Contains(text, "quito"), strings.Contains(text, "guayaquil"):
		return "EC"
	case strings.Contains(text, "venezuela"), strings.Contains(text, "caracas"):
		return "VE"
	case strings.Contains(text, "perú"), strings.Contains(text, "peru"), strings.Contains(text, "lima"):
		return "PE"
	case strings.Contains(text, "bolivia"), strings.Contains(text, "la paz"):
		return "BO"
	case strings.Contains(text, "paraguay"), strings.Contains(text, "asunción"):
		return "PY"
	case strings.Contains(text, "puerto rico"):
		return "PR"
	case strings.Contains(text, "singapur"), strings.Contains(text, "singapore"):
		return "SG"
	case strings.Contains(text, "japón"), strings.Contains(text, "japan"), strings.Contains(text, "tokio"), strings.Contains(text, "tokyo"):
		return "JP"
	case strings.Contains(text, "australia"), strings.Contains(text, "sydney"), strings.Contains(text, "melbourne"):
		return "AU"
	case strings.Contains(text, "nueva zelanda"), strings.Contains(text, "new zealand"), strings.Contains(text, "auckland"):
		return "NZ"
	case strings.Contains(text, "chile"), strings.Contains(text, "santiago"):
		return "CL"
	case strings.Contains(text, "colombia"), strings.Contains(text, "bogotá"), strings.Contains(text, "medellín"):
		return "CO"
	case strings.Contains(text, "uruguay"), strings.Contains(text, "montevideo"):
		return "UY"
	case strings.Contains(text, "estonia"), strings.Contains(text, "tallinn"), strings.Contains(text, "tallin"):
		return "EE"
	case strings.Contains(text, "polonia"), strings.Contains(text, "poland"), strings.Contains(text, "warsaw"), strings.Contains(text, "varsovia"):
		return "PL"
	case strings.Contains(text, "bélgica"), strings.Contains(text, "belgium"), strings.Contains(text, "brussels"), strings.Contains(text, "bruselas"):
		return "BE"
	case strings.Contains(text, "finlandia"), strings.Contains(text, "finland"), strings.Contains(text, "helsinki"):
		return "FI"
	case strings.Contains(text, "sudáfrica"), strings.Contains(text, "south africa"), strings.Contains(text, "cape town"), strings.Contains(text, "johannesburg"):
		return "ZA"
	case strings.Contains(text, "líbano"), strings.Contains(text, "lebanon"):
		return "LB"
	case strings.Contains(text, "dinamarca"), strings.Contains(text, "denmark"), strings.Contains(text, "creativeforce.team"):
		return "DK"
	case strings.Contains(text, "chipre"), strings.Contains(text, "cyprus"), strings.Contains(text, "paphos"):
		return "CY"
	case strings.Contains(text, "european union"), strings.Contains(text, "unión europea"), strings.Contains(text, " eu "):
		return "EU"
	case strings.Contains(text, "tailandia"), strings.Contains(text, "thailand"):
		return "TH"
	case strings.Contains(text, "indonesia"), strings.Contains(text, "bali"), strings.Contains(text, "jakarta"):
		return "ID"
	case strings.Contains(text, "vietnam"), strings.Contains(text, "viet nam"):
		return "VN"
	case strings.Contains(text, "egipto"), strings.Contains(text, "egypt"):
		return "EG"
	case strings.Contains(text, "china"), strings.Contains(text, "bamboo-works.com"):
		return "CN"
	case strings.Contains(text, "north america"), strings.Contains(text, "norteamérica"):
		return "US"
	case strings.Contains(text, "wolfsburg"):
		return "DE"
	case strings.Contains(text, "redwood city"), strings.Contains(text, "san francisco"), strings.Contains(text, "seattle"), strings.Contains(text, "salt lake city"), strings.Contains(text, "dallas"), strings.Contains(text, "new york"):
		return "US"
	case strings.Contains(text, "scoutsolutions.net"), strings.Contains(text, "quinstreet.com"), strings.Contains(text, "theclassconsultinggroup.org"), strings.Contains(text, "intro.io"), strings.Contains(text, "quinstreet"),
		strings.Contains(text, "amgen"):
		return "US"
	case strings.Contains(text, "codersbrain.com"):
		return "IN"
	case strings.Contains(text, "canonical"), strings.Contains(text, "adria solutions"), strings.Contains(text, "trt solutions"):
		return "GB"
	case strings.Contains(text, "serbia"), strings.Contains(text, "belgrade"), strings.Contains(text, "belgrado"):
		return "RS"
	case strings.Contains(text, "grecia"), strings.Contains(text, "greece"), strings.Contains(text, "atenas"), strings.Contains(text, "athens"):
		return "GR"
	case strings.Contains(text, "turquía"), strings.Contains(text, "turkey"), strings.Contains(text, "estambul"), strings.Contains(text, "istanbul"):
		return "TR"
	case strings.Contains(text, "united states"), strings.Contains(text, "u.s.a"), strings.Contains(text, "u.s."), strings.Contains(text, "ee.uu"), strings.Contains(text, "eeuu"), strings.Contains(text, "estados unidos"):
		return "US"
	case strings.Contains(text, "united kingdom"), strings.Contains(text, "reino unido"), strings.Contains(text, "uk"):
		return "GB"
	case strings.Contains(text, "suecia"), strings.Contains(text, "sweden"):
		return "SE"
	case strings.Contains(text, "noruega"), strings.Contains(text, "norway"):
		return "NO"
	}

	// Default to AR if no recognizable country found since most scrapes happen on ar.indeed.com
	return "AR"
}

func writeJobToCSV(writer *csv.Writer, job JobData) error {
	providerID := "Indeed"
	expirationDate := time.Now().AddDate(0, 0, 365).Format("2006-01-02")
	paisID := parseLocation(job.Location)
	record := []string{
		providerID,
		expirationDate,
		paisID,
		job.JobID,
		job.CompanyName,
		job.JobTitle,
		job.JobDescription,
		job.ApplyURL,
		job.Location,
	}
	return writer.Write(record)
}

// ExportCSV saves the job data as a CSV file
func (a *App) ExportCSV(job JobData) error {
	now := time.Now().Format("200601021504")
	jobIDStr := job.JobID
	if jobIDStr == "" {
		jobIDStr = "UNKNOWN"
	}
	filename := fmt.Sprintf("JOB_%s_%s.csv", now, jobIDStr)

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: filename,
		Title:           "Exportar Información de Empleo",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
		},
	})
	if err != nil {
		return err
	}
	if path == "" {
		// User cancelled the dialog
		return nil
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// UTF-8 BOM helps Excel render characters correctly
	file.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"provider_id", "expiration_date", "pais_id", "job_id", "company_name", "title", "description", "apply_url", "location"}
	if err := writer.Write(header); err != nil {
		return err
	}

	return writeJobToCSV(writer, job)
}

// GetDebugHTML returns the last scraped HTML string and its URL for UI viewing.
func (a *App) GetDebugHTML(platform string) (map[string]string, error) {
	if platform == "" {
		platform = "indeed"
	}
	scraper, err := a.registry.GetScraper(platform)
	if err != nil {
		return nil, err
	}

	htmlContent, urlStr := scraper.GetLastDebugHTML()
	if len(htmlContent) == 0 {
		return nil, fmt.Errorf("no hay contenido HTML de diagnóstico almacenado para %s", platform)
	}

	return map[string]string{
		"html": string(htmlContent),
		"url":  urlStr,
	}, nil
}

// OpenURL opens the given URL in the default system browser
func (a *App) OpenURL(targetURL string) {
	runtime.BrowserOpenURL(a.ctx, targetURL)
}

// ExportDebugHTML saves the last scraped HTML content to a file.
func (a *App) ExportDebugHTML(platform string) (string, error) {
	if platform == "" {
		platform = "indeed"
	}
	scraper, err := a.registry.GetScraper(platform)
	if err != nil {
		return "", err
	}

	htmlContent, urlStr := scraper.GetLastDebugHTML()
	if len(htmlContent) == 0 {
		return "", fmt.Errorf("no hay contenido HTML de diagnóstico almacenado para %s", platform)
	}

	now := time.Now().Format("20060102150405")
	defaultName := fmt.Sprintf("debug_%s_%s.html", strings.ToLower(platform), now)

	title := "Guardar HTML de Diagnóstico"
	if urlStr != "" {
		title = fmt.Sprintf("Guardar HTML de Diagnóstico (%s)", urlStr)
	}

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultName,
		Title:           title,
		Filters: []runtime.FileFilter{
			{DisplayName: "HTML Files (*.html)", Pattern: "*.html"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}

	if err := os.WriteFile(path, htmlContent, 0644); err != nil {
		return "", fmt.Errorf("error al guardar archivo HTML: %v", err)
	}

	return path, nil
}

// ExportBulkCSV saves the list of jobs found in search as a CSV file, fetching full details for each
// If existingFilePath is provided, it skips the file dialog and appends to this existing file instead.
func (a *App) ExportBulkCSV(query string, platform string, results []SearchResult, existingFilePath string) (ExportResult, error) {
	var finalRes ExportResult
	now := time.Now().Format("200601021504")

	PreventSleep()
	defer AllowSleep()

	exportCtx, cancel := context.WithCancel(a.ctx)
	a.cancelMu.Lock()
	a.cancelExport = cancel
	a.cancelMu.Unlock()

	defer func() {
		a.cancelMu.Lock()
		a.cancelExport = nil
		a.cancelMu.Unlock()
	}()

	// Format strategy name
	strategy := strings.ToUpper(platform)
	if strategy == "" {
		strategy = "UNKNOWN"
	}

	strQuery := strings.ToUpper(query)
	re := regexp.MustCompile(`[^A-Z0-9]+`)
	strQuery = re.ReplaceAllString(strQuery, "_")
	strQuery = strings.Trim(strQuery, "_")
	if strQuery == "" {
		strQuery = "EMPLEOS"
	}
	defaultName := fmt.Sprintf("BULK_%s_%s_%s.csv", now, strategy, strQuery)

	path := existingFilePath

	if path == "" {
		var err error
		path, err = runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
			DefaultFilename: defaultName,
			Title:           "Exportar Lista Completa de Empleos",
			Filters: []runtime.FileFilter{
				{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
			},
		})
		if err != nil {
			finalRes.Errors = []string{err.Error()}
			return finalRes, err
		}
		if path == "" {
			return finalRes, nil
		}
	}

	finalRes.FilePath = path

	// If appending to an existing list, just open and start appending. 
	// We no longer unmarshal since it's CSV.
	var file *os.File
	var writer *csv.Writer
	var isNewFile bool

	if existingFilePath != "" {
		// Append mode
		var err error
		file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			// If it doesn't exist, we will create it (shouldn't happen typically but just in case)
			file, err = os.Create(path)
			isNewFile = true
			if err != nil {
				finalRes.Errors = []string{"Error opening file: " + err.Error()}
				return finalRes, err
			}
		}
	} else {
		// Create new file
		var err error
		file, err = os.Create(path)
		isNewFile = true
		if err != nil {
			finalRes.Errors = []string{"Error creating file: " + err.Error()}
			return finalRes, err
		}
	}
	defer file.Close()

	if isNewFile {
		// UTF-8 BOM
		file.Write([]byte{0xEF, 0xBB, 0xBF})
	}

	writer = csv.NewWriter(file)
	defer writer.Flush()

	if isNewFile {
		header := []string{"provider_id", "expiration_date", "pais_id", "job_id", "company_name", "title", "description", "apply_url", "location"}
		writer.Write(header)
	}
	scraper, err := a.registry.GetScraper("indeed.com")
	if err != nil {
		errStr := "Error initializing scraper: " + err.Error()
		finalRes.Errors = []string{errStr}
		return finalRes, fmt.Errorf("%s", errStr)
	}

	total := len(results)
	successCount := 0
	errorCount := 0
	var errorDetails []string

	Loop:
	for i, res := range results {
		select {
		case <-exportCtx.Done():
			break Loop
		default:
		}

		jobURL := fmt.Sprintf("https://ar.indeed.com/viewjob?jk=%s", res.JobID)

		jobData, scrapeErr := scraper.Scrape(jobURL)

		if scrapeErr == nil && jobData.JobTitle != "" {
			if writeErr := writeJobToCSV(writer, jobData); writeErr != nil {
				errorCount++
				errorDetails = append(errorDetails, fmt.Sprintf("- %s (ID: %s): Error writing to CSV: %s", res.Title, res.JobID, writeErr.Error()))
			} else {
				writer.Flush() // Flush immediately to see changes continuously
				successCount++
			}
		} else {
			errorCount++
			errMsg := "Error leyendo título/descripción"
			if scrapeErr != nil {
				errMsg = scrapeErr.Error()
			}
			errorDetails = append(errorDetails, fmt.Sprintf("- %s (ID: %s): %s", res.Title, res.JobID, errMsg))
		}

		// Emit progress with live success/error counts
		runtime.EventsEmit(a.ctx, "export-progress", map[string]interface{}{
			"current": i + 1,
			"total":   total,
			"success": successCount,
			"errors":  errorCount,
		})

		// Sleep between requests dynamically
		config := scraper.GetConfig()
		waitMin := config.WaitJobMin * 1000
		waitMax := config.WaitJobMax * 1000
		if waitMin == 0 && waitMax == 0 {
			waitMin = 10000
			waitMax = 15000
		} else if waitMax < waitMin {
			waitMax = waitMin + 1
		}
		sleepTime := waitMin
		if waitMax > waitMin {
			sleepTime += rand.Intn(waitMax - waitMin + 1)
		}
		
		if i < len(results)-1 {
			select {
			case <-exportCtx.Done():
				break Loop
			case <-time.After(time.Duration(sleepTime) * time.Millisecond):
			}
		}
	}

	finalRes.SuccessCount = successCount
	finalRes.ErrorCount = errorCount
	finalRes.Errors = errorDetails

	time.Sleep(500 * time.Millisecond)
	runtime.EventsEmit(a.ctx, "export-finished", finalRes)
	return finalRes, nil
}
