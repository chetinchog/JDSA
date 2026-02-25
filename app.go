package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ScrapeJob extracts job information from a given URL
func (a *App) ScrapeJob(targetURL string) (JobData, error) {
	var job JobData
	job.ApplyURL = targetURL

	// Parse Job ID from URL
	u, err := url.Parse(targetURL)
	if err == nil {
		job.JobID = u.Query().Get("jk")
	}

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"),
	)

	// Set realistic headers
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		r.Headers.Set("Accept-Language", "es-AR,es;q=0.9,en;q=0.8")
		r.Headers.Set("Cache-Control", "max-age=0")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Sec-Ch-Ua", `"Not A(Bit:Major";v="99", "Google Chrome";v="121", "Chromium";v="121"`)
		r.Headers.Set("Sec-Ch-Ua-Mobile", "?0")
		r.Headers.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "none")
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
	})

	// Job Title Selectors
	c.OnHTML("h1.jobsearch-JobInfoHeader-title", func(e *colly.HTMLElement) {
		job.JobTitle = strings.TrimSpace(e.Text)
	})
	c.OnHTML("[data-testid='jobsearch-JobInfoHeader-title']", func(e *colly.HTMLElement) {
		if job.JobTitle == "" {
			job.JobTitle = strings.TrimSpace(e.Text)
		}
	})

	// Company Name Selectors (Prioritize UI)
	c.OnHTML("[data-testid='inlineHeader-companyName']", func(e *colly.HTMLElement) {
		job.CompanyName = strings.TrimSpace(e.Text)
	})
	c.OnHTML("[data-testid='inline-companyname-link']", func(e *colly.HTMLElement) {
		if job.CompanyName == "" {
			job.CompanyName = strings.TrimSpace(e.Text)
		}
	})
	c.OnHTML("[data-testid='inline-companyname']", func(e *colly.HTMLElement) {
		if job.CompanyName == "" {
			job.CompanyName = strings.TrimSpace(e.Text)
		}
	})

	// Location Selectors (Prioritize UI for full strings like "Buenos Aires, Buenos Aires")
	c.OnHTML("[data-testid='jobsearch-JobInfoHeader-companyLocation']", func(e *colly.HTMLElement) {
		job.Location = strings.TrimSpace(e.Text)
	})
	c.OnHTML("[data-testid='job-location']", func(e *colly.HTMLElement) {
		if job.Location == "" {
			job.Location = strings.TrimSpace(e.Text)
		}
	})
	c.OnHTML("[data-testid='text-location']", func(e *colly.HTMLElement) {
		if job.Location == "" {
			job.Location = strings.TrimSpace(e.Text)
		}
	})

	// JSON-LD Fallback (Reliable for Title and Description, but can have codes like "B" for region)
	c.OnHTML("script[type='application/ld+json']", func(e *colly.HTMLElement) {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(e.Text), &data); err == nil {
			if title, ok := data["title"].(string); ok && job.JobTitle == "" {
				job.JobTitle = title
			}
			if org, ok := data["hiringOrganization"].(map[string]interface{}); ok {
				if name, ok := org["name"].(string); ok && job.CompanyName == "" {
					job.CompanyName = name
				}
			}
			if loc, ok := data["jobLocation"].(map[string]interface{}); ok {
				if addr, ok := loc["address"].(map[string]interface{}); ok {
					locality, _ := addr["addressLocality"].(string)
					region, _ := addr["addressRegion"].(string)

					// Map "B" to "Buenos Aires" if it's Indeed's shorthand
					if region == "B" {
						region = "Buenos Aires"
					}

					if locality != "" && region != "" {
						if job.Location == "" {
							job.Location = fmt.Sprintf("%s, %s", locality, region)
						}
					} else if locality != "" {
						if job.Location == "" {
							job.Location = locality
						}
					}
				}
			}
			if desc, ok := data["description"].(string); ok && job.JobDescription == "" {
				// Sometimes LD+JSON has the description too
				job.JobDescription = desc
			}
		}
	})

	// Job Description
	c.OnHTML("#jobDescriptionText", func(e *colly.HTMLElement) {
		if job.JobDescription == "" {
			job.JobDescription = strings.TrimSpace(e.Text)
		}
	})

	err = c.Visit(targetURL)
	if err != nil {
		return job, fmt.Errorf("error visiting URL: %v", err)
	}

	// Post-processing cleanup
	job.CompanyName = cleanScrapedText(job.CompanyName, false)
	job.JobDescription = cleanScrapedText(job.JobDescription, true)

	return job, nil
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
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: fmt.Sprintf("job_%s.json", job.JobID),
		Title:           "Exportar Información de Empleo",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil || path == "" {
		return err
	}

	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
