package service

import (
	"query-engine/internal/model"
)

type Ranker interface {
	Rank(query []string) ([]*model.Page, error)
}

type Speller interface {
	GetSuggestions(q string) []string
}
