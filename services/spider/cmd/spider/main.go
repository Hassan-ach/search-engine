package main

import (
	"fmt"

	"spider/internal/crawler"
)

func main() {
	c, err := crawler.NewCrawler()
	if err != nil {
		fmt.Println(err)
	}
	c.Crawl()
}
