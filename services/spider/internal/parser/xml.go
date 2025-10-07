package parser

import (
	"encoding/xml"

	"spider/internal/utils"
)

type url struct {
	Loc string `xml:"loc"`
}

type sitemapXml struct {
	Urls []url `xml:"url"`
}

func SitMap(file []byte) ([]string, error) {
	log := utils.Log.Parsing().With("operation", "SitMap")
	log.Info("Starting sitemap.xml parsing")

	var sitemap sitemapXml
	err := xml.Unmarshal(file, &sitemap)
	if err != nil {
		log.Error("Failed to unmarshal sitemap XML", "error", err)
		return nil, err
	}

	locs := make([]string, len(sitemap.Urls))
	for i, u := range sitemap.Urls {
		locs[i] = u.Loc
	}

	log.Info("Successfully parsed sitemap.xml", "urlsFound", len(locs))
	return locs, nil
}
