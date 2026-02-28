package ranking

import (
	"fmt"

	"query-engine/internal/config/ranker"
	"query-engine/internal/model"
	"query-engine/internal/store"
)

type RankingService struct {
	store store.Store
	conf  ranker.RankingConfig
}

func NewRankingService(store *store.PsqlStore, conf ranker.RankingConfig) RankingService {
	return RankingService{
		store: store,
		conf:  conf,
	}
}

func (r RankingService) Rank(query []string, pageNum int) ([]*model.Page, error) {
	data, err := r.store.GetData(query, pageNum)
	if err != nil {
		return nil, err
	}

	pages, err := tfIdf(data)
	if err != nil {
		return nil, fmt.Errorf("failed to rank nodes: %w", err)
	}

	normalizeTFIDF(pages)

	rankedPages, err := sort(pages, 0.5, data.PageMapper, data.WordMapper)
	if err != nil {
		return nil, fmt.Errorf("failed to rank nodes: %w", err)
	}

	return rankedPages, nil
}
