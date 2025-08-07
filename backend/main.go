package main

import (
	"context"
	"log"
	"second/db"
	"second/handler"
	"second/middleware"
	"second/repository"
	"second/utils"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	err := utils.LoadStopWords("utils/stopwords.txt")
	if err != nil {
		log.Fatalf("Failed to load stop words: %v", err)
	}

	db.ConnectMongo()
	rdb := db.InitRedis()

	rl, err := middleware.NewRateLimiter(rdb, 0.5, 5)
	if err != nil {
		log.Fatalf("Failed to create rate limiter: %v", err)
	}

	trendingRepo := repository.NewTrendingRepo(rdb)
	docRepo := repository.NewDocumentRepo(rdb)

	docHandler := handler.NewDocHandler(docRepo)
	trendingHandler := handler.NewTrendingHandler(trendingRepo)

	r := gin.Default()

	// Enable CORS for your frontend (Vite runs on :5173)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Then your middleware and routes
	r.Use(rl.Middleware())

	r.GET("/search", docHandler.Search)
	r.POST("/add", docHandler.AddDocument)
	r.GET("/document/:id", docHandler.GetDocumentByID)

	r.GET("/trending", trendingHandler.GetTrendingDocs)

	wsHandler := handler.NewTypeaheadWS(rdb)
	r.GET("/typeahead", wsHandler.Handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startSlidingWindow(ctx, trendingRepo)

	r.Run(":8080")
}

func startSlidingWindow(ctx context.Context, repo *repository.TrendingRepo) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := repo.SlideWindow(); err != nil {
					log.Printf("SlideWindow error: %v\n", err)
				}
				waitUntilNextMinute()
			}
		}
	}()
}
func waitUntilNextMinute() {
	now := time.Now()
	next := now.Truncate(time.Minute).Add(time.Minute)
	time.Sleep(time.Until(next))
}
