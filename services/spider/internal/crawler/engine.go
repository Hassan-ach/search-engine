package crawler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"spider/internal/parser"
	"spider/internal/store"
	"spider/internal/utils"
)

type Engine struct {
	MaxWorkers  int
	CacheClient *store.CacheClient
	Workers     int
	Ctx         context.Context
}

func NewEngin() *Engine {
	ctx := context.Background()
	e := Engine{
		CacheClient: store.Cache,
		Ctx:         ctx,
	}
	utils.Log.General().Info("Crawler engine initialized")

	return &e
}

func (e *Engine) NewCrawler() (crawler *Crawler, err error) {
	start := time.Now()
	log := utils.Log.General().With("operation", "NewCrawler")
	log.Info("Attempting to create new crawler")

	var host string

	defer func() {
		execTime := time.Since(start).Seconds()
		finalLog := log.With("host", host, "execTime", execTime)
		if err != nil {
			finalLog.Error("Crawler creation failed", "error", err)
			return
		}
		e.CacheClient.HSet(e.Ctx, "inProg", host, 1)
		finalLog.Info("Crawler creation completed successfully", "crawler", crawler.String())
	}()

	var ok bool

	host, ok = store.NewHost()
	if !ok {
		err = errors.New("failed to retrieve new host from the store")
		utils.Log.DB().Error(err.Error())
		return nil, err
	}
	utils.Log.DB().Info("New host retrieved", "host", host)

	var u *url.URL
	u, err = url.Parse(host)
	if err != nil {
		return nil, err
	}
	u.Host = strings.TrimPrefix(u.Host, "www.")
	u.RawQuery = ""
	u.Fragment = ""
	u.Path = "/robots.txt"
	robotsURL := u.String()

	var body []byte
	body, _, err = utils.GetReq(robotsURL, 3, 5)
	if err != nil {
		err = fmt.Errorf("failed to get robots.txt: %w", err)
		return nil, err
	}

	allowed, disallow, delay, sitemapsURLs := parser.Robots(string(body), "*")
	sitemaps := sitemapsProcess(sitemapsURLs, u.Host)
	discovedUrls := utils.NewSetQueu[string]()
	for _, v := range sitemaps {
		discovedUrls.Push(v)
	}

	crawler = &Crawler{
		Host: Host{
			MaxRetry:       5,
			Delay:          delay,
			MaxPages:       10,
			Name:           u.Host,
			AllowedUrls:    allowed,
			NotAllwedPaths: disallow,
			DiscovedURLs:   discovedUrls,
			VisitedURLs:    utils.NewSet[string](),
		},
		CacheClient: e.CacheClient,
		Ctx:         e.Ctx,
	}
	log.Debug("Crawler object instantiated", "crawler", crawler.String())

	return crawler, nil
}

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
			"execTime", time.Since(start).Seconds(),
		)
	}()

	for _, sitemapURL := range s {
		file, _, err := utils.GetReq(sitemapURL, 1, 5)
		if err != nil {
			failedSites++
			utils.Log.Network().Warn("Failed to fetch sitemap", "url", sitemapURL, "error", err)
			continue
		}

		d, err := parser.SitMap(file)
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
