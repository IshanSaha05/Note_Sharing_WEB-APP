package useless

/*
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
*/

/*
func Test() gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.Param("email")

		filter := bson.M{"email": email}

		userCollection, _ := database.MongoObject.GetUserCollection()

		var found models.UserDataServer
		userCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&found)

		c.JSON(http.StatusOK, found)
	}
}
*/
