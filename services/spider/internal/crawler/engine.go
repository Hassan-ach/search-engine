package crawler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/redis/go-redis/v9"

	"spider/internal/parser"
	"spider/internal/store"
	"spider/internal/utils"
)

type Engine struct {
	MaxWorkers  int
	Workers     int
	CacheClient *redis.Client
	Ctx         context.Context
}

func NewEngin() *Engine {
	ctx := context.Background()
	e := Engine{
		CacheClient: store.NewCacheClient(),
		Ctx:         ctx,
	}
	return &e
}

func (e *Engine) NewCrawler() (*Crawler, error) {
	domain, ok := store.NewUrl()
	if !ok {
		return nil, errors.New("Fail to retrive new Url from the store")
	}
	fmt.Printf("Creating new Crawler for %s\n", domain)
	u, err := url.Parse(domain)
	if err != nil {
		fmt.Println("Fail to Parse domain url")
	}
	u.Host = strings.TrimPrefix(u.Host, "www.")
	defer func() {
		e.CacheClient.HSet(e.Ctx, "inProg", u.Host, 1)
	}()
	u.RawQuery = ""
	u.Fragment = ""
	u.Path = "/robots.txt"
	fmt.Println(u.String())
	body, _, err := utils.GetReq(u.String(), 3, 5)
	if err != nil {
		fmt.Printf("Fail to GetReq\nERROR: %v\n", err)
		return nil, err
	}
	allowed, disallow, delay, sitemapsURLs := parser.Robots(string(body), "*")
	sitemaps := sitemapsProcess(sitemapsURLs, u.Host)
	discovedUrls := utils.NewSetQueu[string]()
	for _, v := range sitemaps {
		discovedUrls.Push(v)
	}

	crawler := Crawler{
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
	return &crawler, nil
}

func sitemapsProcess(s []string, host string) []string {
	fmt.Println("Start processing SiteMaps")
	var r []string
	for _, url := range s {
		file, _, err := utils.GetReq(url, 1, 5)
		if err != nil {
			fmt.Printf("error while processing Sitemap %s, error: %v\n", url, err)
			continue
		}
		d, err := parser.SitMap(file)
		if err != nil {
			fmt.Printf("error while processing Sitemap %s, error: %v\n", url, err)
			continue
		}
		for _, u := range d {
			x, ok := utils.NormalizeUrl(u, host)
			if !ok {
				continue
			}
			r = append(r, x)
		}

	}
	return r
}
