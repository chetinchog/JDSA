# SPECS: Job Description Scraper Argentina (JDSA)

## Overview
JDSA is a desktop application designed to scrape job descriptions from job boards (initially Indeed Argentina) and present the information in a clean, structured format.

## Technology Stack
- **Backend:** Go (Golang)
- **Scraping Engine:** Colly
- **Frontend Framework:** Vue.js 3
- **Styling:** Tailwind CSS
- **App Bridge:** Wails (Desktop Application Framework)
- **Target OS:** Windows (Executable: `JDSA.exe`)

## Functional Requirements
1. **Scrape Job Data:**
   - Input: URL (e.g., `https://ar.indeed.com/viewjob?jk=...`)
   - Output: JSON object with specific fields.
2. **Display Information:**
   - Material Design inspired UI.
   - Formatted view of the scraped data.
3. **Export Data:**
   - Button to download the scraped information as a `.json` file.

## Data Schema (JSON)
The scraper will target the following fields:
- `job_id`: Unique identifier from the URL.
- `company_name`: Name of the hiring company.
- `job_title`: Title of the position.
- `job_description`: Full text of the job description.
- `location`: Location of the job.
- `apply_url`: Original URL of the job post.
- `linkedin_company_id`: (Optional) ID if found.
- `job_poster_email`: (Optional) Email if found.

## UI/UX Design
- **Language:** Argentinian Spanish ("Descripción de Empleo", "Scrapear", "Exportar").
- **Theme:** Modern Material Design.
- **Layout:**
  - Header with title.
  - Centered input field for URL.
  - Action buttons (Scrap, Export).
  - Main area for results (cards or detailed list).

## Planned Workflow
1. User pastes an Indeed AR URL.
2. User clicks "Scrap".
3. Backend uses Colly to fetch the page and extract data.
4. Frontend displays the extracted data.
5. User clicks "Export" to save the JSON file.
