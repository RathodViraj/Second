package index

import (
	"context"
	"fmt"
	"log"
	"math"
	"second/utils"
	"sort"

	"github.com/redis/go-redis/v9"
)

type InvertedIndex struct {
	Redis *redis.Client
	Ctx   context.Context
}

func NewInvertedIndex(redis *redis.Client) *InvertedIndex {
	return &InvertedIndex{
		Redis: redis,
		Ctx:   context.Background(),
	}
}

func (ii *InvertedIndex) AddDocument(docID, text string) error {
	words := utils.Tokenize(text)

	pipe := ii.Redis.Pipeline()
	for _, word := range words {
		key := fmt.Sprintf("index:%s", word)
		pipe.ZIncrBy(ii.Ctx, key, 1, docID)
	}
	pipe.SAdd(ii.Ctx, "indexed_docs", docID)
	_, err := pipe.Exec(ii.Ctx)
	if err == nil {
		ii.Redis.Incr(ii.Ctx, "meta:totalDocs")
	}

	return err
}

func (ii *InvertedIndex) Search(word string, offset, limit int64) ([]string, error) {
	key := fmt.Sprintf("index:%s", word)

	docs, err := ii.Redis.ZRevRange(ii.Ctx, key, offset, offset+limit-1).Result()
	if err != nil {
		return nil, err
	}

	return docs, nil
}

func (ii *InvertedIndex) SearchIF_IDF(query string, offset int) ([]string, error) {
	log.Println("Incoming query:", query)

	words := utils.Tokenize(query)
	log.Printf("Search words: %v", words)

	totalDocs, err := ii.Redis.Get(ii.Ctx, "meta:totalDocs").Int()
	if err != nil {
		return nil, err
	}
	log.Printf("Total documents: %d", totalDocs)

	scoreMap := make(map[string]float64)
	pipe := ii.Redis.Pipeline()

	// Stage ZCard and ZRangeWithScores for all words
	zcardCmds := make(map[string]*redis.IntCmd)
	zrangeCmds := make(map[string]*redis.ZSliceCmd)
	for _, w := range words {
		zcardCmds[w] = pipe.ZCard(ii.Ctx, "index:"+w)
		zrangeCmds[w] = pipe.ZRangeWithScores(ii.Ctx, "index:"+w, int64(offset), int64(offset+10-1))
	}

	_, err = pipe.Exec(ii.Ctx)
	if err != nil {
		return nil, err
	}

	for _, w := range words {
		docFreq, err := zcardCmds[w].Result()
		if err != nil || docFreq == 0 {
			log.Printf("No index found for word: %s", w)
			continue
		}

		log.Printf("Word '%s' has %d documents", w, docFreq)
		idf := math.Log(float64(totalDocs) / float64(1+docFreq))

		docScores, err := zrangeCmds[w].Result()
		if err != nil {
			log.Printf("Failed to get scores for word '%s': %v", w, err)
			continue
		}

		for _, doc := range docScores {
			docID := doc.Member.(string)
			tf := doc.Score
			scoreMap[docID] += tf * idf
			log.Printf("TF-IDF for doc %s (word: %s): %f", docID, w, tf*idf)
		}
	}

	// Sort and pick top 10
	type docScore struct {
		DocID string
		Score float64
	}
	var results []docScore
	for id, score := range scoreMap {
		results = append(results, docScore{id, score})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	top := min(len(results), 10)

	var rankedDocs []string
	for i := range top {
		rankedDocs = append(rankedDocs, results[i].DocID)
	}
	log.Printf("Ranked Docs: %v", rankedDocs)

	return rankedDocs, nil
}
