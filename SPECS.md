# SPECS: Job Description Scraping Assistant (JDSA)

## Overview
JDSA is a desktop application designed to scrape job descriptions from various job boards and present the information in a clean, structured format.

## Technology Stack
- **Backend:** Go 1.26
- **Scraping Engine:** Colly (with 30s request timeout)
- **Frontend Framework:** Vue.js 3
- **Styling:** Tailwind CSS v4
- **App Bridge:** Wails v2 (Desktop Application Framework)
- **Target OS:** Windows (Executable: `JDSA.exe`)

## Functional Requirements
1. **Scrape Job Data:**
   - Input: Job Post URL (validated — HTTP/HTTPS only, no private IPs).
   - Output: JSON object with specific fields.
2. **Display Information:**
   - Material Design inspired UI with light/dark mode toggle.
   - Formatted view of the scraped data.
3. **Export Data:**
   - Button to download the scraped information as a `.json` file.

## Data Schema (JSON)
The scraper targets the following fields:
- `job_id`: Unique identifier from the source URL.
- `company_name`: Name of the hiring company.
- `job_title`: Title of the position.
- `job_description`: Full text of the job description (preserving formatting).
- `location`: Location of the job.
- `apply_url`: Original URL of the job post.
- `linkedin_company_id` *(optional)*: LinkedIn company identifier.
- `job_poster_email` *(optional)*: Email of the job poster.

## UI/UX Design
- **Theme:** Modern Material Design with light/dark mode (toggle in header, persists via localStorage).
- **Layout:**
  - Header with title and theme toggle button.
  - Centered input field for URL.
  - Action buttons (Scrap, Export).
  - Main area for results with internal scrolling for descriptions.
  - Footer credits: "By iCTG - Powered by Antigravity".

## Planned Workflow
1. User pastes a job post URL.
2. User clicks "Scrap".
3. Backend validates the URL and uses scrapers to fetch and extract data.
4. Frontend displays the extracted data.
5. User clicks "Export" to save the JSON file.
