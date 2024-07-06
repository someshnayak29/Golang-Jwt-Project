package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// URL => Client => mongoDB connect => client object => create collection

// returns mongo client
func DBinstance() *mongo.Client {

	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Retrieve MongoDB connection string from environment variables
	MongoDb := os.Getenv("MONGODB_URL") // Retrieves the MongoDB connection string from environment variables loaded earlier.

	// Set up MongoDB client options
	clientOptions := options.Client().ApplyURI(MongoDb)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!!!")

	return client
}

var Client *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {

	var collection *mongo.Collection = client.Database("cluster0").Collection(collectionName)
	return collection
}
