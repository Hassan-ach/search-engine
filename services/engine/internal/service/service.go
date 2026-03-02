package service

import (
	"github.com/Hassan-ach/boogle/services/engine/internal/model"
	"github.com/Hassan-ach/boogle/services/engine/internal/store"
)

type Ranker interface {
	Rank(data *store.Data) ([]*model.Page, error)
}

type Speller interface {
	GetSuggestions(q string) []string
}
