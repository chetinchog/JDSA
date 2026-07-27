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
	// SetConfig sets the scraper configuration including delays and cookies.
	SetConfig(config ScraperConfig)
	// HasValidConfig returns true if a session cookie or valid config is currently set.
	HasValidConfig() bool
	// GetConfig returns the current scraper configuration.
	GetConfig() ScraperConfig
	// GetLastDebugHTML returns the last response HTML content and its URL for debugging.
	GetLastDebugHTML() ([]byte, string)
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
