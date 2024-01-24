package routes

import (
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/controllers"
	"github.com/gin-gonic/gin"
)

/**
Create routes for signup and login.

	Authentication Endpoints

	POST /api/auth/signup: create a new user account.
	POST /api/auth/login: log in to an existing user account and receive an access token.
**/

func AuthRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/api/auth/signup", controllers.SignUp())
	incomingRoutes.POST("/api/auth/login", controllers.Login())
}
