package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/IshanSaha05/jwt_authentication_rest_api/logger"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	userLimiter = make(map[string]*user)
	mu          sync.Mutex
)

type user struct {
	id      string
	limiter *rate.Limiter
}

func getUserLimiter(userId string) *rate.Limiter {
	// Lock the mutex
	mu.Lock()
	defer mu.Unlock()

	// Check whether the rate limiter for the user already exists or not.
	// If not make a new one.
	userObject, ok := userLimiter[userId]
	if !ok {
		userObject = &user{
			id:      userId,
			limiter: rate.NewLimiter(rate.Every(time.Second), 5),
		}

		userLimiter[userId] = userObject
	}

	return userObject.limiter
}

func InternalRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIdAny, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error: No user id provided during authentication."})
			c.Abort()
			logger.Log.Println("Error: No user id provided during authentication.")

			return
		}

		userId, ok := userIdAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while converting user id from any to string."})
			c.Abort()
			logger.Log.Println("Error: Problem while converting user id from any to string.")

			return
		}

		userLimiter := getUserLimiter(userId)

		if !userLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": fmt.Sprintf("Error: Too many requests made by user id: %s.", userId)})
			c.Abort()
			logger.Log.Printf("Error: Too many requests made by user id: %s.", userId)

			return
		}

		c.Next()
	}
}
