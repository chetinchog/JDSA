package backend

import (
	"testing"
)

func TestIsValidJobURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		// Valid URLs
		{"valid http", "http://www.indeed.com/viewjob?jk=abc123", false},
		{"valid https", "https://ar.indeed.com/viewjob?jk=abc123", false},
		{"valid linkedin", "https://www.linkedin.com/jobs/view/12345", false},

		// Invalid scheme
		{"file scheme", "file:///etc/passwd", true},
		{"ftp scheme", "ftp://example.com/file", true},
		{"no scheme", "www.indeed.com/viewjob", true},
		{"javascript scheme", "javascript:alert(1)", true},

		// SSRF: localhost and private IPs
		{"localhost", "http://localhost/admin", true},
		{"127.0.0.1", "http://127.0.0.1:8080/secret", true},
		{"0.0.0.0", "http://0.0.0.0/", true},
		{"private 192.168", "http://192.168.1.1/admin", true},
		{"private 10.x", "http://10.0.0.1/internal", true},
		{"private 172.16", "http://172.16.0.1/api", true},
		{"link-local", "http://169.254.169.254/latest/meta-data/", true},

		// Edge cases
		{"empty string", "", true},
		{"no host", "https://", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := isValidJobURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("isValidJobURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestCleanScrapedText(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		preserveNewlines bool
		want             string
	}{
		{
			name:             "removes CSS blocks",
			input:            "Hello {color: red; font-size: 14px} World",
			preserveNewlines: false,
			want:             "Hello World",
		},
		{
			name:             "removes CSS class names",
			input:            "css-abc123 Hello .css-xyz789 World",
			preserveNewlines: false,
			want:             "Hello World",
		},
		{
			name:             "removes HTML tags",
			input:            "<p>Hello <strong>World</strong></p>",
			preserveNewlines: false,
			want:             "Hello World",
		},
		{
			name:             "decodes HTML entities",
			input:            "A &amp; B &lt; C &gt; D",
			preserveNewlines: false,
			want:             "A & B < C > D",
		},
		{
			name:             "replaces nbsp",
			input:            "Hello&nbsp;World",
			preserveNewlines: false,
			want:             "Hello World",
		},
		{
			name:             "collapses whitespace without preserving newlines",
			input:            "Hello    World   \n  Foo",
			preserveNewlines: false,
			want:             "Hello World Foo",
		},
		{
			name:             "preserves newlines when requested",
			input:            "Hello    World\nFoo    Bar",
			preserveNewlines: true,
			want:             "Hello World\nFoo Bar",
		},
		{
			name:             "trims leading and trailing whitespace",
			input:            "   Hello World   ",
			preserveNewlines: false,
			want:             "Hello World",
		},
		{
			name:             "handles combined mess",
			input:            "  .css-abc {font: bold} <b>Job</b> &amp; Title  ",
			preserveNewlines: false,
			want:             "Job & Title",
		},
		{
			name:             "empty string",
			input:            "",
			preserveNewlines: false,
			want:             "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanScrapedText(tt.input, tt.preserveNewlines)
			if got != tt.want {
				t.Errorf("cleanScrapedText(%q, %v) = %q, want %q", tt.input, tt.preserveNewlines, got, tt.want)
			}
		})
	}
}

// --- Strategy Pattern Tests ---

func TestScraperRegistry_GetScraper(t *testing.T) {
	registry := NewScraperRegistry(NewIndeedScraper())

	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{"indeed ar", "ar.indeed.com", false},
		{"indeed www", "www.indeed.com", false},
		{"indeed plain", "indeed.com", false},
		{"linkedin", "www.linkedin.com", true},
		{"unknown", "example.com", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scraper, err := registry.GetScraper(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetScraper(%q) error = %v, wantErr %v", tt.host, err, tt.wantErr)
			}
			if !tt.wantErr && scraper == nil {
				t.Errorf("GetScraper(%q) returned nil scraper, expected non-nil", tt.host)
			}
		})
	}
}

func TestIndeedScraper_CanHandle(t *testing.T) {
	s := NewIndeedScraper()

	tests := []struct {
		host string
		want bool
	}{
		{"ar.indeed.com", true},
		{"www.indeed.com", true},
		{"indeed.com", true},
		{"www.linkedin.com", false},
		{"example.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := s.CanHandle(tt.host)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestParseLocation(t *testing.T) {
	tests := []struct {
		name     string
		location string
		want     string
	}{
		{"general remote", "Remoto", "AR"},
		{"home office", "Home Office", "AR"},
		{"desde casa", "Trabajo desde casa", "AR"},
		{"unknown location", "Ubicación Desconocida 123", "AR"},
		{"empty location", "", "AR"},
		{"remote US", "Remote-US", "US"},
		{"explicit USA", "California, USA", "US"},
		{"explicit Spain", "Madrid, España", "ES"},
		{"explicit Argentina", "Buenos Aires, Argentina", "AR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLocation(tt.location)
			if got != tt.want {
				t.Errorf("parseLocation(%q) = %q, want %q", tt.location, got, tt.want)
			}
		})
	}
}

