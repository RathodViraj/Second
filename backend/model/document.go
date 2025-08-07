package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Document struct {
	ID      primitive.ObjectID `json:"_id,omitempty"`
	Title   string             `json:"title"`
	Content string             `json:"content"`
	Tags    []string           `json:"tags"`
	Created time.Time          `json:"created"`
}
