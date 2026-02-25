package internal

import (
	"database/sql"
	"fmt"
	"math"
	"slices"
)

func tfIdf(
	query []string,
	data *Data,
	pageMapper *PageMapper,
	wordMapper *WordMapper,
) (map[*Page]float64, error) {
	wordIdf := data.Idf
	pages := data.Pages

	M := make([][]float64, len(pages))
	for _, page := range pages {
		idx, ok := pageMapper.GetIndex(page.URLID)
		if !ok {
			panic(fmt.Sprintf("page with URLID %s not found in page mapper", page.URLID))
		}

		M[idx] = docVector(page, wordIdf, wordMapper, len(query))
	}

	queryVec := queryVector(query, wordIdf, wordMapper)
	docScores := make(map[*Page]float64, len(pages))
	for _, page := range pages {
		idx, ok := pageMapper.GetIndex(page.URLID)
		if !ok {
			panic(fmt.Sprintf("page with URLID %s not found in page mapper", page.URLID))
		}
		score := cosineSimilarity(M[idx], queryVec)
		docScores[page] = score
	}

	return docScores, nil
}

func rank(pages map[*Page]float64,
	factor float64,
	pageMapper *PageMapper,
	wordMapper *WordMapper,
) ([]*Page, error) {
	pgs := make([]*Page, 0, len(pages))
	for p := range pages {
		p.GlobalScore = factor*pages[p] + (1-factor)*p.PRScore
		pgs = append(pgs, p)
	}

	slices.SortStableFunc(pgs, func(a, b *Page) int {
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

func ranking(conn *sql.DB, query []string) ([]*Page, error) {
	data, err := GetData(conn, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}

	pageMapper := NewPageMapper()
	for _, page := range data.Pages {
		pageMapper.MapUUID(page.URLID)
	}
	wordMapper := NewWordMapper()
	for _, w := range query {
		wordMapper.MapWord(w)
	}

	pages, err := tfIdf(query, data, pageMapper, wordMapper)
	if err != nil {
		return nil, fmt.Errorf("failed to rank nodes: %w", err)
	}

	normalizeTFIDF(pages)

	rankedPages, err := rank(pages, 0.5, pageMapper, wordMapper)
	if err != nil {
		return nil, fmt.Errorf("failed to rank nodes: %w", err)
	}

	return rankedPages, nil
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
		panic(fmt.Sprintf("magnitude sum is NaN for vec: %v", vec))
	}
	if sum == 0 {
		panic(fmt.Sprintf("magnitude sum is zero for vec: %v", vec))
	}
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
	page *Page,
	wordIdf map[string]float64,
	wordMapper *WordMapper,
	N int,
) []float64 {
	vec := make([]float64, N)

	for w, tf := range page.Words {
		idf := wordIdf[w]
		wIdx, ok := wordMapper.GetIndex(w)
		if !ok {
			panic(fmt.Sprintf("word %s not found in word mapper", w))
		}
		vec[wIdx] = float64(tf) * idf
	}
	return vec
}

func queryVector(
	query []string,
	wordIdf map[string]float64,
	wordMapper *WordMapper,
) []float64 {
	queryTF := tf(query)
	return docVector(&Page{Words: queryTF}, wordIdf, wordMapper, len(query))
}

func normalizeTFIDF(pages map[*Page]float64) {
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
