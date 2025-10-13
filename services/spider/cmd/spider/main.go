package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"spider/internal/crawler"
	"spider/internal/store"
	"spider/internal/utils"
)

func main() {
	defer func() {
		utils.Log.Close()
		store.Cache.Close()
		store.DB.Close()
	}()
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Starting...")
	go crawler.Run([]string{})
	<-sigs
	fmt.Println("Exiting gracefully")
	crawler.Working = false
	crawler.WG.Wait()
	store.WG.Wait()
}
