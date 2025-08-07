package db

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	once        sync.Once
)

func ConnectMongo() *mongo.Client {
	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatal("MongoDB connection error:", err)
		}

		if err = client.Ping(ctx, nil); err != nil {
			log.Fatal("MongoDB ping error:", err)
		}

		mongoClient = client
		log.Println("MongoDB connected.")
	})

	return mongoClient
}

func GetCollection(database, collection string) *mongo.Collection {
	client := ConnectMongo()
	return client.Database(database).Collection(collection)
}
