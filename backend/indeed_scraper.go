package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// IndeedScraper handles scraping job data from Indeed job pages.
type IndeedScraper struct {
	jar http.CookieJar
}

// NewIndeedScraper creates a new IndeedScraper with a cookie jar.
func NewIndeedScraper() *IndeedScraper {
	jar, _ := cookiejar.New(nil)
	return &IndeedScraper{jar: jar}
}

// CanHandle returns true if the host contains "indeed".
func (s *IndeedScraper) CanHandle(host string) bool {
	return strings.Contains(strings.ToLower(host), "indeed")
}

// Scrape extracts job data from an Indeed URL.
// Supports both /viewjob?jk=... and /jobs?...&vjk=... URL formats.
func (s *IndeedScraper) Scrape(targetURL string) (JobData, error) {
	var job JobData
	job.ApplyURL = targetURL

	// Parse URL to extract job ID and determine the scrape target
	u, err := url.Parse(targetURL)
	if err != nil {
		return job, fmt.Errorf("error parsing URL: %v", err)
	}

	// Try jk first, then vjk
	jobID := u.Query().Get("jk")
	if jobID == "" {
		jobID = u.Query().Get("vjk")
	}
	job.JobID = jobID

	// If the ID came from vjk (search result page), build the canonical viewjob URL
	scrapeURL := targetURL
	if u.Query().Get("jk") == "" && jobID != "" {
		scrapeURL = fmt.Sprintf("%s://%s/viewjob?jk=%s", u.Scheme, u.Host, jobID)
	}

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"),
		colly.MaxDepth(1),
	)
	c.SetCookieJar(s.jar)
	c.SetRequestTimeout(30 * time.Second)

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
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("Referer", "https://ar.indeed.com/")
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

	// Location Selectors
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

	// Expired job detection — Indeed shows specific elements for expired posts
	c.OnHTML("[data-testid='expired-job']", func(e *colly.HTMLElement) {
		job.IsExpired = true
	})
	c.OnHTML(".jobsearch-JobInfoHeader-expiredHeader", func(e *colly.HTMLElement) {
		job.IsExpired = true
	})
	// Indeed AR specific selectors or general alert boxes
	c.OnHTML(".ekqvxqv5", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "caducó") || strings.Contains(e.Text, "expired") {
			job.IsExpired = true
		}
	})
	c.OnHTML(".icl-Callout--critical", func(e *colly.HTMLElement) {
		job.IsExpired = true
	})

	// Meta robots noindex is a strong signal for expired/inactive jobs on Indeed
	c.OnHTML("meta", func(e *colly.HTMLElement) {
		content := e.Attr("content")
		if strings.Contains(strings.ToLower(content), "noindex") {
			job.IsExpired = true
		}
	})

	// JSON-LD Fallback
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
				job.JobDescription = desc
			}
			if jobStatus, ok := data["jobStatus"].(string); ok {
				jsLower := strings.ToLower(jobStatus)
				if strings.Contains(jsLower, "closed") || strings.Contains(jsLower, "expired") {
					job.IsExpired = true
				}
			}
		}
	})

	// Job Description
	c.OnHTML("#jobDescriptionText", func(e *colly.HTMLElement) {
		if job.JobDescription == "" {
			job.JobDescription = strings.TrimSpace(e.Text)
		}
	})

	err = c.Visit(scrapeURL)
	if err != nil {
		return job, fmt.Errorf("error visiting URL: %v", err)
	}

	// Post-processing cleanup
	job.CompanyName = cleanScrapedText(job.CompanyName, false)
	job.JobDescription = cleanScrapedText(job.JobDescription, true)

	// If no title and no description were found, likely expired
	if job.JobTitle == "" && job.JobDescription == "" {
		job.IsExpired = true
	}

	return job, nil
}

// ScrapeSearch extracts a list of jobs from a search query on Indeed, navigating a batch of pages.
func (s *IndeedScraper) ScrapeSearch(ctx context.Context, query string, startOffset int) (SearchResponse, error) {
	var results []SearchResult
	seenIDs := make(map[string]bool)
	pagesPerBatch := 10 // Scrape 10 pages per batch
	limit := 10         // Indeed usually shows 10-15 results per page

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"),
	)
	c.SetCookieJar(s.jar)
	c.SetRequestTimeout(30 * time.Second)

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
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("Referer", "https://ar.indeed.com/")
	})

	c.OnHTML("div.job_seen_beacon", func(e *colly.HTMLElement) {
		var res SearchResult

		// Find Job ID from the link
		link := e.DOM.Find("a.jcs-JobTitle")

		jk, hasJk := link.Attr("data-jk")
		if hasJk && jk != "" {
			res.JobID = jk
		} else {
			href, hasHref := link.Attr("href")
			if hasHref {
				u, _ := url.Parse(href)
				res.JobID = u.Query().Get("jk")
			}
		}

		res.Title = strings.TrimSpace(link.Text())
		res.Company = strings.TrimSpace(e.DOM.Find("[data-testid='company-name']").Text())
		res.Location = strings.TrimSpace(e.DOM.Find("[data-testid='text-location']").Text())

		// Create a composite key to detect sponsored vs organic duplicates which often have distinct IDs
		dupeKey := fmt.Sprintf("%s|%s", strings.ToLower(res.Title), strings.ToLower(res.Company))

		// Skip if we couldn't find an ID or if we have already seen this job ID or title+company combo
		if res.JobID == "" || seenIDs[res.JobID] || seenIDs[dupeKey] {
			return
		}

		if res.Title != "" {
			seenIDs[res.JobID] = true
			seenIDs[dupeKey] = true
			results = append(results, res)
		}
	})

	lastPageFoundResults := false
	isBlocked := false

	c.OnHTML("title", func(e *colly.HTMLElement) {
		title := strings.ToLower(e.Text)
		if strings.Contains(title, "inicia sesión") || strings.Contains(title, "login") || strings.Contains(title, "crea una cuenta") {
			isBlocked = true
		}
	})

	for page := 0; page < pagesPerBatch; page++ {
		// If we already detected a block, stop
		if isBlocked {
			break
		}

		// Emit progress to the frontend
		runtime.EventsEmit(ctx, "scraping-progress", map[string]interface{}{
			"current": page + 1,
			"total":   pagesPerBatch,
			"found":   len(results),
		})

		currentStart := startOffset + (page * limit)
		searchURL := fmt.Sprintf("https://ar.indeed.com/jobs?q=%s&start=%d", url.QueryEscape(query), currentStart)

		countBefore := len(results)
		err := c.Visit(searchURL)
		if err != nil {
			break
		}

		if isBlocked {
			// Redirected to login page
			lastPageFoundResults = false
			break
		}

		if len(results) > countBefore {
			lastPageFoundResults = true
		} else {
			lastPageFoundResults = false
			// If we didn't find any new results on this page, it's likely the end of results
			if page > 0 { // Allow the first page of the batch to be empty just in case
				break
			}
		}

		// Small delay to be respectful and avoid blocks
		time.Sleep(600 * time.Millisecond)
	}

	return SearchResponse{
		Results:          results,
		HasMore:          lastPageFoundResults && !isBlocked,
		NextOffset:       startOffset + (pagesPerBatch * limit),
		IsBlockedByLogin: isBlocked,
	}, nil
}
