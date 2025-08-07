package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"second/db"
	"second/model"
	"second/utils"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DocumentRepo struct {
	rdb *redis.Client
	ctx context.Context
}

type docScore struct {
	DocID string
	Score float64
}

func NewDocumentRepo(rdb *redis.Client) *DocumentRepo {
	return &DocumentRepo{
		rdb: rdb,
		ctx: context.Background(),
	}
}

func (r *DocumentRepo) AddDocumentToDB(doc model.Document) (string, error) {
	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	collection := db.GetCollection("secondDB", "documents")

	if doc.Created.IsZero() {
		doc.Created = time.Now()
	}

	result, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}

	objectID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("failed to parse inserted ID")
	}

	return objectID.Hex(), nil
}

func (r *DocumentRepo) IndexDoc(docID, text string) error {
	words := utils.Tokenize(text)

	pipe := r.rdb.Pipeline()
	for _, word := range words {
		key := fmt.Sprintf("index:%s", word)
		pipe.ZIncrBy(r.ctx, key, 1, docID)
	}
	pipe.SAdd(r.ctx, "indexed_docs", docID)
	_, err := pipe.Exec(r.ctx)
	if err == nil {
		r.rdb.Incr(r.ctx, "meta:totalDocs")
	}

	return err
}

func (r *DocumentRepo) GetDocumentByID(docID string) (model.Document, error) {
	collection := db.GetCollection("mydocsdb", "documents")
	var doc model.Document

	if err := collection.FindOne(context.TODO(), bson.M{"_id": docID}).Decode(&doc); err != nil {
		return model.Document{}, err
	}

	return doc, nil
}

func (r *DocumentRepo) GetDocTitle(docID string) (string, error) {
	collection := db.GetCollection("mydocsdb", "documents")
	var doc model.Document

	if err := collection.FindOne(context.TODO(), bson.M{"_id": docID}).Decode(&doc); err != nil {
		return "", err
	}

	return doc.Title, nil
}

func (r *DocumentRepo) SearchIF_IDF(query string, offset int) ([]string, error) {
	words := utils.Tokenize(query)

	totalDocs, err := r.rdb.Get(r.ctx, "meta:totalDocs").Int()
	if err != nil {
		return nil, err
	}

	scoreMap := make(map[string]float64)
	pipe := r.rdb.Pipeline()

	// Stage ZCard and ZRangeWithScores for all words
	zcardCmds := make(map[string]*redis.IntCmd)
	zrangeCmds := make(map[string]*redis.ZSliceCmd)
	for _, w := range words {
		zcardCmds[w] = pipe.ZCard(r.ctx, "index:"+w)
		zrangeCmds[w] = pipe.ZRangeWithScores(r.ctx, "index:"+w, int64(offset), int64(offset+10-1))
	}

	_, err = pipe.Exec(r.ctx)
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
		}
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

	return rankedDocs, nil
}
