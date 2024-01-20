package helper

import (
	"fmt"
	"os"
	"time"

	"github.com/IshanSaha05/jwt_authentication_rest_api/logger"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
)

func ValidateToken(clientToken string) (*models.SignedDetails, error) {
	// Parse the token with claims.
	err := godotenv.Load("C:\\Users\\User\\Desktop\\GoLang\\Project\\.env")
	if err != nil {
		logger.Log.Printf("Error: Problem while loading environment variables.")
		return nil, err
	}
	secretKey := os.Getenv("SECRET_KEY")

	// Parse the token.
	token, err := jwt.ParseWithClaims(clientToken, &models.SignedDetails{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		logger.Log.Printf("Error: Problem while parsing token.")
		return nil, err
	}

	// Check token validation --> if not valid return error.
	if !token.Valid {
		logger.Log.Printf("Error: Invalid token.")
		return nil, fmt.Errorf("invalid token is passed")
	}

	// Extract the claims.
	claims, ok := token.Claims.(*models.SignedDetails)
	if !ok {
		logger.Log.Printf("Error: Problem while extracting claims from the token.")
		return nil, fmt.Errorf("problem while trying to extract claims from the token")
	}

	// If token has expired, send error.
	if claims.ExpiresAt < time.Now().Local().Unix() {
		logger.Log.Printf("Error: Passed user token has expired.")
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}
