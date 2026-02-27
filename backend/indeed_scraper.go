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

// cookieBypassTransport handles the low-level injection of cookies.
// It merges cookies from the Jar (legal) with the manual session cookie (potentially illegal bits).
type cookieBypassTransport struct {
	Base          http.RoundTripper
	SessionCookie string
}

func (t *cookieBypassTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// If we have manual cookies, we merge them with any cookies colly/jar already set.
	if t.SessionCookie != "" {
		existing := req.Header.Get("Cookie")
		finalCookie := t.SessionCookie
		if existing != "" {
			// JAR cookies (existing) must overwrite manual ones because they contain
			// the latest security rotations from the response of the previous page.
			finalCookie = mergeCookies(t.SessionCookie, existing)
		}

		// Force the raw string into the map to bypass net/http's AddCookie sanitization.
		// Go's Transport will write this string exactly as-is to the wire.
		req.Header["Cookie"] = []string{finalCookie}
	}

	return t.Base.RoundTrip(req)
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
	transport *cookieBypassTransport
}

// NewIndeedScraper creates a new IndeedScraper with a real jar for rotated tokens.
func NewIndeedScraper() *IndeedScraper {
	jar, _ := cookiejar.New(nil)
	return &IndeedScraper{
		jar: jar,
		transport: &cookieBypassTransport{
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

// Scrape extracts job data from an Indeed URL.
func (s *IndeedScraper) Scrape(targetURL string) (JobData, error) {
	var job JobData
	job.ApplyURL = targetURL

	c := colly.NewCollector(
		colly.UserAgent(getModernUA()),
	)

	c.SetCookieJar(s.jar)
	c.WithTransport(s.transport)
	c.SetRequestTimeout(30 * time.Second)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "es-AR,es;q=0.9,en;q=0.8")
		r.Headers.Set("Sec-Ch-Ua", `"Not A(Bit:Major";v="99", "Google Chrome";v="121", "Chromium";v="121"`)
		r.Headers.Set("Sec-Ch-Ua-Mobile", "?0")
		r.Headers.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		r.Headers.Set("Sec-Ch-Ua-Platform-Version", `"10.0.0"`)
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "same-origin")
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("Referer", "https://ar.indeed.com/")
	})

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
	pagesPerBatch := 10
	limit := 10

	c := colly.NewCollector(
		colly.UserAgent(getModernUA()),
	)

	c.SetCookieJar(s.jar)
	c.WithTransport(s.transport)
	c.SetRequestTimeout(30 * time.Second)

	isBlocked := false
	currentReferer := "https://ar.indeed.com/"

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "es-AR,es;q=0.9,en;q=0.8")
		r.Headers.Set("Sec-Ch-Ua", `"Not A(Bit:Major";v="99", "Google Chrome";v="121", "Chromium";v="121"`)
		r.Headers.Set("Sec-Ch-Ua-Mobile", "?0")
		r.Headers.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		r.Headers.Set("Sec-Ch-Ua-Platform-Version", `"10.0.0"`)
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "same-origin")
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
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
			}
		}
	})

	lastPageFoundResults := false
	for page := 0; page < pagesPerBatch; page++ {
		if isBlocked {
			break
		}

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
			lastPageFoundResults = false
			break
		}

		if len(results) > countBefore {
			lastPageFoundResults = true
		} else {
			lastPageFoundResults = false
			if page > 0 {
				break
			}
		}

		// Randomized delay to mimic human behavior (1s to 2.5s)
		time.Sleep(time.Duration(1000+rand.Intn(1500)) * time.Millisecond)
	}

	return SearchResponse{
		Results:          results,
		HasMore:          lastPageFoundResults && !isBlocked,
		NextOffset:       startOffset + (pagesPerBatch * limit),
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
