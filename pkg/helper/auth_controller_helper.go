package helper

import (
	"fmt"
	"os"
	"time"

	"github.com/IshanSaha05/jwt_authentication_rest_api/logger"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password *string) string {
	bytesHashPassword, err := bcrypt.GenerateFromPassword([]byte(*password), 14)
	if err != nil {
		logger.Log.Printf("Error: Probelem while hashing password.\n\tError: %s", err.Error())
	}

	return string(bytesHashPassword)
}

/**
In update function pass the claims and the refresh token, then set in the new expiry date and then the refresh token to make the new access token.
Also, check when the refresh token is expired what is done.
Maybe you need to create a new refresh token and then the access token.
**/

func UpdateAllToken(email string, firstName string, lastName string, userId string, refreshTokenString string) (string, string, error) {
	// Load the environment file to get the secret key.
	err := godotenv.Load(".env")
	if err != nil {
		logger.Log.Printf("Error: Problem while loading environment variables.")
		return "", "", err
	}
	secretKey := os.Getenv("SECRET_KEY")

	// Parse the refresh token.
	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, &models.SignedDetails{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		logger.Log.Printf("Error: Problem while parsing token.")
		return "", "", err
	}

	// Check before updating the access token that the refresh token is valid.
	if !refreshToken.Valid {

		// If refresh token is not valid update the refresh token.
		refreshClaims := &models.SignedDetails{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Local().Add(time.Duration(168) * time.Hour).Unix(),
			},
		}

		refreshTokenString, err = jwt.NewWithClaims(jwt.SigningMethodES256, refreshClaims).SignedString([]byte(secretKey))
		if err != nil {
			logger.Log.Fatalf("Error: Problem while renewing refresh token for the user.")
			return "", "", err
		}
	}

	// Now update the access token.
	claims := &models.SignedDetails{
		Email:         email,
		First_Name:    firstName,
		Last_Name:     lastName,
		User_Id:       userId,
		Refresh_Token: refreshTokenString,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Duration(24) * time.Hour).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodES256, claims).SignedString([]byte(secretKey))
	if err != nil {
		logger.Log.Fatalf("Error: Problem while renewing access token for the user.")
		return "", "", err
	}

	return token, refreshTokenString, nil
}

func GenerateAllToken(email string, firstName string, lastName string, userId string) (string, string, error) {
	err := godotenv.Load(".env")
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

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodES256, refreshClaims).SignedString([]byte(secretKey))
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

	token, err := jwt.NewWithClaims(jwt.SigningMethodES256, claims).SignedString([]byte(secretKey))
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

func ValidateToken(tokenString string, refreshTokenString string) error {
	// Parse the token with claims.
	err := godotenv.Load(".env")
	if err != nil {
		logger.Log.Printf("Error: Problem while loading environment variables.")
		return err
	}
	secretKey := os.Getenv("SECRET_KEY")

	token, err := jwt.ParseWithClaims(tokenString, &models.SignedDetails{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		logger.Log.Printf("Error: Problem while parsing token.")
		return err
	}

	// Check token validation if not return error or if expired refresh the token.
	if !token.Valid {
		// If the token has expired, send an error to update the token using the refresh token.
		claims, ok := token.Claims.(*models.SignedDetails)
		if !ok {
			logger.Log.Printf("Error: Problem while extracting claims from the token.")
			return fmt.Errorf("problem while trying to extract claims from the token")
		}

		if claims.ExpiresAt < time.Now().Local().Unix() {
			logger.Log.Printf("Error: Passed user token has expired.")
			return fmt.Errorf("token has expired")
		} else {
			logger.Log.Printf("Error: Invalid token.")
			return fmt.Errorf("invalid token is passed")
		}
	}

	// Send error as nil
	return nil
}
