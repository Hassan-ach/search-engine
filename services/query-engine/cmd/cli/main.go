package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"query-engine/internal/config"
	"query-engine/internal/service/ranking"
	"query-engine/internal/store"
)

type mockSpeller struct{}

func (s mockSpeller) GetSuggestions(q string) []string {
	return strings.Split(q, " ")
}

func main() {
	start := time.Now()
	if len(os.Args) < 2 {
		fmt.Println("missing input")
		os.Exit(1)
	}

	conf, err := config.LoadConfig("../../.env")
	store := store.NewStore(conf.Store)

	// speller, err := service.NewAspellSpellingService()
	// if err != nil {
	// 	panic("failed to initialize speller: " + err.Error())
	// }
	speller := mockSpeller{}

	ranker := ranking.NewRankingService(&store, conf.Ranker)

	query := speller.GetSuggestions(os.Args[1])
	fmt.Printf("query: %s\n", query)

	s := time.Now()
	pages, err := ranker.Rank(query, 0)
	fmt.Printf("ranking took: %d ms\n", time.Since(s).Abs().Milliseconds())
	if err != nil {
		fmt.Printf("err: %v", err)
	}

	for _, page := range pages {
		fmt.Printf("%s: %f  Keywords: %v\n ", page.URL, page.GlobalScore, page.MetaData.Keywords)
	}
	fmt.Printf("total execution time: %d ms\n", time.Since(start).Abs().Milliseconds())
}
