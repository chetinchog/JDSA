# SPECS: Job Description Scraping Assistant (JDSA)

## Overview
JDSA is a desktop application designed to scrape job descriptions from various job boards and present the information in a clean, structured format.

## Technology Stack
- **Backend:** Go (Golang)
- **Scraping Engine:** Colly
- **Frontend Framework:** Vue.js 3
- **Styling:** Tailwind CSS
- **App Bridge:** Wails (Desktop Application Framework)
- **Target OS:** Windows (Executable: `JDSA.exe`)

## Functional Requirements
1. **Scrape Job Data:**
   - Input: Job Post URL.
   - Output: JSON object with specific fields.
2. **Display Information:**
   - Material Design inspired UI.
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

## UI/UX Design
- **Theme:** Modern Material Design.
- **Layout:**
  - Header with title.
  - Centered input field for URL.
  - Action buttons (Scrap, Export).
  - Main area for results with internal scrolling for descriptions.

## Planned Workflow
1. User pastes a job post URL.
2. User clicks "Scrap".
3. Backend uses scrapers to fetch and extract data.
4. Frontend displays the extracted data.
5. User clicks "Export" to save the JSON file.
