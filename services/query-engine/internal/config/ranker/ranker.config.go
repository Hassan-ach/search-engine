package ranker

import "github.com/Hassan-ach/boogle/services/engine/internal/util"

type RankingConfig struct {
	MaxResults int
	WeightTF   float64
}

func NewRankingConfig() RankingConfig {
	maxResults := util.GetIntWithDefault("RANKER_MAX_RESULTS", 100)
	weightTF := util.GetFloatWithDefault("RANKER_WEIGHT_TF", 0.5)

	return RankingConfig{
		maxResults,
		weightTF,
	}
}
