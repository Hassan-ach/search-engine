package entity

import (
	"fmt"
	"time"
)

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
		m.Url,
		m.Title,
		m.Description,
		m.Type,
		m.SiteName,
		m.Local,
		m.Keywords,
		m.Icons,
		m.CrawledAt.Format(time.RFC3339),
	)
}
