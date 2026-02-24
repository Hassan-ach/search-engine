package crawler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"spider/internal/config"
	"spider/internal/entity"
	"spider/internal/parser"
	"spider/internal/store"
	"spider/internal/utils"
)

type Spider struct {
	config     *config.Config
	httpClient *http.Client
	store      *store.Store
	parser     *parser.Parser

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	fetchpool chan struct{}
	logger    *utils.Logger
}

type Metrics struct{}

func NewSpider(conf *config.Config) *Spider {
	ctx, cancel := context.WithCancel(context.Background())
	httpClient := &http.Client{Timeout: 10 * time.Second}
	logger := utils.NewMultiLogger(conf.App.LogsPath)

	s := &Spider{
		config:     conf,
		httpClient: httpClient,
		parser:     parser.NewParser(httpClient, logger),
		store:      store.NewStore(conf.Store, logger),
		wg:         sync.WaitGroup{},
		ctx:        ctx,
		cancel:     cancel,
		fetchpool:  make(chan struct{}, conf.App.MaxConcurrentFetch),
		logger:     logger,
	}

	fmt.Printf("Spider initialized: %+v\n", s)
	return s
}

func (s *Spider) Start(startUrls []string) {
	if err := s.store.Init(startUrls); err != nil {
		s.logger.Error("Failed to initialize store with start URLs", "error", err)
		return
	}

	for i := 0; i < s.config.App.MaxWorkers; i++ {
		s.wg.Add(1)
		s.logger.Info("Starting worker", "worker_id", i)
		go s.worker()
	}
}

func (s *Spider) Stop() {
	s.cancel()
	close(s.fetchpool)
	s.wg.Wait()
}

func (s *Spider) Close() {
	s.store.Close()
	s.logger.Close()
}

func (s *Spider) worker() {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Duration(s.config.App.ClawlerDelay) * time.Microsecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.crawl()
		}
	}
}

// crawl fetches a URL from the store, retrieves host metadata (from Redis or parser),
// processes the page, normalizes links (removing disallowed paths), updates
// Redis sets and counters, persists the page, and logs success or errors.
func (s *Spider) crawl() {
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	select {
	case s.fetchpool <- struct{}{}:
	case <-ctx.Done():
		fmt.Println("Worker timed out waiting for fetch slot")
		return
	}
	defer func() { <-s.fetchpool }()

	rawUrl, ok, err := s.store.GetNextUrl(ctx)
	if err != nil || !ok {
		fmt.Println("No URL fetched from store or error occurred:", err)
		return
	}
	s.logger.Info("Fetched URL from store", "url", rawUrl)

	u, err := url.Parse(rawUrl)
	if err != nil {
		return
	}

	var host *entity.Host
	host, err = s.getOrBuildHostMeta(u.Host)
	if err != nil {
		return
	}
	fmt.Printf("Host metadata for %s: %+v\n", u.Host, host)

	page, err := s.fetchAndParse(rawUrl, host.MaxRetry, host.Delay)
	if err != nil {
		s.logger.Error("Failed to process page", "url", rawUrl, "error", err)
		return
	}

	s.logger.Info(
		"Successfully processed page",
		"url",
		page.URL,
		"status_code",
		page.StatusCode,
		"links_found",
		len(page.Links),
		"image_count",
		len(page.Images),
	)

	normUrls := normalizeLinks(page.Links, host.NotAllowedPaths)
	page.Links = normUrls

	host.PagesCrawled++
	s.store.Persist(ctx, page, host)
}

func (s *Spider) getOrBuildHostMeta(h string) (host *entity.Host, err error) {
	host, ok, err := s.store.GetHostMetaData(s.ctx, strings.TrimPrefix(h, "www."))
	if !ok {
		// Host metadata missing; generate using parser
		host, err = s.NewHostMetaData(h)
		if err != nil {
			return
		}
	}
	return host, nil
}

func normalizeLinks(links []string, disallowed []string) []string {
	normUrls := utils.NewSet[string]()
	for _, x := range links {
		ur, err := url.Parse(x)
		if err != nil || isDisallowed(ur.Path, disallowed) {
			continue
		}
		normUrls.Add(ur.String())
	}
	return normUrls.GetAll()
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

func (s *Spider) fetchAndParse(
	u string,
	maxRetry, delay int,
) (*entity.Page, error) {
	s.logger.Info("Start crawling", "url", u)

	body, statusCode, err := utils.GetReq(s.httpClient, u, maxRetry, delay)
	if err != nil {
		// Failed to fetch page after retries
		// Suggest logging the URL and retry parameters
		return nil, fmt.Errorf("GET request failed: %w", err)
	}

	page, err := s.parser.ParseHTML(bytes.NewReader(body), u)
	if err != nil {
		// Failed to parse HTML
		return nil, fmt.Errorf("HTML parsing failed: %w", err)
	}
	s.logger.Info("Successfully crawled page", "url", u, "status_code", statusCode)

	if page.URL == "" {
		// Ensure Page.Url is always set
		page.URL = u
	}
	page.StatusCode = statusCode // Store HTTP status code
	page.HTML = body             // Store raw HTML

	return page, nil
}

func (s *Spider) NewHostMetaData(raw string) (host *entity.Host, err error) {
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}

	u, err := url.Parse(strings.TrimPrefix(raw, "www."))
	if err != nil {
		return
	}

	r, err := parser.NewRobots(s.httpClient, u)

	sitemaps := parser.FetchSitemaps(s.httpClient, r.SiteMaps, u)

	// create Host object
	host = &entity.Host{
		MaxRetry:        5,
		Delay:           r.CrawlDelay,
		MaxPages:        10,
		PagesCrawled:    0,
		Name:            u.Host,
		AllowedUrls:     r.Allow,
		NotAllowedPaths: r.Disallow,
	}

	s.logger.Info(
		"Generated host metadata",
		"host",
		host.Name,
		"max_retry",
		host.MaxRetry,
		"delay",
		host.Delay,
		"max_pages",
		host.MaxPages,
	)

	// persist in store
	cache := s.store.GetCache()
	cache.AddHostMetaData(s.ctx, host.Name, host)
	cache.AddUrls(s.ctx, sitemaps)

	return host, nil
}
