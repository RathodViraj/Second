package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	RedisClient *redis.Client
	Script      *redis.Script
	Rate        float64
	Capacity    float64
}

func NewRateLimiter(rdb *redis.Client, rate float64, capacity float64) (*RateLimiter, error) {
	scriptBytes, err := os.ReadFile("./middleware/leaky_bucket.lua")
	if err != nil {
		return nil, fmt.Errorf("failed to load lua script: %w", err)
	}
	script := redis.NewScript(string(scriptBytes))
	return &RateLimiter{RedisClient: rdb, Script: script, Rate: rate, Capacity: capacity}, nil
}

func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := "rl:ip:" + ip
		now := float64(time.Now().Unix())

		ctx := context.Background()
		result, err := r.Script.Run(ctx, r.RedisClient, []string{key},
			now, r.Rate, r.Capacity).Int()

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "rate limiter failed"})
			return
		}
		if result == 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
