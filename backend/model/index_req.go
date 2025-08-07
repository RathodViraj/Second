package model

type IndexRequest struct {
	DocID string `json:"doc_id" binding:"required"`
	Text  string `json:"text" binding:"required"`
}
