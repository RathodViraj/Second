package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var doc_repo *DocumentRepo

type TrendingDoc struct {
	ID    string
	Views int
	Title string
}

type TrendingRepo struct {
	rdb *redis.Client
	ctx context.Context
}

func NewTrendingRepo(rdb *redis.Client) *TrendingRepo {
	doc_repo = &DocumentRepo{rdb: rdb}

	return &TrendingRepo{
		rdb: rdb,
		ctx: context.Background(),
	}
}

func (r *TrendingRepo) GetTrendingDocs() ([]TrendingDoc, error) {
	pairs, err := r.rdb.ZRevRangeWithScores(r.ctx, "trending_docs", 0, 49).Result()
	if err != nil {
		return nil, err
	}

	var out []TrendingDoc
	for _, p := range pairs {
		ID := p.Member.(string)
		title, err := doc_repo.GetDocTitle(ID)
		if err != nil {
			continue
		}
		out = append(out, TrendingDoc{
			ID:    ID,
			Title: title,
			Views: int(p.Score),
		})
	}
	return out, nil
}

func (r *TrendingRepo) SlideWindow() error {
	expired := time.Now().UTC().Add(-time.Hour)
	minute := expired.Format("200601021504")

	docsKey := fmt.Sprintf("doc:views:%s:docs", minute)
	docIDs, err := r.rdb.SMembers(r.ctx, docsKey).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	for _, docID := range docIDs {
		perMinKey := fmt.Sprintf("doc:%s:views:%s", docID, minute)
		count, _ := r.rdb.Get(r.ctx, perMinKey).Int()

		if count > 0 {
			r.rdb.ZIncrBy(r.ctx, "trending_docs", float64(-count), docID)
		}
	}

	r.rdb.Del(r.ctx, docsKey)
	return nil
}
