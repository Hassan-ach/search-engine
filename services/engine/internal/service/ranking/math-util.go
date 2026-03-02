package ranking

import (
	"fmt"
	"math"
)

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
