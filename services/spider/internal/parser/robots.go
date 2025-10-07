package parser

import (
	"strconv"
	"strings"
	"time"

	"spider/internal/utils"
)

func Robots(file, userAgent string) (allow, disallow []string, delay int, sitemaps []string) {
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
