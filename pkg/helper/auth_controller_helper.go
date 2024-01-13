package helper

import (
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

func GenerateAllToken(email string, firstName string, lastName string, userId string) (string, string, error) {
	claims := &models.SignedDetails{
		Email:      email,
		First_Name: firstName,
		Last_Name:  lastName,
		User_Id:    userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Duration(24) * time.Hour).Unix(),
		},
	}

	refreshClaims := &models.SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Duration(24) * time.Hour).Unix(),
		},
	}

	err := godotenv.Load(".env")
	if err != nil {
		logger.Log.Printf("Error: Problem while loading environment variables.")
		return "", "", err
	}
	secretKey := os.Getenv("SECRET_KEY")

	token, err := jwt.NewWithClaims(jwt.SigningMethodES256, claims).SignedString([]byte(secretKey))
	if err != nil {
		logger.Log.Printf("Error: Problem while creating token for the user.")
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodES256, refreshClaims).SignedString([]byte(secretKey))
	if err != nil {
		logger.Log.Printf("Error: Problem while creating refresh token for the user.")
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

func ValidateToken(user *models.UserData, foundUser *models.UserData) error {
	// Validate the token in the user is same as token in the found user.

	// If not return error.

	// If the token has expired, update the token using the refresh token.

	// Send error as nil
}
