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
	if strings.Index(u.Host, "www.") == 0 {
		u.Host = u.Host[4:]
	}
	u.RawQuery = ""
	u.Fragment = ""
	u.Path = "/robots.txt"
	fmt.Println(u.String())
	body, err := utils.GetReq(u.String(), 3)
	if err != nil {
		fmt.Printf("Fail to GetReq\nERROR: %v\n", err)
		return nil, err
	}
	allowed, disallow, delay, sitemapsURLs := parser.Robots(string(body), "*")
	sitemaps := sitemapsProcess(sitemapsURLs, u.Host)
	// fmt.Printf("ALLOWED URLs: %s\n", strings.Join(allowed, "\n"))
	// fmt.Printf("NOT ALLOWED URLs: %s\n", strings.Join(disallow, "\n"))
	// fmt.Printf("SITEMAPS URLs: %s\n", strings.Join(sitemaps, "\n"))
	discovedUrls := utils.NewSetQueu[string]()
	for _, v := range sitemaps {
		discovedUrls.Push(v)
	}
	// fmt.Printf("Discoved URLs:\n")
	// discovedUrls.Print()

	crawler := Crawler{
		MaxRetry:     5,
		Delay:        delay,
		MaxPages:     10,
		Host:         u.Host,
		DiscovedURLs: discovedUrls,
		VisitedURLs: &utils.Set[string]{
			Elements: map[string]bool{},
		},
		AllowedUrls:   allowed,
		NotAllwedUrls: disallow,
		CacheClient:   e.CacheClient,
		Ctx:           e.Ctx,
	}
	fmt.Println(crawler)
	return &crawler, nil
}

func sitemapsProcess(s []string, host string) []string {
	fmt.Println("Start processing SiteMaps")
	var r []string
	for _, url := range s {
		file, err := utils.GetReq(url, 1)
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
			x, err := urlClean(u, host)
			if err != nil {
				continue
			}
			r = append(r, x)
		}

	}
	return r
}

func urlClean(u string, host string) (string, error) {
	ob, err := url.Parse(u)
	if err != nil {
		fmt.Printf("Error while Cleaning URL: %s\n %v\n", u, err)
		return "", err
	}
	u, ok := utils.NormalizeUrl(ob.String(), host)
	if !ok {
		return "", errors.New("Not Allowed Urls")
	}
	return u, nil
}
