package entity

import (
	"time"
)

type Host struct {
	MaxRetry        int      // Maximum retries per URL
	MaxPages        int      // Maximum pages to crawl for this host
	PagesCrawled    int      // Pages already crawled
	Delay           int      // Delay between requests in seconds
	Name            string   // Hostname
	AllowedUrls     []string // URL patterns allowed to crawl
	NotAllowedPaths []string // Paths disallowed to crawl (typo kept for backward compatibility)
}
type MetaData struct {
	URL         string    `json:"url"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	Type        string    `json:"type,omitempty"`
	SiteName    string    `json:"siteName,omitempty"`
	Locale      string    `json:"locale,omitempty"`
	Keywords    []string  `json:"keywords,omitempty"`
	Icons       []string  `json:"icons,omitempty"`
	CrawledAt   time.Time `json:"crawledAt"`
}

type Robots struct {
	Allow      []string
	Disallow   []string
	SiteMaps   []string
	CrawlDelay int
}

type Page struct {
	MetaData          // embeds MetaData
	StatusCode int    // HTTP response code
	HTML       []byte // Raw HTML content
	Images     []string
	Links      []string
}
