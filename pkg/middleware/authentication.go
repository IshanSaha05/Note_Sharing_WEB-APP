package middleware

import (
	"net/http"

	"github.com/IshanSaha05/jwt_authentication_rest_api/logger"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/helper"
	"github.com/gin-gonic/gin"
)

// Write the authenticate function.

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: No token provided, authentication cannot be done."})
			c.Abort()
			logger.Log.Println("Error: No token provided, authentication cannot be done.")
			return
		}

		claims, err := helper.ValidateToken(clientToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("firstName", claims.First_Name)
		c.Set("lastName", claims.Last_Name)
		c.Set("userID", claims.User_Id)
		c.Set("refreshToken", claims.Refresh_Token)

		c.Next()
	}
}
