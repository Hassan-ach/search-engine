package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Hassan-ach/boogle/services/spider/internal/config"
	"github.com/Hassan-ach/boogle/services/spider/internal/spider"
)

type Spider struct {
	Config *config.Config
}

func main() {
	conf, _ := config.LoadConfig()

	spider := spider.NewSpider(conf)
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
