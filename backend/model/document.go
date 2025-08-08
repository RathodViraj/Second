package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Document struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title   string             `bson:"title" json:"title"`
	Content string             `bson:"content" json:"content"`
	Tags    []string           `bson:"tags" json:"tags"`
	Created time.Time          `bson:"created" json:"created"`
}
