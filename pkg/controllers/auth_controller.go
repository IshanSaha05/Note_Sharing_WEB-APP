package controllers

import (
	"context"
	"errors"
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
		// Make a context for the request.
		ctxRequest, cancel := context.WithTimeout(c.Request.Context(), time.Second*100)
		defer cancel()

		// Bind the json into a userdata variable.
		var userClient models.UserDataClient
		err := c.BindJSON(&userClient)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while binding the json.\n\tError: %s", err.Error())
		}

		// Validate the data.
		validate := validator.New()
		err = validate.Struct(&userClient)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while validating data.\n\tError: %s", err.Error())
		}

		// Find if there is any user with the same email id presenet in the user collection already.
		mongoObject := database.DBInstance(ctxRequest, 20)
		defer mongoObject.Client.Disconnect(mongoObject.Ctx)

		userCollection, err := mongoObject.GetUserCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem with opening the user collection.\n\tError: %s", err.Error())
		}

		count, err := userCollection.CountDocuments(mongoObject.Ctx, bson.M{"email": userClient.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while finding whether the same user already exists in the database or not.\n\tError: %s", err.Error())
		}

		// If yes send a bad request message.
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: User already exists.")
		}

		// Otherwise, hash the password.
		hashedPassword := helper.HashPassword(userClient.Password)

		// Create the tokens.
		token, refreshToken, err := helper.GenerateAllToken(*userClient.Email, *userClient.First_Name, *userClient.Last_Name, userClient.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while creating tokens for the new user.\n\tError: %s", err.Error())
		}

		createdAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			logger.Log.Fatalf("Error: Problem while storing the creation stime tamp for the user.\n\tError: %s", err.Error())
		}

		updatedAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			logger.Log.Fatalf("Error: Problem while storing the creation stime tamp for the user.\n\tError: %s", err.Error())
		}

		// Fill fields for user data client side.
		userClient.ID = primitive.NewObjectID()
		userClient.Token = &token
		userClient.Refresh_Token = &refreshToken
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
		userServer.UserID = userClient.UserID

		// Save the user data struct inside the user data collection.
		_, err = userCollection.InsertOne(mongoObject.Ctx, userServer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Pronlem while storing the new user in the database.\n\tError: %s", err.Error())
		}

		// Send the respective user data struct with all the fields in the response and the correct response code.
		c.JSON(http.StatusOK, userClient)
	}
}

// POST /api/auth/login: log in to an existing user account and receive an access token.

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a request context.
		ctxRequest, cancel := context.WithTimeout(c.Request.Context(), time.Second*100)
		defer cancel()

		// Extract login info from the user.
		var user models.UserDataClient
		err := c.BindJSON(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			logger.Log.Fatalf("Error: Problem while binding the json.\n\tError: %s", err.Error())
		}

		// Create a database context.
		mongoObject := database.DBInstance(ctxRequest, 20)
		defer mongoObject.Client.Disconnect(mongoObject.Ctx)

		userCollection, err := mongoObject.GetUserCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while opening the user collection.\n\tError: %s", err.Error())
		}

		// Check whether the user exist in the user collection.
		findErr := userCollection.FindOne(mongoObject.Ctx, bson.M{"email": user.Email})

		// If no, send bad request.
		if findErr.Err() == mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Email or password is wrong.\n\tError: %s", err.Error())
		}

		// If yes, verify whether the password matches with the given password.
		var foundUser models.UserDataServer

		err = findErr.Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while decoding the found user.\n\tError: %s", err.Error())
		}

		boolVal, err := helper.VerifyPassword(*user.Password, *foundUser.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while verifying password.\n\tError: %s", err.Error())
		}

		// If password does not match, send bad request.
		if !boolVal {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Email or password is wrong.\n\tError: %s", err.Error())
		}

		//If password matches, validate the token.
		err = helper.ValidateToken(*user.Token, *user.Refresh_Token)
		if err != nil {
			if !errors.Is(err, errors.New("token has expired")) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				logger.Log.Fatalf("Error: Problem while validating user token.\n\tError: %s", err.Error())
			} else {
				token, refreshToken, err := helper.UpdateAllToken(*user.Email, *user.First_Name, *user.Last_Name, user.UserID, *user.Refresh_Token)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					logger.Log.Fatalf("Error: Problem while updating token.\n\tError: %s", err.Error())
				}

				user.Token = &token
				user.Refresh_Token = &refreshToken
			}
		}

		c.JSON(http.StatusOK, user)
	}
}
