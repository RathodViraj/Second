package repository

import (
	"context"
	"encoding/json"
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

func (r *DocumentRepo) AddDocumentToDB(doc *model.Document) error {
	doc.Created = time.Now()

	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	collection := db.GetCollection("secondDB", "documents")

	result, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}

	objectID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return errors.New("failed to parse inserted ID")
	}
	doc.ID = objectID

	return nil
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
		return nil
	}

	retryData := map[string]string{
		"docID": docID,
		"text":  text,
	}
	data, _ := json.Marshal(retryData)

	if pushErr := r.rdb.LPush(r.ctx, "retry:index_queue", data).Err(); pushErr != nil {
		return fmt.Errorf("index failed: %v, retry queue push failed: %v", err, pushErr)
	}

	return fmt.Errorf("index failed, added to retry queue: %w", err)
}

func (r *DocumentRepo) GetDocumentByID(docID string) (model.Document, error) {
	collection := db.GetCollection("secondDB", "documents")
	var doc model.Document

	objectID, err := primitive.ObjectIDFromHex(docID)
	if err != nil {
		return model.Document{}, err
	}

	if err := collection.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&doc); err != nil {
		return model.Document{}, err
	}

	score := r.rdb.ZScore(r.ctx, "trending_docs", docID).Val()
	r.rdb.ZAdd(r.ctx, "trending_docs",
		redis.Z{
			Score:  score + 1,
			Member: docID,
		})

	return doc, nil
}

func (r *DocumentRepo) GetDocTitle(docID string) (string, error) {
	collection := db.GetCollection("secondDB", "documents")
	var doc model.Document

	if err := collection.FindOne(context.TODO(), bson.M{"id": docID}).Decode(&doc); err != nil {
		return "", err
	}

	return doc.Title, nil
}

func (r *DocumentRepo) SearchIF_IDF(query string, offset int) ([]model.Document, error) {
	words := utils.Tokenize(query)
	log.Printf("Searching for words: %v", words)

	totalDocs, err := r.rdb.Get(r.ctx, "meta:totalDocs").Int()
	if err != nil {
		return nil, err
	}

	scoreMap := make(map[string]float64)
	pipe := r.rdb.Pipeline()

	zcardCmds := make(map[string]*redis.IntCmd)
	zrangeCmds := make(map[string]*redis.ZSliceCmd)
	for _, w := range words {
		zcardCmds[w] = pipe.ZCard(r.ctx, "index:"+w)
		zrangeCmds[w] = pipe.ZRangeWithScores(r.ctx, "index:"+w, 0, -1)
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

	var rankedDocs []model.Document
	for i := 0; i < top; i++ {

		objectID, err := primitive.ObjectIDFromHex(results[i].DocID)
		if err != nil {
			log.Printf("Invalid ObjectID: %s, error: %v\n", results[i].DocID, err)
			continue
		}

		doc, err := r.GetDocumentByID(objectID.Hex())
		if err != nil {
			log.Printf("Could not fetch doc for ID %s: %v\n", objectID.Hex(), err)
			continue
		}

		rankedDocs = append(rankedDocs, doc)
	}

	for _, w := range words {
		r.rdb.ZIncrBy(r.ctx, "typeahead", 1, w)
	}

	return rankedDocs, nil
}

func (r *DocumentRepo) StartRetryWorker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		for {
			res, err := r.rdb.BRPop(r.ctx, 1*time.Second, "retry:index_queue").Result()
			if err == redis.Nil {
				break
			}
			if err != nil {
				log.Println("Error popping from retry queue:", err)
				break
			}

			var job map[string]string
			if err := json.Unmarshal([]byte(res[1]), &job); err != nil {
				log.Println("Invalid job data:", err)
				continue
			}

			if err := r.IndexDoc(job["docID"], job["text"]); err != nil {
				log.Println("Retry failed for doc:", job["docID"], err)
			}
		}
	}
}
