package entity

import "fmt"

type Host struct {
	MaxRetry       int      // Maximum retries per URL
	MaxPages       int      // Maximum pages to crawl for this host
	PagesCrawled   int      // Pages already crawled
	Delay          int      // Delay between requests in seconds
	Name           string   // Hostname
	AllowedUrls    []string // URL patterns allowed to crawl
	NotAllwedPaths []string // Paths disallowed to crawl (typo kept for backward compatibility)
}

// String returns a human-readable summary of the host configuration.
func (h *Host) String() string {
	return fmt.Sprintf(
		"Host Info\n"+
			"  Host: %s\n"+
			"  MaxRetry: %d\n"+
			"  MaxPages: %d\n"+
			"  PagesCrawled: %d\n"+
			"  Delay: %ds\n"+
			"  Allowed URLs: %v\n"+
			"  Disallowed Paths: %v\n",
		h.Name,
		h.MaxRetry,
		h.MaxPages,
		h.PagesCrawled,
		h.Delay,
		h.AllowedUrls,
		h.NotAllwedPaths,
	)
}
