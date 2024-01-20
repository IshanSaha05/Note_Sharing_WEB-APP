package routes

import (
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/controllers"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/middleware"
	"github.com/gin-gonic/gin"
)

/**
Add in a middleware for authentication purpose.

Create routes for doing various functions with the notes collection.

	Note Endpoints

	GET /api/notes: get a list of all notes for the authenticated user.
	GET /api/notes/:id: get a note by ID for the authenticated user.
	POST /api/notes: create a new note for the authenticated user.
	PUT /api/notes/:id: update an existing note by ID for the authenticated user.
	DELETE /api/notes/:id: delete a note by ID for the authenticated user.
	POST /api/notes/:id/share: share a note with another user for the authenticated user.
	GET /api/search?q=:query: search for notes based on keywords for the authenticated user.
**/

func NotesRoutes(incomingRoutes *gin.Engine) {
	// Remove the return statement and write your code.
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.Use(middleware.InternalRateLimiter())
	incomingRoutes.GET("/api/notes", controllers.GetAllNotes())
	incomingRoutes.GET("/api/notes/:id", controllers.GetNotesByID())
	incomingRoutes.POST("/api/notes", controllers.CreateNotes())
	incomingRoutes.PUT("/api/notes/:id", controllers.UpdateNotesByID())
	incomingRoutes.DELETE("/api/notes/:id", controllers.DeleteNotesByID())
	incomingRoutes.POST("/api/notes/:id/share", controllers.ShareNotesByID())
	incomingRoutes.GET("/api/search?q=:query", controllers.SearchNotesByKeywords())
}
