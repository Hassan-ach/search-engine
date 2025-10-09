package entity

import "time"

type MetaData struct {
	Url         string
	Title       string
	Description string
	Type        string
	SiteName    string
	Local       string
	Keywords    []string
	Icons       []string
	CrawledAt   time.Time
}
