package main

import (
	"fmt"
	"os"

	"spider/internal/crawler"
	"spider/internal/utils"
)

func main() {
	defer func() { utils.Log.Close() }()
	utils.Log.General().Info("Starting...")
	engin := crawler.NewEngin()
	defer engin.CacheClient.Close()
	c, err := engin.NewCrawler()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	c.Crawl()
}
