package backend

import (
	"context"
	"encoding/json"
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

func (a *App) LoadCookiesFromFrontend(platform, cookie string) {
	if s, err := a.registry.GetScraper(platform); err == nil {
		s.SetSessionCookie(cookie)
	}
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

// SetSessionCookie sets the manual session cookie for a platform.
func (a *App) SetSessionCookie(platform string, cookie string) error {
	scraper, err := a.registry.GetScraper(platform)
	if err != nil {
		return err
	}
	scraper.SetSessionCookie(cookie)
	// Optionally emit an event to the frontend to save this to Firebase
	runtime.EventsEmit(a.ctx, "cookie-updated", map[string]string{
		"platform": platform,
		"cookie":   cookie,
	})
	return nil
}

// CheckSessionCookie returns true if the given platform has a cookie set.
func (a *App) CheckSessionCookie(platform string) bool {
	scraper, err := a.registry.GetScraper(platform)
	if err != nil {
		return false
	}
	return scraper.HasSessionCookie()
}

// GetSessionCookie returns the current session cookie for a platform.
func (a *App) GetSessionCookie(platform string) (string, error) {
	scraper, err := a.registry.GetScraper(platform)
	if err != nil {
		return "", err
	}
	return scraper.GetSessionCookie(), nil
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

// ExportJSON saves the job data as a JSON file
func (a *App) ExportJSON(job JobData) error {
	now := time.Now().Format("200601021504")
	jobIDStr := job.JobID
	if jobIDStr == "" {
		jobIDStr = "UNKNOWN"
	}
	filename := fmt.Sprintf("JOB_%s_%s.json", now, jobIDStr)

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: filename,
		Title:           "Exportar Información de Empleo",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil {
		return err
	}
	if path == "" {
		// User cancelled the dialog
		return nil
	}

	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// OpenURL opens the given URL in the default system browser
func (a *App) OpenURL(targetURL string) {
	runtime.BrowserOpenURL(a.ctx, targetURL)
}

// ExportBulkJSON saves the list of jobs found in search as a JSON file, fetching full details for each
// If existingFilePath is provided, it skips the file dialog and appends to this existing file instead.
func (a *App) ExportBulkJSON(query string, platform string, results []SearchResult, existingFilePath string) (ExportResult, error) {
	var finalRes ExportResult
	now := time.Now().Format("200601021504")

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
	defaultName := fmt.Sprintf("BULK_%s_%s_%s.json", now, strategy, strQuery)

	path := existingFilePath

	if path == "" {
		var err error
		path, err = runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
			DefaultFilename: defaultName,
			Title:           "Exportar Lista Completa de Empleos",
			Filters: []runtime.FileFilter{
				{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
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

	// Fetch full details for each job
	var fullJobs []JobData

	// If appending to an existing file, read the existing data first
	if existingFilePath != "" {
		existingData, loadErr := os.ReadFile(existingFilePath)
		if loadErr == nil {
			// Ignore error on unmarshal - if it fails to parse, we just overwrite/append to an empty list
			_ = json.Unmarshal(existingData, &fullJobs)
		}
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

	for i, res := range results {
		jobURL := fmt.Sprintf("https://ar.indeed.com/viewjob?jk=%s", res.JobID)

		jobData, scrapeErr := scraper.Scrape(jobURL)

		if scrapeErr == nil && jobData.JobTitle != "" {
			fullJobs = append(fullJobs, jobData)
			successCount++
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

		// Sleep between requests (1s to 2s)
		time.Sleep(time.Duration(1000+rand.Intn(1000)) * time.Millisecond)
	}

	finalRes.SuccessCount = successCount
	finalRes.ErrorCount = errorCount
	finalRes.Errors = errorDetails

	if len(fullJobs) > 0 {
		data, marshalErr := json.MarshalIndent(fullJobs, "", "  ")
		if marshalErr != nil {
			finalRes.Errors = append(finalRes.Errors, "Error al procesar JSON: "+marshalErr.Error())
			runtime.EventsEmit(a.ctx, "export-finished", finalRes)
			return finalRes, marshalErr
		}

		if writeErr := os.WriteFile(path, data, 0644); writeErr != nil {
			finalRes.Errors = append(finalRes.Errors, "Error al escribir archivo: "+writeErr.Error())
			runtime.EventsEmit(a.ctx, "export-finished", finalRes)
			return finalRes, writeErr
		}
	}

	time.Sleep(500 * time.Millisecond)
	runtime.EventsEmit(a.ctx, "export-finished", finalRes)
	return finalRes, nil
}
