package handler

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type TypeaheadWS struct {
	Redis *redis.Client
	Ctx   context.Context
}

type TypeaheadRequest struct {
	Prefix string `json:"prefix"`
	Limit  int    `json:"limit"`
}

type TypeaheadResponse struct {
	Suggestions []string `json:"suggestions"`
}

func NewTypeaheadWS(r *redis.Client) *TypeaheadWS {
	return &TypeaheadWS{
		Redis: r,
		Ctx:   context.Background(),
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // or add your origin check
	},
}

func (ws *TypeaheadWS) Handler(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		var req TypeaheadRequest
		if err := json.Unmarshal(message, &req); err != nil {
			log.Println("Unmarshal error:", err)
			continue
		}

		if len(req.Prefix) < 3 || len(req.Prefix) > 20 {
			continue
		}
		if req.Limit <= 0 {
			req.Limit = 10
		}

		suggestions, err := ws.getSuggestions(req.Prefix, req.Limit)
		if err != nil {
			log.Println("Error getting suggestions:", err)
			continue
		}

		resp := TypeaheadResponse{Suggestions: suggestions}
		conn.WriteJSON(resp)

		if len(suggestions) > 0 {
			go ws.incrementScore(suggestions[0])
		}
	}
}

func (ws *TypeaheadWS) getSuggestions(prefix string, k int) ([]string, error) {
	suggestions, err := ws.Redis.ZRevRangeWithScores(ws.Ctx, "typeahead", 0, -1).Result()
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

func (ws *TypeaheadWS) incrementScore(term string) {
	ws.Redis.ZIncrBy(ws.Ctx, "typeahead", 1, term)
}
