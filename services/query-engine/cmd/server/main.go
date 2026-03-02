package main

import (
	"embed"

	"query-engine/internal/config"
	"query-engine/internal/handlers"
	"query-engine/internal/service/ranking"
	"query-engine/internal/service/spellchecker"
	"query-engine/internal/store"

	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
)

var embeddedFiles embed.FS

func main() {
	conf, err := config.LoadConfig("../../.env")
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	store := store.NewStore(conf.Store)

	speller, err := spellchecker.NewAspellSpellingService()
	if err != nil {
		panic("failed to initialize speller: " + err.Error())
	}
	ranker := ranking.NewRankingService(conf.Ranker)

	homeHandler := &handlers.HomeHandler{}
	rankingHandler := handlers.NewSearchHandler(store, ranker, speller)

	e := echo.New()

	e.HTTPErrorHandler = handlers.HandleError
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.RequestID())
	e.Use(echoMiddleware.RequestLogger())
	// e.Use(echoMiddleware.CORS())
	// e.Use(echoMiddleware.Secure())
	e.Use(echoMiddleware.GzipWithConfig(echoMiddleware.GzipConfig{Level: 5}))
	e.Use(echoMiddleware.RateLimiter(echoMiddleware.NewRateLimiterMemoryStore(20)))

	e.Static("/public", "public")
	e.GET("/", homeHandler.Handle)
	e.GET("/search", rankingHandler.Handle)
	e.GET("/feeling-lucky", func(c *echo.Context) error {
		return c.String(200, "GOOD FOR YOU")
	})

	if err := e.Start(":1323"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
