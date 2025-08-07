package typeahead

import (
	"context"
	"log"
	"strings"

	"github.com/redis/go-redis/v9"
)

type Typeahead struct {
	Redis *redis.Client
	Ctx   context.Context
}

func NewTypeahead(r *redis.Client) *Typeahead {
	return &Typeahead{
		Redis: r,
		Ctx:   context.Background(),
	}
}

// Get top-k suggestions based on prefix
func (ta *Typeahead) GetSuggestions(prefix string, k int) ([]string, error) {
	if len(prefix) < 3 || len(prefix) > 20 {
		return nil, nil
	}
	log.Fatal(prefix)

	// Fetch all suggestions (could be optimized with Trie if needed)
	suggestions, err := ta.Redis.ZRevRangeWithScores(ta.Ctx, "typeahead", 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var result []string
	for _, z := range suggestions {
		term := z.Member.(string)
		if strings.HasPrefix(strings.ToLower(term), strings.ToLower(prefix)) {
			result = append(result, term)
			if len(result) >= k {
				break
			}
		}
	}
	return result, nil
}

// Async increment score when term is searched
func (ta *Typeahead) IncrementTerm(term string) {
	if len(term) < 3 || len(term) > 20 {
		return
	}
	go func() {
		_ = ta.Redis.ZIncrBy(ta.Ctx, "typeahead", 1, term)
	}()
}

// Optional: Add new term to typeahead list manually (with initial score)
func (ta *Typeahead) AddTerm(term string, initialScore float64) error {
	if len(term) < 3 || len(term) > 20 {
		return nil
	}
	return ta.Redis.ZAdd(ta.Ctx, "typeahead", redis.Z{
		Score:  initialScore,
		Member: term,
	}).Err()
}
