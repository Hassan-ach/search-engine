package service

import (
	"fmt"
	"math"
	"slices"

	"query-engine/internal/config"
	"query-engine/internal/model"
	"query-engine/internal/store"
	"query-engine/internal/util"

	"github.com/google/uuid"
)

type RankingService struct {
	store store.Store
	conf  config.RankingConfig
}

func NewRankingService(store *store.PsqlStore, conf config.RankingConfig) RankingService {
	return RankingService{
		store: store,
		conf:  conf,
	}
}

func (r RankingService) Rank(query []string) ([]*model.Page, error) {
	data, err := r.store.GetData(query)
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

func tfIdf(
	data *store.Data,
) (map[*model.Page]float64, error) {
	wordIdf := data.Idf
	pages := data.Pages
	pageMapper := data.PageMapper
	wordMapper := data.WordMapper
	query := wordMapper.GetValues()

	M := make([][]float64, len(pages))
	for _, page := range pages {
		idx, ok := pageMapper.GetIndex(page.URLID)
		if !ok {
			// this should never happen
			panic(fmt.Sprintf("page with URLID %s not found in page mapper", page.URLID))
		}

		M[idx] = docVector(page, wordIdf, wordMapper, len(query))
	}

	queryVec := queryVector(query, wordIdf, wordMapper)
	docScores := make(map[*model.Page]float64, len(pages))
	for _, page := range pages {
		idx, ok := pageMapper.GetIndex(page.URLID)
		if !ok {
			// this should never happen
			panic(fmt.Sprintf("page with URLID %s not found in page mapper", page.URLID))
		}
		score := cosineSimilarity(M[idx], queryVec)
		docScores[page] = score
	}

	return docScores, nil
}

func sort(pages map[*model.Page]float64,
	factor float64,
	pageMapper util.Mapper[uuid.UUID],
	wordMapper util.Mapper[string],
) ([]*model.Page, error) {
	pgs := make([]*model.Page, 0, len(pages))
	for p := range pages {
		p.GlobalScore = factor*pages[p] + (1-factor)*p.PRScore
		pgs = append(pgs, p)
	}

	slices.SortStableFunc(pgs, func(a, b *model.Page) int {
		if a.GlobalScore > b.GlobalScore {
			return 1
		}
		if a.GlobalScore == b.GlobalScore {
			return 0
		}
		return -1
	})

	return pgs, nil
}

func cosineSimilarity(vecA, vecB []float64) float64 {
	return dotProduct(vecA, vecB) / (magnitude(vecA) * magnitude(vecB))
}

func dotProduct(vecA, vecB []float64) float64 {
	dot := 0.0
	for i := range vecA {
		dot += vecA[i] * vecB[i]
	}
	if math.IsNaN(dot) {
		// this should never happen
		panic(fmt.Sprintf("dot product is NaN for vecA: %v, vecB: %v", vecA, vecB))
	}

	return dot
}

func magnitude(vec []float64) float64 {
	sum := 0.0
	for _, v := range vec {
		sum += v * v
	}
	if math.IsNaN(sum) {
		// this should never happen
		panic(fmt.Sprintf("magnitude sum is NaN for vec: %v", vec))
	}
	// if sum == 0 {
	// 	panic(fmt.Sprintf("magnitude sum is zero for vec: %v", vec))
	// }
	return math.Sqrt(sum)
}

func tf(query []string) map[string]int {
	tf := make(map[string]int, len(query))

	for _, word := range query {
		tf[word]++
	}
	return tf
}

func docVector(
	page *model.Page,
	wordIdf map[string]float64,
	wordMapper util.Mapper[string],
	N int,
) []float64 {
	vec := make([]float64, N)

	for w, tf := range page.Words {
		idf := wordIdf[w]
		wIdx, ok := wordMapper.GetIndex(w)
		if !ok {
			// this should never happen
			panic(fmt.Sprintf("word %s not found in word mapper", w))
		}
		vec[wIdx] = float64(tf) * idf
	}
	return vec
}

func queryVector(
	query []string,
	wordIdf map[string]float64,
	wordMapper util.Mapper[string],
) []float64 {
	queryTF := tf(query)
	return docVector(&model.Page{Words: queryTF}, wordIdf, wordMapper, len(query))
}

func normalizeTFIDF(pages map[*model.Page]float64) {
	if len(pages) == 0 {
		return
	}

	var minTFIDF, maxTFIDF float64
	first := true

	for _, v := range pages {
		if first {
			minTFIDF = v
			maxTFIDF = v
			first = false
			continue
		}
		if v > maxTFIDF {
			maxTFIDF = v
		}
		if v < minTFIDF {
			minTFIDF = v
		}
	}

	dom := maxTFIDF - minTFIDF

	for p, v := range pages {
		if dom == 0 {
			pages[p] = 0
		} else {
			pages[p] = (v - minTFIDF) / dom
		}
	}
}
