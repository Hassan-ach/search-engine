package parser

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"

	"spider/internal/utils"
)

type u struct {
	Loc string `xml:"loc"`
}

func (u u) Parse(raw string) (any, error) {
	panic("unimplemented")
}

type SiteMaps struct {
	Urls []u `xml:"url"`
}

// ParseSitemap parses a sitemap XML file and extracts all URLs listed in <loc> tags.
// Returns a slice of normalized URL strings and any parsing error encountered.
// Logs start, end, number of URLs found, and errors.
func parseSitemap(file []byte) (*SiteMaps, error) {
	fmt.Println("Starting sitemap.xml parsing")

	var sitemap SiteMaps
	// Unmarshal XML into sitemapXml struct
	err := xml.Unmarshal(file, &sitemap)
	if err != nil {
		return nil, err
	}

	// Extract URLs from <loc> tags
	// locs := make([]string, len(sitemap.Urls))
	// for i, u := range sitemap.Urls {
	// 	locs[i] = u.Loc
	// }

	return &sitemap, nil
}

func fetchSitemap(client *http.Client, sitemapURL string, host *url.URL) ([]string, error) {
	var r []string

	siteUrl, _ := url.Parse(sitemapURL)
	if siteUrl.Scheme == "" {
		siteUrl.Scheme = "https"
	}
	if siteUrl.Host == "" {
		siteUrl.Host = host.Host
	}

	file, _, err := utils.GetReq(client, siteUrl.String(), 1, 5)
	if err != nil {
		return nil, fmt.Errorf("fetching sitemap: %w", err)
	}

	d, err := parseSitemap(file)
	if err != nil {
		return nil, fmt.Errorf("parsing sitemap: %w", err)
	}

	for _, u := range d.Urls {
		x, ok := utils.NormalizeUrl(u.Loc, host)
		if !ok {
			return nil, fmt.Errorf("fetching sitemap: invalid URL %s", u.Loc)
		}
		r = append(r, x)
	}

	return r, nil
}

func FetchSitemaps(client *http.Client, s []string, host *url.URL) []string {
	var r []string
	for _, sitemapURL := range s {
		if siteUrls, err := fetchSitemap(client, sitemapURL, host); err == nil {
			r = append(r, siteUrls...)
		}
	}
	return r
}
