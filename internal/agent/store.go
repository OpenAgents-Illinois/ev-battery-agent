package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
)

// embeddingCache is the on-disk format for cached embeddings.
type embeddingCache struct {
	Version string          `json:"version"`
	Entries []cacheEntry    `json:"entries"`
}

type cacheEntry struct {
	Text     string         `json:"text"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Vector   []float32      `json:"vector"`
}

const cacheVersion = "go-v1"

// vectorStore holds documents and their pre-computed embeddings.
// Supports cosine-similarity search without re-embedding stored docs.
type vectorStore struct {
	entries []cacheEntry
}

// addChunks embeds the given text chunks using the provided embedder and adds them to the store.
func (vs *vectorStore) addChunks(ctx context.Context, embedder embedderFunc, chunks []textChunk) error {
	const batchSize = 50
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.text
	}

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]
		vectors, err := embedder(ctx, batch)
		if err != nil {
			return fmt.Errorf("embed batch %d-%d: %w", i, end, err)
		}
		for j, v := range vectors {
			vs.entries = append(vs.entries, cacheEntry{
				Text:   batch[j],
				Vector: v,
			})
		}
	}
	return nil
}

// search returns the top-k most similar documents to the query.
func (vs *vectorStore) search(ctx context.Context, queryVec []float32, k int) []string {
	type scored struct {
		text  string
		score float64
	}
	results := make([]scored, len(vs.entries))
	for i, e := range vs.entries {
		results[i] = scored{e.Text, cosineSimilarity(queryVec, e.Vector)}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})
	if k > len(results) {
		k = len(results)
	}
	texts := make([]string, k)
	for i := range texts {
		texts[i] = results[i].text
	}
	return texts
}

func (vs *vectorStore) saveCache(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	cache := embeddingCache{Version: cacheVersion, Entries: vs.entries}
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func loadCache(path string) (*vectorStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cache embeddingCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	if cache.Version != cacheVersion {
		return nil, fmt.Errorf("cache version mismatch: %q (want %q)", cache.Version, cacheVersion)
	}
	return &vectorStore{entries: cache.Entries}, nil
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
