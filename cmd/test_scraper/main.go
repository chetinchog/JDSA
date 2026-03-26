package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"
	"github.com/gocolly/colly/v2"
	"jdsa/backend"
)

func main() {
	fmt.Println("Starting Colly test for 'Rust developer'...")
	
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"),
	)
	jar, _ := cookiejar.New(nil)
	c.SetCookieJar(jar)
	c.WithTransport(&backend.IndeedTransport{Base: http.DefaultTransport})
	c.SetRequestTimeout(30 * time.Second)

	c.OnResponse(func(r *colly.Response) {
		fmt.Printf("Status: %d\n", r.StatusCode)
		body := string(r.Body)
		fmt.Printf("Body length: %d\n", len(body))
		if len(body) > 500 {
			fmt.Println("Body preview:", body[:500])
		} else {
			fmt.Println("Body:", body)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Error: %v (Status: %d)\n", err, r.StatusCode)
	})

	c.Visit("https://ar.indeed.com/jobs?q=Rust+developer")
}
