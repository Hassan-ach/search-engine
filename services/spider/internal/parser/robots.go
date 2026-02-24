package parser

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"spider/internal/utils"
)

type Robots struct {
	Allow      []string
	Disallow   []string
	SiteMaps   []string
	CrawlDelay int
}

func parseRobots(txt, ua string) *Robots {
	r := &Robots{
		CrawlDelay: 5,
	}

	if ua == "" {
		ua = "*"
	}

	var uaActive bool
	for _, line := range strings.Split(txt, "\n") {

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		lower := strings.ToLower(line)

		switch {
		case strings.HasPrefix(lower, "user-agent:"):
			userAgent := strings.TrimSpace(line[11:])
			uaActive = (userAgent == ua || userAgent == "*")
		case uaActive:
			switch {
			case strings.HasPrefix(lower, "disallow:"):
				rule := strings.TrimSpace(line[9:])
				r.Disallow = append(r.Disallow, rule)
			case strings.HasPrefix(lower, "allow:"):
				rule := strings.TrimSpace(line[6:])
				r.Allow = append(r.Allow, rule)
			case strings.HasPrefix(lower, "crawl-delay:"):
				if d, err := strconv.Atoi(strings.TrimSpace(line[12:])); err == nil && d > 0 {
					r.CrawlDelay = d
				}
			}

		case strings.HasPrefix(lower, "sitemap:"):
			sitemapURL := strings.TrimSpace(line[len("Sitemap:"):])
			r.SiteMaps = append(r.SiteMaps, sitemapURL)
		}
	}
	return r
}

func NewRobots(client *http.Client, u *url.URL) (*Robots, error) {
	h := strings.TrimPrefix(u.Host, "www.")

	u.Scheme = "https"
	u.Host = h
	u.RawQuery = ""
	u.Fragment = ""
	u.Path = "/robots.txt"

	robotsURL := u.String()

	// fetch robots.txt
	var body []byte
	body, _, err := utils.GetReq(client, robotsURL, 3, 5)
	if err != nil {
		return nil, fmt.Errorf("get robots.txt: %w", err)
	}

	// parse robots.txt for rules and sitemaps
	r := parseRobots(string(body), "*")
	return r, nil
}
