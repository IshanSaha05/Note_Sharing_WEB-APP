package main

import (
	"log"
	"os"

	"github.com/IshanSaha05/jwt_authentication_rest_api/logger"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")

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

	routes.AuthRoutes(router)
	routes.NotesRoutes(router)

	router.Run(":" + port)
}
