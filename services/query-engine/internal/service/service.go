package service

import (
	"query-engine/internal/model"
	"query-engine/internal/store"
)

type Ranker interface {
	Rank(data *store.Data) ([]*model.Page, error)
}

type Speller interface {
	GetSuggestions(q string) []string
}
