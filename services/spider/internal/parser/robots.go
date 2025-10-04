package parser

import (
	"fmt"
	"strconv"
	"strings"
)

func Robots(file, userAgent string) (allow, disallow []string, delay int, sitemaps []string) {
	fmt.Println("StartParsing robots.txt")
	if userAgent == "" {
		userAgent = "*"
	}

	lines := strings.Split(file, "\n")
	var active bool

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lower := strings.ToLower(line)

		switch {
		case strings.HasPrefix(lower, "user-agent:"):
			ua := strings.TrimSpace(line[len("User-agent:"):])
			active = (ua == userAgent || ua == "*")

		case strings.HasPrefix(lower, "disallow:") && active:
			disallow = append(disallow, strings.TrimSpace(line[len("Disallow:"):]))

		case strings.HasPrefix(lower, "allow:") && active:
			allow = append(allow, strings.TrimSpace(line[len("Allow:"):]))

		case strings.HasPrefix(lower, "crawl-delay:") && active:
			if d, err := strconv.Atoi(strings.TrimSpace(line[len("Crawl-delay:"):])); err == nil {
				delay = d
			}

		case strings.HasPrefix(lower, "sitemap:"):
			sitemaps = append(sitemaps, strings.TrimSpace(line[len("Sitemap:"):]))
		}
	}
	if delay == 0 {
		delay = 5
	}
	return
}
