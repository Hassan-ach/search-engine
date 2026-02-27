package model

import (
	"time"

	"github.com/google/uuid"
)

type Page struct {
	URLID       uuid.UUID      `json:"urlid"`
	URL         string         `json:"url"`
	PRScore     float64        `json:"pr_score"`
	Words       map[string]int `json:"words"`
	GlobalScore float64        `json:"global_score"`
	MetaData    MetaData       `json:"metadata"`
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
