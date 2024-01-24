package main

import (
	"log"
	"os"

	"github.com/IshanSaha05/jwt_authentication_rest_api/logger"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/database"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/middleware"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("C:\\Users\\User\\Desktop\\GoLang\\Project\\.env")

	if err != nil {
		log.Fatalf("Error: Problem while loading environment variables. \n\t Error: %s", err)
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	gin.DefaultWriter = logger.Log.Writer()

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(middleware.ExternalRateLimiter())

	routes.AuthRoutes(router)
	routes.NotesRoutes(router)

	logger.Log.Printf("Message: Running the server at port: %s", port)
	router.Run(":" + port)

	defer func() {
		database.MongoObject.Client.Disconnect(database.MongoObject.Ctx)
		database.MongoObject.Cancel()
	}()
}
