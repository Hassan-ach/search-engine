package ranking

import (
	"fmt"

	"github.com/Hassan-ach/boogle/services/engine/internal/model"
	"github.com/Hassan-ach/boogle/services/engine/internal/store"
)

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
		idx, ok := pageMapper.GetIndex(page.ID)
		if !ok {
			// this should never happen
			panic(fmt.Sprintf("page with ID %s not found in page mapper", page.ID))
		}

		M[idx] = docVector(page, wordIdf, wordMapper, len(query))
	}

	queryVec := queryVector(query, wordIdf, wordMapper)
	docScores := make(map[*model.Page]float64, len(pages))
	for _, page := range pages {
		idx, ok := pageMapper.GetIndex(page.ID)
		if !ok {
			// this should never happen
			panic(fmt.Sprintf("page with ID %s not found in page mapper", page.ID))
		}
		score := cosineSimilarity(M[idx], queryVec)
		docScores[page] = score
	}

	return docScores, nil
}

func tf(query []string) map[string]int {
	tf := make(map[string]int, len(query))

	for _, word := range query {
		tf[word]++
	}
	return tf
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
