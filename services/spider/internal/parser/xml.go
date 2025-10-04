package parser

import (
	"encoding/xml"
	"fmt"
)

type url struct {
	Loc string `xml:"loc"`
}

type sitemapXml struct {
	Urls []url `xml:"url"`
}

func SitMap(file []byte) ([]string, error) {
	var sitemap sitemapXml
	err := xml.Unmarshal(file, &sitemap)
	if err != nil {
		fmt.Printf("error while parsing sitemap.xml file: %v\n", err)
		return nil, err
	}

	locs := make([]string, len(sitemap.Urls))
	for i, u := range sitemap.Urls {
		locs[i] = u.Loc
	}
	return locs, nil
}
