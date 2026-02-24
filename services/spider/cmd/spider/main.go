package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"spider/internal/config"
	"spider/internal/crawler"
)

type Spider struct {
	Config *config.Config
}

func main() {
	conf, err := config.LoadConfig("../../.env")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	spider := crawler.NewSpider(conf)
	defer func() {
		spider.Close()
	}()

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go spider.Start([]string{"https://en.wikipedia.org/wiki/Hairy_ball_theorem"})
	<-sigs
	fmt.Println("Exiting gracefully")

	spider.Stop()
}
