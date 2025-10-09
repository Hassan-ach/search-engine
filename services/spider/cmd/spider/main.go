package main

import (
	"fmt"
	"os"

	"spider/internal/crawler"
	"spider/internal/store"
	"spider/internal/utils"
)

func main() {
	defer func() { utils.Log.Close() }()
	utils.Log.General().Info("Starting...")
	engin := crawler.NewEngin()
	defer store.Cache.Close()
	c, err := engin.NewCrawler()
	if err != nil {
		utils.Log.General().Error(fmt.Sprintf("%v", err))
		os.Exit(1)
	}
	c.Crawl()
}
