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
}
