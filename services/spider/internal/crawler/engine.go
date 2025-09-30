package crawler

import (
	"fmt"
	"net/url"

	"spider/internal/parser"
	"spider/internal/utils"
)

type Engine struct {
	Urls []string
}

func NewCrawler(domain string) (*Crawler, error) {
	url, err := url.Parse(domain)
	if err != nil {
		fmt.Println("Fail to Parse domain url")
	}
	url.RawQuery = ""
	url.Fragment = ""
	url.Path = "/robots.txt"
	fmt.Println(url.String())
	body, err := utils.GetReq(url.String(), 1)
	if err != nil {
		fmt.Printf("Fail to GetReq\n%v", err)
	}
	allowed, disallow, delay, sitemaps := parser.Robots(string(body), "*")
	crawler := Crawler{
		MaxRetry: 5,
		Delay:    delay,
		MaxPages: 10,
		StartUrl: domain,
		DiscovedURLs: utils.Stack[string]{
			Elements: sitemaps,
		},
		VisitedURLs: utils.Set[string]{
			Elements: make(map[string]bool),
		},
		AllowedUrls:   allowed,
		NotAllwedUrls: disallow,
	}
	fmt.Println(crawler)
	return &crawler, nil
}
