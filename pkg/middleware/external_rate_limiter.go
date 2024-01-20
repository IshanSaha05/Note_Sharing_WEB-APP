package middleware

import (
	"net/http"
	"time"

	"github.com/IshanSaha05/jwt_authentication_rest_api/logger"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var globalRateLimiter *rate.Limiter

func init() {
	globalRateLimiter = rate.NewLimiter(rate.Every(time.Second), 5)
}

func ExternalRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !globalRateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"Error": "Error: Too many requests made."})
			c.Abort()
			logger.Log.Println("Error: Too many requests made.")
			return
		}

		c.Next()
	}
}
