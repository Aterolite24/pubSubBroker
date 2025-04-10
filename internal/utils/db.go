package utils

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	clientInstance    *mongo.Collection
	clientInstanceErr error
	mongoOnce         sync.Once
	connectionString  string
	dbName            string
	collectionName    string
)

func init() {

	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Could not load .env file: %v. Relying on environment variables.\n", err)
	}

	connectionString = os.Getenv("MONGO_CONNECTION_STRING")
	dbName = os.Getenv("MONGO_DB_NAME")
	collectionName = os.Getenv("MONGO_COLLECTION_NAME")

	if connectionString == "" {
		log.Fatal("Error: MONGO_CONNECTION_STRING environment variable not set.")
	}
	if dbName == "" {
		log.Fatal("Error: MONGO_DB_NAME environment variable not set.")
	}
	if collectionName == "" {
		log.Fatal("Error: MONGO_COLLECTION_NAME environment variable not set.")
	}

}

func GetMongoClient() (*mongo.Collection, error) {
	mongoOnce.Do(func() {
		clientOpts := options.Client().ApplyURI(connectionString)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(options.Client(), clientOpts)
		if err != nil {
			log.Fatal("MongoDB connection error:", err)
		}

		err = client.Ping(ctx, nil)
		if err != nil {
			clientInstanceErr = err
			return
		}

		clientInstance = client.Database(dbName).Collection(collectionName)
	})

	return clientInstance, clientInstanceErr
}

func DeleteRowsOlderThan(duration time.Duration) error {
	client, err := GetMongoClient()
	if err != nil {
		return err
	}

	cutoffTime := time.Now().Add(-duration)
	filter := bson.M{"timestamp": bson.M{"$lt": cutoffTime}}

	deleteResult, err := client.DeleteMany(context.Background(), filter)
	if err != nil {
		return err
	}

	log.Printf("Deleted %d rows older than %v\n", deleteResult.DeletedCount, duration)
	return nil
}

func StartRowDeletionJob(interval time.Duration, duration time.Duration) {
	go func() {
		for {
			err := DeleteRowsOlderThan(duration)
			if err != nil {
				log.Println("Error deleting rows:", err)
			}
			time.Sleep(interval)
		}
	}()
}
