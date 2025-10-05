package main

import (
	"fmt"

	"spider/internal/crawler"
)

func main() {
	engin := crawler.NewEngin()
	defer engin.CacheClient.Close()
	c, err := engin.NewCrawler()
	if err != nil {
		fmt.Println(err)
	}
	c.Crawl()
}
