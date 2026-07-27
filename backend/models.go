package backend

type JobData struct {
	IsExpired         bool   `json:"is_expired"`
	JobID             string `json:"job_id"`
	CompanyName       string `json:"company_name"`
	JobTitle          string `json:"job_title"`
	JobDescription    string `json:"job_description"`
	Location          string `json:"location"`
	ApplyURL          string `json:"apply_url"`
	LinkedinCompanyID string `json:"linkedin_company_id,omitempty"`
	JobPosterEmail    string `json:"job_poster_email,omitempty"`
}

type SearchResult struct {
	JobID    string `json:"job_id"`
	Title    string `json:"title"`
	Company  string `json:"company"`
	Location string `json:"location"`
}

type ExportResult struct {
	SuccessCount int      `json:"successCount"`
	ErrorCount   int      `json:"errorCount"`
	Errors       []string `json:"errors"`
	FilePath     string   `json:"filePath"`
}

type SearchResponse struct {
	Results          []SearchResult `json:"results"`
	HasMore          bool           `json:"hasMore"`
	NextOffset       int            `json:"nextOffset"`
	IsBlockedByLogin bool           `json:"isBlockedByLogin"`
	BlockedReason    string         `json:"blockedReason,omitempty"`
	FailedURL        string         `json:"failedURL,omitempty"`
	HasDebugHTML     bool           `json:"hasDebugHTML,omitempty"`
}

type ScraperConfig struct {
	SessionCookie string `json:"session_cookie"`
	WaitPagesMin  int    `json:"wait_pages_min"`
	WaitPagesMax  int    `json:"wait_pages_max"`
	WaitJobMin    int    `json:"wait_job_min"`
	WaitJobMax    int    `json:"wait_job_max"`
}
