package main

import (
	"spider/internal/crawler"
	"spider/internal/store"
	"spider/internal/utils"
)

func main() {
	defer func() {
		utils.Log.Close()
		store.Cache.Close()
	}()
	utils.Log.General().Info("Starting...")
	crawler.Run()
}
