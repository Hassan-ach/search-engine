package main

import (
	"fmt"

	"query-engine/internal/config"
	"query-engine/internal/handlers"
	"query-engine/internal/service/ranking"
	"query-engine/internal/service/spellchecker"
	"query-engine/internal/store"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	conf, err := config.LoadConfig("../../.env")
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	e := echo.New()
	e.Use(middleware.RequestLogger())

	store := store.NewStore(conf.Store)

	speller, err := spellchecker.NewAspellSpellingService()
	if err != nil {
		panic("failed to initialize speller: " + err.Error())
	}

	ranker := ranking.NewRankingService(&store, conf.Ranker)

	homeHandler := &handlers.HomeHandler{}
	rankingHandler := handlers.NewSearchHandler(ranker, speller)

	fmt.Printf("starting server on port 1323\n")

	e.Static("/public", "public")
	e.GET("/", homeHandler.Handle)
	e.GET("/search", rankingHandler.Handle)

	if err := e.Start(":1323"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
