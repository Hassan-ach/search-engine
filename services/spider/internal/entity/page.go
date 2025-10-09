package entity

import (
	"fmt"
	"strings"
	"time"

	"spider/internal/utils"
)

type Page struct {
	MetaData
	StatusCode int
	HTML       []byte
	// Text       string
	Images *utils.Set[string]
	Links  *utils.Set[string]
}

func (p *Page) String() {
	fmt.Println("===== Page Info =====")
	fmt.Printf("URL: %s\n", p.Url)
	fmt.Printf("Status Code: %d\n", p.StatusCode)
	fmt.Printf("Crawled At: %s\n", p.CrawledAt.Format(time.RFC3339))
	fmt.Println()

	fmt.Println("----- Meta Data -----")
	fmt.Printf("Title: %s\n", p.Title)
	fmt.Printf("Description: %s\n", p.Description)
	fmt.Printf("Type: %s\n", p.Type)
	fmt.Printf("Site Name: %s\n", p.SiteName)
	fmt.Printf("Locale: %s\n", p.Local)
	if len(p.Keywords) > 0 {
		fmt.Printf("Keywords: %s\n", strings.Join(p.Keywords, ", "))
	}
	if len(p.Icons) > 0 {
		fmt.Printf("Icons: %s\n", strings.Join(p.Icons, ", "))
	}
	fmt.Println()

	fmt.Println("----- Links -----")
	for _, l := range p.Links.GetAll() {
		fmt.Printf("  - %s\n", l)
	}

	fmt.Println("----- Images -----")
	for _, img := range p.Images.GetAll() {
		fmt.Printf("  - %s\n", img)
	}

	fmt.Println("=====================")
}
