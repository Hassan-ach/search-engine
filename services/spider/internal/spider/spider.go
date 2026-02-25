package spider

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
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

	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	crawlerTimeout time.Duration
	crawlerDelay   time.Duration

	fetchpool chan struct{}
	logger    *utils.Logger
}

func NewSpider(conf *config.Config) *Spider {
	httpClient := &http.Client{Timeout: time.Duration(conf.App.HttpTimeout) * time.Second}
	logger := utils.NewMultiLogger(conf.App.LogsPath)
	ctx, cancel := context.WithCancel(context.Background())

	s := &Spider{
		config:         conf,
		httpClient:     httpClient,
		parser:         parser.NewParser(httpClient, logger),
		store:          store.NewStore(conf.Store, logger),
		wg:             sync.WaitGroup{},
		ctx:            ctx,
		cancel:         cancel,
		crawlerTimeout: time.Duration(conf.App.CrawlerTimeout) * time.Second,
		crawlerDelay:   time.Duration(conf.App.ClawlerDelay) * time.Microsecond,
		fetchpool:      make(chan struct{}, conf.App.MaxConcurrentFetch),
		logger:         logger,
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

	ticker := time.NewTicker(s.crawlerDelay)
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

func (s *Spider) crawl() {
	ctx, cancel := context.WithTimeout(s.ctx, s.crawlerTimeout)
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

	host, ok, err := s.store.GetHostMetaData(ctx, strings.TrimPrefix(u.Host, "www."))
	if !ok {
		// Host metadata missing; generate using parser
		host, err = s.newHostMetaData(ctx, u.Host)
		if err != nil {
			return
		}
	}
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

	normUrls := utils.ValidateLinks(page.Links, host.NotAllowedPaths)
	page.Links = normUrls

	host.PagesCrawled++
	s.store.Persist(ctx, page, host)
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

func (s *Spider) newHostMetaData(ctx context.Context, raw string) (host *entity.Host, err error) {
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}

	u, err := url.Parse(strings.TrimPrefix(raw, "www."))
	if err != nil {
		return
	}

	r, err := s.newRobots(u)

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
	cache.AddHostMetaData(ctx, host.Name, host)
	cache.AddUrls(ctx, sitemaps)

	return host, nil
}

func (s *Spider) newRobots(u *url.URL) (*entity.Robots, error) {
	h := strings.TrimPrefix(u.Host, "www.")

	u.Scheme = "https"
	u.Host = h
	u.RawQuery = ""
	u.Fragment = ""
	u.Path = "/robots.txt"

	robotsURL := u.String()

	// fetch robots.txt
	var body []byte
	body, _, err := utils.GetReq(s.httpClient, robotsURL, 3, 5)
	if err != nil {
		return nil, fmt.Errorf("get robots.txt: %w", err)
	}

	// parse robots.txt for rules and sitemaps
	r := s.parser.ParseRobots(string(body), "*")
	return r, nil
}
