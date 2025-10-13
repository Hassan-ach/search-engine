package crawler

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"spider/internal/entity"
	"spider/internal/parser"
	"spider/internal/store"
	"spider/internal/utils"
)

var (
	WG            sync.WaitGroup
	maxGoroutines               = 10
	CH            chan struct{} = make(chan struct{}, maxGoroutines)
	Working       bool          = true
)

// Run continuously executes the crawl process in an infinite loop.
// Each iteration fetches a URL, processes the page, updates the store,
// and logs success or failure.
func Run(starters []string) {
	store.AddUrls(starters)

	for Working {
		WG.Add(1)
		CH <- struct{}{}
		go func() {
			defer WG.Done()
			defer func() { <-CH }()
			crawl()
		}()
	}
}

// crawl fetches a URL from the store, retrieves host metadata (from Redis or parser),
// processes the page, normalizes links (removing disallowed paths), updates
// Redis sets and counters, persists the page, and logs success or errors.
func crawl() {
	start := time.Now()
	log := utils.Log.General().With("operation", "Crawl")

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
		// err = fmt.Errorf("Failed to Get Url from store")
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
	host, err = getOrBuildHostMeta(u.Host)
	if err != nil {
		// switch log to parsing context
		log = utils.Log.Parsing().With("operation", "Crawl")
		return
	}

	page, err := process(rawUrl, host.MaxRetry, host.Delay)
	if err != nil {
		// Page failed to download or parse
		return
	}

	normUrls := normalizeLinks(page.Links.GetAll(), host.NotAllwedPaths)
	page.Links = normUrls

	host.PagesCrawled++
	store.AddToVisitedUrl(rawUrl)
	store.AddToWaitedHost(host.Name, host.Delay)
	store.WG.Add(1)
	go func() {
		defer store.WG.Done()
		store.Page(*page)
	}()
	store.AddUrls(page.Links.GetAll())
	log.Info("Page crawled successfully", "host", host.Name)
}

func getOrBuildHostMeta(h string) (host *entity.Host, err error) {
	host, ok := store.GetHostMetaData(strings.TrimPrefix(h, "www."))
	if !ok {
		// Host metadata missing; generate using parser
		host, err = parser.NewHostMetaDta(h)
		if err != nil {
			return
		}
	}
	return host, nil
}

func normalizeLinks(links []string, disallowed []string) *utils.Set[string] {
	normUrls := utils.NewSet[string]()
	for _, x := range links {
		ur, err := url.Parse(x)
		if err != nil || isDisallowed(ur.Path, disallowed) {
			continue
		}
		normUrls.Add(ur.String())
	}
	return normUrls
}

func isDisallowed(path string, disallowed []string) bool {
	for _, d := range disallowed {
		// detect if pattern looks like regex
		if strings.ContainsAny(d, `.^$*+?[]|()`) {
			matched, err := regexp.MatchString(d, path)
			if err != nil {
				continue
			}
			if matched {
				return true
			}
		} else {
			if strings.HasPrefix(path, d) {
				return true
			}
		}
	}
	return false
}

// process performs an HTTP GET request to the given URL, applies retry and delay logic,
// parses the HTML content into a Page struct, and returns the fully populated Page.
// Returns an error if the request fails or parsing fails.
func process(u string, maxRetry, delay int) (*entity.Page, error) {
	utils.Log.General().Info("Start crawling", "url", u)
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

	if page.URL == "" {
		// Ensure Page.Url is always set
		page.URL = u
	}
	page.StatusCode = statusCode // Store HTTP status code
	page.HTML = body             // Store raw HTML

	return page, nil
}
