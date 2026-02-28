package ranking

import (
	"fmt"
	"slices"

	"query-engine/internal/model"
	"query-engine/internal/util"

	"github.com/google/uuid"
)

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
		if a.GlobalScore < b.GlobalScore {
			return 1
		}
		if a.GlobalScore == b.GlobalScore {
			return 0
		}
		return -1
	})

	return pgs, nil
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
