package crawler

import (
	"fmt"

	"spider/internal/utils"
)

// Crawler is type that contain all the necessary informations about
// how Crawler can crawl the Domain
type Crawler struct {
	MaxRetry      uint8
	Delay         int
	MaxPages      uint8
	StartUrl      string
	DiscovedURLs  utils.Stack[string]
	VisitedURLs   utils.Set[string]
	AllowedUrls   []string
	NotAllwedUrls []string
}

type MetaData struct {
	Url         string
	Title       string
	Description string
}

type Domain struct {
	name string
}

func (c *Crawler) Run() {
	_, err := utils.GetReq(c.StartUrl, c.MaxRetry)
	if err != nil {
		fmt.Printf("Fail to Crawl this Url: %s", c.StartUrl)
	}
}
