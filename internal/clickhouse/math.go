package clickhouse

import (
	"encoding/json"
	"math"
	"sort"
)

func cosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		af := float64(a[i])
		bf := float64(b[i])
		dot += af * bf
		normA += af * af
		normB += bf * bf
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func sortHits(hits []ChunkHit) {
	sort.Slice(hits, func(i, j int) bool {
		return hits[i].Score > hits[j].Score
	})
}

func jsonUnmarshal(body []byte, out any) error {
	return json.Unmarshal(body, out)
}