package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/IshanSaha05/jwt_authentication_rest_api/logger"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/database"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/helper"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// POST /api/auth/signup: create a new user account.

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bind the json into a userdata variable.
		var userClient models.UserDataClient
		err := c.BindJSON(&userClient)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Printf("\nError: Problem while binding the json.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Validate the data.
		validate := validator.New()
		err = validate.Struct(&userClient)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Printf("\nError: Problem while validating data.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Get the user collection.
		userCollection, err := database.MongoObject.GetUserCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Printf("\nError: Problem with opening the user collection.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Find whether user with the same email address exists in the collection already or not.
		var foundDocument models.UserDataServer
		filter := bson.D{{Key: "email", Value: userClient.Email}}

		err = userCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&foundDocument)

		// If there is a document, error will be given.
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: User with same email id: %s, already exists.", *userClient.Email)})
			logger.Log.Printf("\nError: User with same email id: %s, already exists.", *userClient.Email)
			c.Abort()
			return
		}

		// If there is an error and it is not equivalent to errnodocuments then it is a decoding error.
		if err != nil && err != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while decoding document.\n\tError: %s", err.Error())})
			logger.Log.Printf("\nError: Problem while decoding document.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// There is an error from the previous step but it is due to no document found and thus no document for decoding, hence, next step is to hash the password.
		hashedPassword := helper.HashPassword(userClient.Password)

		// Create the tokens.
		token, refreshToken, err := helper.GenerateAllToken(*userClient.Email, *userClient.First_Name, *userClient.Last_Name, userClient.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Printf("\nError: Problem while creating tokens for the new user.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		createdAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			logger.Log.Printf("\nError: Problem while storing the creation time tamp for the user.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		updatedAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			logger.Log.Printf("\nError: Problem while storing the update time tamp for the user.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		lastLogin, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			logger.Log.Printf("\nError: Problem while storing the lost login time tamp for the user.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Fill fields for user data client side.
		userClient.ID = primitive.NewObjectID()
		userClient.Token = &token
		userClient.UserID = userClient.ID.Hex()

		// Fill fields for user data server side.
		var userServer models.UserDataServer
		userServer.ID = userClient.ID
		userServer.First_Name = userClient.First_Name
		userServer.Last_Name = userClient.Last_Name
		userServer.Password = &hashedPassword
		userServer.Email = userClient.Email
		userServer.Created_At = createdAt
		userServer.Updated_At = updatedAt
		userServer.Last_Login = lastLogin
		userServer.Refresh_Token = &refreshToken
		userServer.UserID = userClient.UserID

		// Save the user data struct inside the user data collection.
		_, err = userCollection.InsertOne(database.MongoObject.Ctx, userServer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Printf("\nError: Pronlem while storing the new user in the database.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Send the respective user data struct with all the fields in the response and the correct response code.
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Message: Successful signing up of user with user id: %s and user email id: %s", userClient.UserID, *userClient.Email), "data": userClient})
		logger.Log.Printf("\nMessage: Successful signing up of user with user id: %s and user email id: %s", userClient.UserID, *userClient.Email)
	}
}

// POST /api/auth/login: log in to an existing user account and receive an access token.
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract login info from the user.
		var user models.UserDataClient
		err := c.BindJSON(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			logger.Log.Printf("Error: Problem while binding the json.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Get the user collection.
		userCollection, err := database.MongoObject.GetUserCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Printf("Error: Problem while opening the user collection.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Check whether the user exist in the user collection.
		var foundUser models.UserDataServer
		filter := bson.D{{Key: "email", Value: user.Email}}

		err = userCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&foundUser)

		// If there is an error due to decoding or no document, error will be given.
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: No user with email id: %s registered.", *user.Email)})
				logger.Log.Printf("Error: No user with email id: %s registered.", *user.Email)
				c.Abort()
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while decoding found user.\n\tError: %s", err.Error())})
				logger.Log.Printf("Error: Problem while decoding found user.\n\tError: %s", err.Error())
				c.Abort()
				return
			}
		}

		// Document has been found so match the password.
		boolVal, err := helper.VerifyPassword(*user.Password, *foundUser.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Printf("Error: Problem while verifying password.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// If password does not match, send bad request.
		if !boolVal {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			logger.Log.Printf("Error: Email or password is wrong.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// If password matches, generate all tokens.
		token, refreshToken, err := helper.GenerateAllToken(*foundUser.Email, *foundUser.First_Name, *foundUser.Last_Name, foundUser.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Printf("\nError: Problem while generating tokens to update for the existing user.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Update the last login date and refresh token in the server side user database.
		lastLogin, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			logger.Log.Printf("Error: Problem while trying to update the last login date.\n\tError: %s", err.Error())
			c.Abort()
			return
		}
		err = helper.UpdateLastLoginAndRefreshToken(foundUser.UserID, lastLogin, refreshToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Printf("Error: Problem while updating the last login date for the user.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Update the token and send it to the user along with status ok.
		user.ID = foundUser.ID
		user.First_Name = foundUser.First_Name
		user.Last_Name = foundUser.Last_Name
		user.Token = &token
		user.UserID = foundUser.UserID

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Message: Successful logging up of user with user id: %s and user email id: %s", user.UserID, *user.Email), "data": user})
		logger.Log.Printf("Message: Successful logging up of user with user id: %s and user email id: %s", user.UserID, *user.Email)
	}
}
