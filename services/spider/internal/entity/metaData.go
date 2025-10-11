package entity

import (
	"fmt"
	"time"
)

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

// String returns a human-readable summary of metadata.
func (m *MetaData) String() string {
	return fmt.Sprintf(
		"MetaData\n"+
			"  URL: %s\n"+
			"  Title: %s\n"+
			"  Description: %s\n"+
			"  Type: %s\n"+
			"  SiteName: %s\n"+
			"  Locale: %s\n"+
			"  Keywords: %v\n"+
			"  Icons: %v\n"+
			"  CrawledAt: %s\n",
		m.URL,
		m.Title,
		m.Description,
		m.Type,
		m.SiteName,
		m.Locale,
		m.Keywords,
		m.Icons,
		m.CrawledAt.Format(time.RFC3339),
	)
}
