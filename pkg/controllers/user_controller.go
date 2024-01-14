package controllers

import (
	"github.com/gin-gonic/gin"
)

// GET /api/notes: get a list of all notes for the authenticated user.

func GetAllNotes() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// GET /api/notes/:id: get a note by ID for the authenticated user.

func GetNotesByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// POST /api/notes: create a new note for the authenticated user.

func CreateNotes() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// PUT /api/notes/:id: update an existing note by ID for the authenticated user.

func UpdateNotesByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// DELETE /api/notes/:id: delete a note by ID for the authenticated user.

func DeleteNotesByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// POST /api/notes/:id/share: share a note with another user for the authenticated user.

func ShareNotesByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// GET /api/search?q=:query: search for notes based on keywords for the authenticated user.

func SearchNotesByKeywords() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
