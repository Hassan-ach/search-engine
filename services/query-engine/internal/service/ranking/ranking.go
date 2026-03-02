package ranking

import (
	"fmt"

	"query-engine/internal/config/ranker"
	"query-engine/internal/model"
	"query-engine/internal/store"
)

type RankingService struct {
	conf ranker.RankingConfig
}

func NewRankingService(conf ranker.RankingConfig) RankingService {
	return RankingService{
		conf: conf,
	}
}

func (r RankingService) Rank(data *store.Data) ([]*model.Page, error) {
	pages, err := tfIdf(data)
	if err != nil {
		return nil, fmt.Errorf("failed to rank nodes: %w", err)
	}

	normalizeTFIDF(pages)

	rankedPages, err := sort(pages, 0.5)
	if err != nil {
		return nil, fmt.Errorf("failed to rank nodes: %w", err)
	}

	return rankedPages, nil
}
