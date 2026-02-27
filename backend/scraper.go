package backend

import (
	"context"
	"fmt"
	"strings"
)

// Scraper defines the contract for site-specific job scrapers.
type Scraper interface {
	// CanHandle returns true if this scraper supports the given URL host.
	CanHandle(host string) bool
	// Scrape extracts job data from the given URL.
	Scrape(targetURL string) (JobData, error)
	// ScrapeSearch extracts a list of jobs from a search query.
	ScrapeSearch(ctx context.Context, query string, start int) (SearchResponse, error)
	// SetSessionCookie sets a custom cookie string to be used for requests.
	SetSessionCookie(cookie string)
	// HasSessionCookie returns true if a session cookie is currently set.
	HasSessionCookie() bool
	// GetSessionCookie returns the current session cookie.
	GetSessionCookie() string
}

// ScraperRegistry holds all registered scrapers and selects the right one.
type ScraperRegistry struct {
	scrapers []Scraper
}

// NewScraperRegistry creates a registry with the given scrapers.
func NewScraperRegistry(scrapers ...Scraper) *ScraperRegistry {
	return &ScraperRegistry{scrapers: scrapers}
}

// GetScraper returns the first scraper that can handle the given host.
func (r *ScraperRegistry) GetScraper(host string) (Scraper, error) {
	host = strings.ToLower(host)
	for _, s := range r.scrapers {
		if s.CanHandle(host) {
			return s, nil
		}
	}
	return nil, fmt.Errorf("no hay un scraper disponible para el host: %s", host)
}
