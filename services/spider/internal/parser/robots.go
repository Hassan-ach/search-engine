package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"spider/internal/entity"
	"spider/internal/store"
	"spider/internal/utils"
)

// NewHostMetaDta creates a new Host metadata object for the given raw URL.
// It fetches the site's robots.txt, extracts allowed/disallowed paths, crawl delay,
// and sitemaps. It persists the Host object in the store and returns it.
// Logs the full process including execution time and success/failure.
func NewHostMetaDta(raw string) (host *entity.Host, err error) {
	start := time.Now()
	log := utils.Log.General()
	log = log.With("operation", "NewHostMetaDta")
	log.Info("Attempting to create new Host Meta Data Object")

	u, err := url.Parse(raw)
	if err != nil {
		return
	}

	var h string
	defer func() {
		execTime := time.Since(start)
		finalLog := log.With("host", h, "execTime", execTime)
		if err != nil {
			finalLog.Error("Host Meta Data retrieve failed", "error", err)
			return
		}
		finalLog.Info("Host Meta Data retrieve completed successfully")
		finalLog.Debug(
			"Host Meta Data retrieve completed successfully",
			"Host Meta Data",
			host.String(),
		)
	}()

	// normalize host
	h = strings.TrimPrefix(u.Host, "www.")
	u.Host = h
	u.RawQuery = ""
	u.Fragment = ""
	u.Path = "/robots.txt"
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	// fetch robots.txt
	robotsURL := u.String()
	var body []byte
	body, _, err = utils.GetReq(robotsURL, 3, 5)
	if err != nil {
		err = fmt.Errorf("failed to get robots.txt: %w", err)
		return nil, err
	}

	// parse robots.txt for rules and sitemaps
	allowed, disallow, delay, sitemapsURLs := parseRobots(string(body), "*")
	sitemaps := sitemapsProcess(sitemapsURLs, u.Host)

	// create Host object
	host = &entity.Host{
		MaxRetry:       5,
		Delay:          delay,
		MaxPages:       10,
		PagesCrawled:   0,
		Name:           u.Host,
		AllowedUrls:    allowed,
		NotAllwedPaths: disallow,
	}

	// persist in store
	store.AddHostMetaData(host.Name, host)
	store.AddUrls(sitemaps)

	return host, nil
}

// parseRobots reads the robots.txt file content for a given user-agent.
// Returns lists of allowed URLs, disallowed paths, crawl delay, and sitemap URLs.
// Logs start, end, and details about each rule parsed.
func parseRobots(file, userAgent string) (allow, disallow []string, delay int, sitemaps []string) {
	log := utils.Log.Parsing().With("operation", "Robots")
	start := time.Now()
	log.Info("Starting robots.txt parsing")

	defer func() {
		log.Info(
			"Finished parsing robots.txt",
			"execTime", time.Since(start),
			"allowRules", len(allow),
			"disallowRules", len(disallow),
			"sitemapsFound", len(sitemaps),
			"finalCrawlDelay", delay,
		)
	}()

	if userAgent == "" {
		userAgent = "*"
	}

	lines := strings.Split(file, "\n")
	var activeAgent bool

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lower := strings.ToLower(line)

		switch {
		case strings.HasPrefix(lower, "user-agent:"):
			ua := strings.TrimSpace(line[len("User-agent:"):])
			activeAgent = (ua == userAgent || ua == "*")
			log.Debug(
				"Processing User-agent",
				"line",
				line,
				"targetAgent",
				userAgent,
				"isActive",
				activeAgent,
			)

		case strings.HasPrefix(lower, "disallow:") && activeAgent:
			rule := strings.TrimSpace(line[len("Disallow:"):])
			disallow = append(disallow, rule)
			log.Debug("Found Disallow rule", "rule", rule)

		case strings.HasPrefix(lower, "allow:") && activeAgent:
			rule := strings.TrimSpace(line[len("Allow:"):])
			allow = append(allow, rule)
			log.Debug("Found Allow rule", "rule", rule)

		case strings.HasPrefix(lower, "crawl-delay:") && activeAgent:
			delayStr := strings.TrimSpace(line[len("Crawl-delay:"):])
			if d, err := strconv.Atoi(delayStr); err == nil {
				delay = d
				log.Debug("Found Crawl-delay", "delay", d)
			} else {
				log.Warn("Failed to parse Crawl-delay value", "value", delayStr, "error", err)
			}

		case strings.HasPrefix(lower, "sitemap:"):
			sitemapURL := strings.TrimSpace(line[len("Sitemap:"):])
			sitemaps = append(sitemaps, sitemapURL)
			log.Debug("Found Sitemap", "url", sitemapURL)
		}
	}

	if delay == 0 {
		delay = 5
		log.Debug("No Crawl-delay found for active agent, applying default", "defaultDelay", delay)

	}

	return
}

// sitemapsProcess fetches and parses sitemap URLs for a host.
// Returns a flattened list of normalized URLs found across all sitemaps.
// Logs start, end, execution time, number of extracted links, failed sitemaps.
func sitemapsProcess(s []string, host string) []string {
	start := time.Now()
	log := utils.Log.General().With("operation", "sitemapsProcess", "host", host)
	log.Info("Starting sitemap processing")

	var r []string
	failedSites := 0

	defer func() {
		log.Info(
			"Sitemap processing finished",
			"extractedLinks", len(r),
			"failedSitemaps", failedSites,
			"totalSitemaps", len(s),
			"execTime", time.Since(start),
		)
	}()

	for _, sitemapURL := range s {
		file, _, err := utils.GetReq(sitemapURL, 1, 5)
		if err != nil {
			failedSites++
			utils.Log.Network().Warn("Failed to fetch sitemap", "url", sitemapURL, "error", err)
			continue
		}

		d, err := sitMap(file)
		if err != nil {
			failedSites++
			utils.Log.Parsing().Warn("Failed to parse sitemap", "url", sitemapURL, "error", err)
			continue
		}

		for _, u := range d {
			x, ok := utils.NormalizeUrl(u, host)
			if !ok {
				log.Debug("URL normalization failed", "rawURL", u, "host", host)
				continue
			}
			r = append(r, x)
		}

	}
	return r
}
