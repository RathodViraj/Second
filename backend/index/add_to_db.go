package index

import (
	"context"
	"errors"
	"log"
	"second/model"
	"time"

	"second/db"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AddDocumentService struct {
}

func NewAddDocumentService() *AddDocumentService {
	return &AddDocumentService{}
}

func (s *AddDocumentService) AddDocument(doc model.Document) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetCollection("secondDB", "documents")

	if doc.Created.IsZero() {
		doc.Created = time.Now()
	}

	log.Printf("Adding document with ID: %s", doc.ID.Hex())
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
