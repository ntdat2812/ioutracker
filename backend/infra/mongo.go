package infra

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoDB *mongo.Client
var DB string

func ConnectToMongoDB() {

	user := os.Getenv("MONGO_USER")
	pwd := os.Getenv("MONGO_PASSWORD")
	uri := os.Getenv("MONGO_URI")
	DB = os.Getenv("MONGO_DB")

	mongoURI := fmt.Sprintf(uri, user, pwd)

	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1))

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Fatalf("Error connecting mongodb %v", err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Send a ping to confirm a successful connection
	if err = client.Ping(ctx, nil); err != nil {
		panic(err)
	}

	MongoDB = client

	fmt.Printf("Pinged your deployment. You successfully connected to %v!\n", mongoURI)
}

func DisconnectMongoDB() {
	err := MongoDB.Disconnect(context.TODO())
	if err != nil {
		log.Fatalf("Error disconnecting mongodb %v", err.Error())
	}
}
