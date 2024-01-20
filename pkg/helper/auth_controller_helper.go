package helper

import (
	"context"
	"os"
	"time"

	"github.com/IshanSaha05/jwt_authentication_rest_api/logger"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/database"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password *string) string {
	bytesHashPassword, err := bcrypt.GenerateFromPassword([]byte(*password), 14)
	if err != nil {
		logger.Log.Printf("Error: Probelem while hashing password.\n\tError: %s", err.Error())
	}

	return string(bytesHashPassword)
}

func GenerateAllToken(email string, firstName string, lastName string, userId string) (string, string, error) {
	err := godotenv.Load("C:\\Users\\User\\Desktop\\GoLang\\Project\\.env")
	if err != nil {
		logger.Log.Printf("Error: Problem while loading environment variables.")
		return "", "", err
	}
	secretKey := os.Getenv("SECRET_KEY")

	refreshClaims := &models.SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Duration(168) * time.Hour).Unix(),
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(secretKey))
	if err != nil {
		logger.Log.Printf("Error: Problem while creating refresh token for the user.")
		return "", "", err
	}

	claims := &models.SignedDetails{
		Email:         email,
		First_Name:    firstName,
		Last_Name:     lastName,
		User_Id:       userId,
		Refresh_Token: refreshToken,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Duration(24) * time.Hour).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))
	if err != nil {
		logger.Log.Printf("Error: Problem while creating token for the user.")
		return "", "", err
	}

	return token, refreshToken, nil
}

func VerifyPassword(userPassword string, foundUserPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(foundUserPassword), []byte(userPassword))
	if err != nil {
		logger.Log.Printf("Error: Problem while comparing passwords.")
		return false, err
	}

	return true, nil
}

func UpdateLastLoginAndRefreshToken(userId string, lastLogin time.Time, refreshToken string) error {
	// Get the access to the user collection.
	userCollection, err := database.MongoObject.GetUserCollection()
	if err != nil {
		logger.Log.Println("Error: Problem with opening the user collection.")
		return err
	}

	// Create and fill the update object.
	var updateMiniObj primitive.D

	updateMiniObj = append(updateMiniObj, bson.E{Key: "lastLogin", Value: lastLogin})
	updateMiniObj = append(updateMiniObj, bson.E{Key: "refreshToken", Value: refreshToken})

	updateObj := primitive.D{
		{
			Key: "$set", Value: updateMiniObj,
		},
	}

	filter := bson.M{"userId": userId}

	options := options.Update().SetUpsert(true)

	// Create an update context.
	ctxUpdate, cancel := context.WithTimeout(database.MongoObject.Ctx, time.Second*10)
	defer cancel()

	_, err = userCollection.UpdateOne(ctxUpdate, filter, updateObj, options)
	if err != nil {
		logger.Log.Println("Error: Problem while trying to update the field in the database.")
		return err
	}

	return nil
}
