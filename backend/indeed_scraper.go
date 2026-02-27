package backend

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// IndeedTransport acts as a middleware (decorator pattern) for all outbound HTTP requests.
// It handles the low-level injection of cookies and headers required to bypass Indeed's anti-bot measures uniformly.
type IndeedTransport struct {
	Base          http.RoundTripper
	SessionCookie string
}

func (t *IndeedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid mutating the original request, which is bad practice in RoundTrip.
	modReq := req.Clone(req.Context())

	// Apply headers uniformly to every request
	modReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	modReq.Header.Set("Accept-Language", "es-AR,es;q=0.9,en;q=0.8")
	modReq.Header.Set("Sec-Ch-Ua", `"Not A(Bit:Major";v="99", "Google Chrome";v="121", "Chromium";v="121"`)
	modReq.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	modReq.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	modReq.Header.Set("Sec-Ch-Ua-Platform-Version", `"10.0.0"`)
	modReq.Header.Set("Sec-Fetch-Dest", "document")
	modReq.Header.Set("Sec-Fetch-Mode", "navigate")
	modReq.Header.Set("Sec-Fetch-Site", "same-origin")
	modReq.Header.Set("Sec-Fetch-User", "?1")
	modReq.Header.Set("Upgrade-Insecure-Requests", "1")

	// Set a default Referer if not already present
	if modReq.Header.Get("Referer") == "" {
		modReq.Header.Set("Referer", "https://ar.indeed.com/")
	}

	// If we have manual cookies, we merge them with any cookies colly/jar already set.
	if t.SessionCookie != "" {
		existing := modReq.Header.Get("Cookie")
		finalCookie := t.SessionCookie
		if existing != "" {
			// JAR cookies (existing) must overwrite manual ones because they contain
			// the latest security rotations from the response of the previous page.
			finalCookie = mergeCookies(t.SessionCookie, existing)
		}

		// Force the raw string into the map to bypass net/http's AddCookie sanitization.
		// Go's Transport will write this string exactly as-is to the wire.
		modReq.Header["Cookie"] = []string{finalCookie}
	}

	return t.Base.RoundTrip(modReq)
}

// mergeCookies attempts a naive merge of two cookie strings.
func mergeCookies(c1, c2 string) string {
	m := make(map[string]string)

	parse := func(s string) {
		parts := strings.Split(s, ";")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			kv := strings.SplitN(p, "=", 2)
			if len(kv) == 2 {
				m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}

	parse(c1)
	parse(c2) // c2 (manual session) wins on conflicts

	var res []string
	for k, v := range m {
		res = append(res, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(res, "; ")
}

// IndeedScraper handles scraping job data from Indeed job pages.
type IndeedScraper struct {
	jar       http.CookieJar
	transport *IndeedTransport
}

// NewIndeedScraper creates a new IndeedScraper with a real jar for rotated tokens.
func NewIndeedScraper() *IndeedScraper {
	jar, _ := cookiejar.New(nil)
	return &IndeedScraper{
		jar: jar,
		transport: &IndeedTransport{
			Base: http.DefaultTransport,
		},
	}
}

// CanHandle returns true if the host contains "indeed".
func (s *IndeedScraper) CanHandle(host string) bool {
	return strings.Contains(strings.ToLower(host), "indeed")
}

func getModernUA() string {
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"
}

// newCollector creates a pre-configured colly collector with all required
// headers, cookie jar, transport, and timeout. This is the SINGLE source
// of truth for request configuration — every outbound request goes through here.
func (s *IndeedScraper) newCollector() *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent(getModernUA()),
	)

	c.SetCookieJar(s.jar)
	c.WithTransport(s.transport)
	c.SetRequestTimeout(30 * time.Second)

	// Note: Headers are now injected robustly at the Transport layer (IndeedTransport)
	// so we don't need to manually configure them in colly's OnRequest. They will
	// be applied to *every* request automatically, resolving the 403 Forbidden issues.

	return c
}

// Scrape extracts job data from an Indeed URL.
func (s *IndeedScraper) Scrape(targetURL string) (JobData, error) {
	var job JobData
	job.ApplyURL = targetURL

	// Attempt to parse JobID from the URL query params
	if u, err := url.Parse(targetURL); err == nil {
		job.JobID = u.Query().Get("jk")
	}

	c := s.newCollector()

	c.OnHTML("h1.jobsearch-JobInfoHeader-title", func(e *colly.HTMLElement) {
		job.JobTitle = strings.TrimSpace(e.Text)
	})
	c.OnHTML("[data-testid='jobsearch-JobInfoHeader-title']", func(e *colly.HTMLElement) {
		if job.JobTitle == "" {
			job.JobTitle = strings.TrimSpace(e.Text)
		}
	})
	c.OnHTML("[data-testid='inlineHeader-companyName']", func(e *colly.HTMLElement) {
		job.CompanyName = strings.TrimSpace(e.Text)
	})
	c.OnHTML("[data-testid='inlineHeader-companyLocation']", func(e *colly.HTMLElement) {
		job.Location = strings.TrimSpace(e.Text)
	})
	c.OnHTML("[data-testid='job-location']", func(e *colly.HTMLElement) {
		if job.Location == "" {
			job.Location = strings.TrimSpace(e.Text)
		}
	})
	c.OnHTML("div#jobLocationText", func(e *colly.HTMLElement) {
		if job.Location == "" {
			job.Location = strings.TrimSpace(e.Text)
		}
	})
	c.OnHTML("#jobDescriptionText", func(e *colly.HTMLElement) {
		job.JobDescription = strings.TrimSpace(e.Text)
	})

	err := c.Visit(targetURL)
	if err != nil {
		return job, fmt.Errorf("error visiting URL: %v", err)
	}

	job.CompanyName = cleanScrapedText(job.CompanyName, false)
	job.JobDescription = cleanScrapedText(job.JobDescription, true)

	if job.JobTitle == "" && job.JobDescription == "" {
		job.IsExpired = true
	}

	return job, nil
}

// ScrapeSearch extracts a list of jobs from a search query on Indeed.
func (s *IndeedScraper) ScrapeSearch(ctx context.Context, query string, startOffset int) (SearchResponse, error) {
	var results []SearchResult
	seenIDs := make(map[string]bool)
	limit := 10

	c := s.newCollector()

	isBlocked := false
	currentReferer := "https://ar.indeed.com/"
	page := 0

	// Override Referer dynamically for pagination (adds on top of base OnRequest)
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", currentReferer)
	})

	c.OnHTML("title", func(e *colly.HTMLElement) {
		title := strings.ToLower(e.Text)
		if (strings.Contains(title, "iniciar sesión") && strings.Contains(title, "cuentas")) ||
			strings.Contains(title, "captcha") || strings.Contains(title, "challenge") {
			isBlocked = true
		}
	})

	c.OnResponse(func(r *colly.Response) {
		urlStr := r.Request.URL.String()
		if strings.Contains(urlStr, "/auth") || strings.Contains(urlStr, "captcha") {
			isBlocked = true
		}
	})

	c.OnHTML("div.job_seen_beacon", func(e *colly.HTMLElement) {
		var res SearchResult
		link := e.DOM.Find("a.jcs-JobTitle")
		jk, _ := link.Attr("data-jk")
		if jk == "" {
			href, _ := link.Attr("href")
			u, _ := url.Parse(href)
			jk = u.Query().Get("jk")
		}

		if jk != "" {
			res.JobID = jk
			res.Title = strings.TrimSpace(link.Text())
			res.Company = strings.TrimSpace(e.DOM.Find("[data-testid='company-name']").Text())
			res.Location = strings.TrimSpace(e.DOM.Find("[data-testid='text-location']").Text())

			dupeKey := fmt.Sprintf("%s|%s", strings.ToLower(res.Title), strings.ToLower(res.Company))
			if !seenIDs[jk] && !seenIDs[dupeKey] {
				seenIDs[jk] = true
				seenIDs[dupeKey] = true
				results = append(results, res)

				// Emit progress per job found (cumulative)
				runtime.EventsEmit(ctx, "scraping-progress", map[string]interface{}{
					"found": len(results),
					"page":  page + 1,
				})
			}
		}
	})

	cancelled := false
	for {
		// Check for cancellation
		select {
		case <-ctx.Done():
			cancelled = true
		default:
		}
		if cancelled || isBlocked {
			break
		}

		currentStart := startOffset + (page * limit)
		searchURL := fmt.Sprintf("https://ar.indeed.com/jobs?q=%s&start=%d", url.QueryEscape(query), currentStart)

		countBefore := len(results)
		err := c.Visit(searchURL)
		if err != nil {
			break
		}

		if isBlocked {
			break
		}

		// If no new results found on this page, we've exhausted results
		if len(results) <= countBefore {
			if page > 0 {
				break
			}
		}

		currentReferer = searchURL
		page++

		// Randomized delay to mimic human behavior (1s to 2.5s)
		time.Sleep(time.Duration(1000+rand.Intn(1500)) * time.Millisecond)
	}

	return SearchResponse{
		Results:          results,
		HasMore:          false, // We now scan all pages in one go
		NextOffset:       startOffset + (page * limit),
		IsBlockedByLogin: isBlocked,
	}, nil
}

// SetSessionCookie sets the manual session cookie to bypass login walls.
func (s *IndeedScraper) SetSessionCookie(cookie string) {
	s.transport.SessionCookie = strings.TrimSpace(cookie)
}

// HasSessionCookie returns true if a session cookie is set.
func (s *IndeedScraper) HasSessionCookie() bool {
	return s.transport.SessionCookie != ""
}

// GetSessionCookie returns the current session cookie.
func (s *IndeedScraper) GetSessionCookie() string {
	return s.transport.SessionCookie
}
