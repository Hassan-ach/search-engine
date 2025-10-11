package parser

import (
	"encoding/xml"

	"spider/internal/utils"
)

type u struct {
	Loc string `xml:"loc"`
}

func (u u) Parse(raw string) (any, error) {
	panic("unimplemented")
}

type sitemapXml struct {
	Urls []u `xml:"url"`
}

// sitMap parses a sitemap XML file and extracts all URLs listed in <loc> tags.
// Returns a slice of normalized URL strings and any parsing error encountered.
// Logs start, end, number of URLs found, and errors.
func siteMap(file []byte) ([]string, error) {
	log := utils.Log.Parsing().With("operation", "SitMap")
	log.Info("Starting sitemap.xml parsing")

	var sitemap sitemapXml
	// Unmarshal XML into sitemapXml struct
	err := xml.Unmarshal(file, &sitemap)
	if err != nil {
		log.Error("Failed to unmarshal sitemap XML", "error", err)
		return nil, err
	}

	// Extract URLs from <loc> tags
	locs := make([]string, len(sitemap.Urls))
	for i, u := range sitemap.Urls {
		locs[i] = u.Loc
	}

	log.Info("Successfully parsed sitemap.xml", "urlsFound", len(locs))
	return locs, nil
}
