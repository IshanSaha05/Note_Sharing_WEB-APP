package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBObject struct {
	Client *mongo.Client
	Ctx    context.Context
}

func DBInstance(parentContext context.Context, timeoutValue int) *MongoDBObject {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error: Problem while loading the environment variables.\n\tError: %s", err)
	}

	MongoDB_URI := os.Getenv("MONGODB_URI")

	ctx, cancel := context.WithTimeout(parentContext, time.Second*time.Duration(timeoutValue))
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoDB_URI))
	if err != nil {
		log.Fatalf("Error: Problem while connecting to the database.\n\tError: %s", err)
	}

	return &MongoDBObject{
		Client: client,
		Ctx:    ctx,
	}
}
