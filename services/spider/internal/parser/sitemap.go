package parser

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Hassan-ach/boogle/services/spider/internal/utils"
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

func parseSitemap(file []byte) (*SiteMaps, error) {
	fmt.Println("Starting sitemap.xml parsing")

	var sitemap SiteMaps
	// Unmarshal XML into sitemapXml struct
	err := xml.Unmarshal(file, &sitemap)
	if err != nil {
		return nil, err
	}

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
		x, ok := utils.NormalizeUrl(u.Loc, host.Host)
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
