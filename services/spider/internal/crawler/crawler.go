package crawler

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"time"

	"spider/internal/entity"
	"spider/internal/parser"
	"spider/internal/store"
	"spider/internal/utils"
)

// Run continuously executes the crawl process in an infinite loop.
// Each iteration fetches a URL, processes the page, updates the store,
// and logs success or failure.
func Run() {
	// store.AddUrls(starters)
	for count := 1; ; count++ {
		utils.Log.General().Debug("Starting crawl iteration", "iteration", count)
		crawl()
		// Small delay to avoid busy-loop
		// time.Sleep(100 * time.Millisecond)
	}
}

// crawl fetches a URL from the store, retrieves host metadata (from Redis or parser),
// processes the page, normalizes links (removing disallowed paths), updates
// Redis sets and counters, persists the page, and logs success or errors.
func crawl() {
	start := time.Now()
	log := utils.Log.General().With("operation", "Crawl")
	log.Info("Starting crawl process...")

	var (
		err    error
		ok     bool
		rawUrl string
	)

	// Start and deferred logging
	defer func() {
		execTime := time.Since(start)
		// Logs failed crawl if err is set
		if err != nil {
			log.Warn("Failed to crawl", "error", err, "url", rawUrl)
			return
		}
		// Logs crawl process completion and duration
		log.Info("Crawl process finished.", "execTime", execTime, "url", rawUrl)
	}()

	// Fetches next URL from Redis and handles empty set or errors.
	rawUrl, ok, _ = store.GetUrl()
	if !ok {
		err = fmt.Errorf("Failed to Get Url from store")
		// switch log to Cache context
		log = utils.Log.Cache().With("operation", "Crawl")
		return
	}

	u, err := url.Parse(rawUrl)
	if err != nil {
		// URL is invalid, skip crawl
		return
	}

	var host *entity.Host
	host, ok = store.GetHostMetaData(strings.TrimPrefix(u.Host, "www."))
	if !ok {
		// Host metadata missing; generate using parser
		host, err = parser.NewHostMetaDta(u.String())
		if err != nil {
			// switch log to parsing context
			log = utils.Log.Parsing().With("operation", "Crawl")
			return
		}
	}

	page, err := process(rawUrl, host.MaxRetry, host.Delay)
	if err != nil {
		// Page failed to download or parse
		return
	}

	normUrls := utils.NewSet[string]()
	for _, x := range page.Links.GetAll() {
		ur, err := url.Parse(x)
		if err != nil {
			// Skip invalid URL
			continue
		}

		skip := false
		for _, disallowed := range host.NotAllwedPaths {
			if strings.HasPrefix(ur.Path, disallowed) {
				skip = true
				break
			}
		}
		if skip {
			// Skip disallowed paths
			continue
		}

		normUrls.Add(ur.String())
	}

	normUrls.Print()
	page.Links = normUrls

	host.PagesCrawled++
	store.AddToVisitedUrl(rawUrl)
	store.AddToWaitedHost(host.Name, host.Delay)
	go store.Page(*page)
	store.AddUrls(page.Links.GetAll())
	log.Info("Page crawled successfully", "host", host.Name)
}

// process performs an HTTP GET request to the given URL, applies retry and delay logic,
// parses the HTML content into a Page struct, and returns the fully populated Page.
// Returns an error if the request fails or parsing fails.
func process(u string, maxRetry, delay int) (*entity.Page, error) {
	body, statusCode, err := utils.GetReq(u, maxRetry, delay)
	if err != nil {
		// Failed to fetch page after retries
		// Suggest logging the URL and retry parameters
		return nil, fmt.Errorf("GET request failed: %w", err)
	}

	page, err := parser.Html(bytes.NewReader(body))
	if err != nil {
		// Failed to parse HTML
		return nil, fmt.Errorf("HTML parsing failed: %w", err)
	}

	if page.Url == "" {
		// Ensure Page.Url is always set
		page.Url = u
	}
	page.StatusCode = statusCode // Store HTTP status code
	page.HTML = body             // Store raw HTML

	return page, nil
}
