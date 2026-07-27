package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"

	"jdsa/backend"
)

func main() {
	// Reemplazar con una cookie de prueba de la sesión si es necesario
	sampleCookie := "" 

	fmt.Println("--- Test 1: Peticion a /viewjob sin cookie ---")
	{
		req, _ := http.NewRequest("GET", "https://ar.indeed.com/viewjob?jk=1234567890abcdef", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
		req.Header.Set("Accept-Language", "es-AR,es;q=0.9,en-US;q=0.8,en;q=0.7")
		req.Header.Set("Referer", "https://ar.indeed.com/jobs?q=developer")
		req.Header.Set("Sec-Ch-Ua", `"Not(A:Brand";v="99", "Microsoft Edge";v="133", "Chromium";v="133"`)
		req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		req.Header.Set("Sec-Fetch-Dest", "document")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		req.Header.Set("Sec-Fetch-User", "?1")

		jar, _ := cookiejar.New(nil)
		client := &http.Client{Jar: jar}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("viewjob status: %d | BodyLen: %d\n", resp.StatusCode, len(body))
		}
	}

	fmt.Println("\n--- Test 2: IndeedScraper con SetConfig (cookie) ---")
	{
		scraper := backend.NewIndeedScraper()
		if sampleCookie != "" {
			scraper.SetConfig(backend.ScraperConfig{
				SessionCookie: sampleCookie,
			})
		}
		res, _ := scraper.ScrapeSearch(context.Background(), "Python", 0)
		fmt.Printf("ScrapeSearch resultados: %d | Blocked: %v | Reason: %s\n", len(res.Results), res.IsBlockedByLogin, res.BlockedReason)
	}
}
