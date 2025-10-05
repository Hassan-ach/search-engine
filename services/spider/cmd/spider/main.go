package main

import (
	"fmt"
	"os"

	"spider/internal/crawler"
)

func main() {
	engin := crawler.NewEngin()
	defer engin.CacheClient.Close()
	c, err := engin.NewCrawler()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	c.Crawl()
}
