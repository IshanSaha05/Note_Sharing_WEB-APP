package database

import (
	"os"

	"github.com/IshanSaha05/jwt_authentication_rest_api/logger"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

func getDatabase(mongoObject *MongoDBObject) (*mongo.Database, error) {
	err := godotenv.Load(".env")
	if err != nil {
		logger.Log.Println("Error: Problem while loading environment variables.")
		return nil, err
	}

	databaseName := os.Getenv("MONGODB_DATABASE_NAME")

	return mongoObject.Client.Database(databaseName), nil
}

func (mongoObject *MongoDBObject) GetUserCollection() (*mongo.Collection, error) {
	err := godotenv.Load(".env")
	if err != nil {
		logger.Log.Println("Error: Problem while loading environment variables.")
		return nil, err
	}

	database, err := getDatabase(mongoObject)
	if err != nil {
		logger.Log.Println("Error: Problem while loading the database.")
		return nil, err
	}

	userCollectionName := os.Getenv("USERS_COLLECTION")

	return database.Collection(userCollectionName), nil
}

func (mongoObject *MongoDBObject) GetNoteCollection() (*mongo.Collection, error) {
	err := godotenv.Load(".env")
	if err != nil {
		logger.Log.Println("Error: Problem while loading environment variables.")
		return nil, err
	}

	database, err := getDatabase(mongoObject)
	if err != nil {
		logger.Log.Println("Error: Problem while loading the database.")
		return nil, err
	}

	noteCollectionName := os.Getenv("NOTES_COLLECTION")

	return database.Collection(noteCollectionName), nil
}
