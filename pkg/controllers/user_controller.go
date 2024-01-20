package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/IshanSaha05/jwt_authentication_rest_api/logger"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/database"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GET /api/notes: get a list of all notes for the authenticated user.

func GetAllNotes() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the authenticated user id.
		userIdAny, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error: No user id present."})
			logger.Log.Fatal("Error: No user id present.")
		}

		userId, ok := userIdAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error:": "Error: Problem while converting user id to string."})
			logger.Log.Fatal("Error: Problem while converting user if to string.")
		}

		// Get the note collection.
		noteCollection, err := database.MongoObject.GetNoteCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while getting the note collection.\n\tError: %s", err.Error())})
			logger.Log.Fatalf("Error: Problem while getting the note collection.\n\tError: %s", err)
		}

		// Create a filter.
		filter := bson.D{}

		// Create a cursor for the whole document.
		cursor, err := noteCollection.Find(database.MongoObject.Ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while creating cursor for the whole collection."})
			logger.Log.Fatalf("Error: Problem while creating cursor for the whole collection.")
		}
		defer cursor.Close(database.MongoObject.Ctx)

		// Declare the list which will contain all the documents.
		var foundDocuments []models.NoteData

		// Iterate through each document.
		for cursor.Next(database.MongoObject.Ctx) {
			// Decode the document.
			var foundDocument models.NoteData

			if err := cursor.Decode(&foundDocument); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while decoding document.\n\tError: %s", err.Error())})
				logger.Log.Fatalf("Error: Problem while decoding document.\n\tError: %s", err.Error())
			}

			if *foundDocument.User_Id == userId || *foundDocument.Sharable {
				foundDocuments = append(foundDocuments, foundDocument)
			}
		}

		// Check for any error during cursor iteration.
		if err := cursor.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while iterating through the collection using cursor.\n\tError: %s", err.Error())})
			logger.Log.Fatalf("Error: Problem while iterating through the collection using cursor.\n\tError: %s", err.Error())
		}

		// Send a response ok with the list of documents.
		c.JSON(http.StatusOK, foundDocuments)
		logger.Log.Println("Message: Successfully responded with the list of all notes for the authenticated user.")
	}
}

// GET /api/notes/:id: get a note by ID for the authenticated user.

func GetNotesByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the note id from the url.
		notesId := c.Param("id")

		// Get the authenticated user id.
		userIdAny, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error: No user id present."})
			logger.Log.Fatal("Error: No user id present.")
		}

		userId, ok := userIdAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error:": "Error: Problem while converting user id to string."})
			logger.Log.Fatal("Error: Problem while converting user if to string.")
		}

		// Get the note collection.
		noteCollection, err := database.MongoObject.GetNoteCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while getting the note collection.\n\tError: %s", err.Error())})
			logger.Log.Fatalf("Error: Problem while getting the note collection.\n\tError: %s", err)
		}

		// Make the filter for search --> use the note id.
		filter := bson.M{"_id": notesId}

		// Decode and bind it in a note user.
		var note models.NoteData

		err = noteCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&note)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: No such note present./Problem while decoding the note.\n\tError: %s", err.Error())})
			logger.Log.Fatalf("Error: No such note present./Problem while decoding the note.\n\tError: %s", err.Error())
		}

		// Check whether user id in the decoded version is same as the authenticated user id.
		// If yes, then send status ok and send the data with header field.
		// If no, then check if sharable is true.
		// If yes, then send status ok and send the data with header field.
		// Otherwise, send bad status.
		if userId == *note.User_Id {
			c.JSON(http.StatusOK, note)
			logger.Log.Printf("Message: Note belongs to user with user id: %s and present in the db with notes id: %s. It is successfully shared.", userId, notesId)
		} else {
			if *note.Sharable {
				c.JSON(http.StatusOK, note)
				logger.Log.Printf("Message: Note does not belong to user with user with user id: %s, but sharable and thus successfully shared notes with notes id: %s", userId, notesId)
			} else {
				c.JSON(http.StatusBadRequest, nil)
				logger.Log.Fatalf("Message: Note does not belong to user with user id: %s and it is also not sharable publically. Hence, it cannot be shared.", userId)
			}
		}
	}
}

// POST /api/notes: create a new note for the authenticated user.

func CreateNotes() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bind the json and extract the data into the notes struct variable.
		var note models.NoteData
		err := c.BindJSON(&note)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while binding the data to the note struct.\n\tError: %s", err.Error())
		}

		// Validate the data.
		validate := validator.New()

		err = validate.Struct(&note)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while validating the data.\n\tError: %s", err.Error())
		}

		// Get the notes collection.
		notesCollection, err := database.MongoObject.GetNoteCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while trying to open notes collection.\n\tError: %s", err.Error())
		}

		// Update the values of the field which are not given by the user for the note struct variable.
		userIdAny, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: The user-id is not passed from the middleware the create handler through context."})
			logger.Log.Fatal("Error: The user-id is not passed from the middleware the create handler through context.")
		}
		userId, ok := userIdAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while converting any to string."})
			logger.Log.Fatal("Error: Problem while converting any to string.")
		}

		userEmailAny, exists := c.Get("email")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: The user-id is not passed from the middleware the create handler through context."})
			logger.Log.Fatal("Error: The user-id is not passed from the middleware the create handler through context.")
		}
		userEmail, ok := userEmailAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while converting any to string."})
			logger.Log.Fatal("Error: Problem while converting any to string.")
		}

		note.Email = &userEmail

		uniqueHeaderString := fmt.Sprintf("%s%s", *note.User_Id, *note.Header)
		note.Unique_Header = &uniqueHeaderString

		createdAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			logger.Log.Fatalf("Error: Problem while storing the creation time tamp for the user.\n\tError: %s", err.Error())
		}

		updatedAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			logger.Log.Fatalf("Error: Problem while storing the update time tamp for the user.\n\tError: %s", err.Error())
		}

		note.Created_At = createdAt
		note.Updated_At = updatedAt
		note.User_Id = &userId
		note.ID = primitive.NewObjectID()

		// Check whether the notes already exist or note.
		count, err := notesCollection.CountDocuments(database.MongoObject.Ctx, bson.M{"uniqueHeader": note.Unique_Header})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while finding whether the same user already exists in the database or not.\n\tError: %s", err.Error())
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Notes already exist with the same header. Change the header.")
		}

		// If does not exist insert the notes.
		_, err = notesCollection.InsertOne(database.MongoObject.Ctx, note)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while inserting new data into the notes collection.\n\tError: %s", err.Error())
		}

		// Get the updated object from the database and send it to the client as a response with ok status.
		var noteInserted models.NoteData

		filter := bson.M{"_id": note.ID}

		err = notesCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&noteInserted)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while trying the find the inserted document in the DB."})
			logger.Log.Fatal("Error: Problem while trying the find the inserted document in the DB.")
		}

		c.JSON(http.StatusOK, noteInserted)
		logger.Log.Printf("Message: Successful creation of new note document with note id: %s", noteInserted.ID)
	}
}

// PUT /api/notes/:id: update an existing note by ID for the authenticated user.

func UpdateNotesByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Getting the notes id from the url.
		notesId := c.Param("id")

		// Getting the user id from the authenticator.
		userIdAny, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: The user-id is not passed from the middleware the create handler through context."})
			logger.Log.Fatal("Error: The user-id is not passed from the middleware the create handler through context.")
		}
		userId, ok := userIdAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while converting any to string."})
			logger.Log.Fatal("Error: Problem while converting any to string.")
		}

		// Get the notes collection.
		notesCollection, err := database.MongoObject.GetNoteCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while trying to open notes collection.\n\tError: %s", err.Error())
		}

		// Find whether there is any notes present with the passed notes id in the url.
		filter := bson.M{"_id": notesId}

		var foundNotes models.NoteData

		err = notesCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&foundNotes)

		// If no notes present, send bad request.
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: No notes with id: %s is present in the database.", notesId)})
			logger.Log.Fatalf("Error: No notes with id: %s is present in the database.", notesId)
		}

		// If authenticator user id is not same as document user id, send status bad request.
		if *foundNotes.User_Id == userId {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: Notes is not accessible to the user with user id: %s", userId)})
			logger.Log.Fatalf("Error: Notes is not accessible to the user with user id: %s", userId)
		}

		// If same, then the user has access to the note document and can update.
		// Bind the json and extract the data into the note struct.
		var note *models.NoteData

		err = c.BindJSON(&note)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while binding the data to the note struct.\n\tError: %s", err.Error())
		}

		// Fill in the other fields in the note which needs to be changed.
		updatedAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			logger.Log.Fatalf("Error: Problem while storing the update time tamp for the user.\n\tError: %s", err.Error())
		}

		note.Updated_At = updatedAt

		// Make the filter to update the data.
		filter = bson.M{"_id": notesId}

		// Make the update object.
		var updateObj primitive.D

		if note.Header != nil {
			updateObj = append(updateObj, bson.E{Key: "header", Value: note.Header})

			uniqueHeader := fmt.Sprintf("%s%s", userId, *note.Header)
			updateObj = append(updateObj, bson.E{Key: "uniqueHeader", Value: uniqueHeader})
		}

		if note.Data != nil {
			updateObj = append(updateObj, bson.E{Key: "notesData", Value: note.Data})
		}

		if note.Sharable != nil {
			updateObj = append(updateObj, bson.E{Key: "sharable", Value: note.Sharable})
		}

		updateObj = append(updateObj, bson.E{Key: "updateAt", Value: note.Updated_At})

		// Make the options for the find and update operation.
		options := options.FindOneAndUpdate().SetReturnDocument(options.After)

		// Make the update.
		var updateNote models.NoteData

		err = notesCollection.FindOneAndUpdate(database.MongoObject.Ctx, filter, updateObj, options).Decode(&updateNote)

		// If could not, send bad request.
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while upating data.\n\tError: %s", err.Error())})
			logger.Log.Fatalf("Error: Problem while upating data.\n\tError: %s", err.Error())
		}

		// If update successful Send status ok.
		c.JSON(http.StatusOK, updateNote)
		logger.Log.Printf("Message: Updated notes with notes id: %s successfully.", notesId)
	}
}

// DELETE /api/notes/:id: delete a note by ID for the authenticated user.

func DeleteNotesByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the notes id.
		noteId := c.Param("id")

		// Get the user id from the authentication.
		userIdAny, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error: No user id given for authentication."})
			logger.Log.Fatal("Error: No user id given for authentication.")
		}

		userId, ok := userIdAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while converting user id from any to string."})
			logger.Log.Fatal("Error: Problem while converting user id from any to string.")
		}

		// Get access to the user collection.
		noteCollection, err := database.MongoObject.GetNoteCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while getting the note collection.\n\tError: %s", err.Error())})
			logger.Log.Fatalf("Error: Problem while getting the note collection.\n\tError: %s", err.Error())
		}

		// Make the filter.
		filter := bson.M{"_id": noteId}

		// Find the object and decode it.
		var foundNote models.NoteData

		err = noteCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&foundNote)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: No such document with note id: %s found.\n\tError: %s", noteId, err.Error())})
			logger.Log.Fatalf("Error: No such document with note id: %s found.\n\tError: %s", noteId, err.Error())
		}

		// If decoded userid is equivalent to the authenticated user id delete it and send status ok with the deleted notes details.
		// If not, then send bad request.
		if *foundNote.User_Id == userId {
			_, err = noteCollection.DeleteOne(database.MongoObject.Ctx, filter)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while deleting the note with note id: %s by the user with user id: %s.\n\tError: %s", noteId, userId, err.Error())})
				logger.Log.Fatalf("Error: Problem while deleting the note with note id: %s by the user with user id: %s.\n\tError: %s", noteId, userId, err.Error())
			}

			c.JSON(http.StatusOK, foundNote)
			logger.Log.Printf("Message: Successfully deleted note with note id: %s by the user with user id: %s", noteId, userId)

		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: User with user id: %s is not allowed to delete the notes with note id: %s", userId, noteId)})
			logger.Log.Fatalf("Error: User with user id: %s is not allowed to delete the notes with note id: %s", userId, noteId)
		}
	}
}

// POST /api/notes/:id/share: share a note with another user for the authenticated user.

func ShareNotesByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the note collection.
		noteCollection, err := database.MongoObject.GetNoteCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while getting the note collection.\n\tError: %s", err.Error())})
			logger.Log.Fatalf("Error: Problem while getting the note collection.\n\tError: %s", err.Error())
		}

		// Get the note id from the url.
		noteIdAny, exists := c.Get("id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error: No note id given, hence could not find any note in the db."})
			logger.Log.Fatalf("Error: No note id given, hence could not find any note in the db.")
		}

		noteId, ok := noteIdAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while converting note id from any to string."})
			logger.Log.Fatal("Error: Problem while converting note id from any to string.")
		}

		// Create a filter with the note id.
		filter := bson.M{"_id": noteId}

		// Find the note to be shared and store it in a note data struct.
		var foundNote models.NoteData

		err = noteCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&foundNote)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: Problem while trying to find the note with note id: %s\n\tError: %s", noteId, err.Error())})
			logger.Log.Fatalf("Error: Problem while trying to find the note with note id: %s\n\tError: %s", noteId, err.Error())
		}

		// Get the other user id to whom the note will be shared from the request body.
		var userId string
		err = c.BindJSON(&userId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: No user id with whom the data is to be shared is given./Problem while trying to bind the data.\n\tError: %s", err.Error())})
			logger.Log.Fatalf("Error: No user id with whom the data is to be shared is given./Problem while trying to bind the data.\n\tError: %s", err.Error())
		}

		// Store the copy of the document for the other user.
		var insertNote models.NoteData

		insertNote.ID = primitive.NewObjectID()
		insertNote.User_Id = &userId
		insertNote.Header = foundNote.Header

		unqiueHeader := fmt.Sprintf("%s%s", *insertNote.User_Id, *insertNote.Header)
		insertNote.Unique_Header = &unqiueHeader

		// Get the email of the other user from the user collection.
		userCollection, err := database.MongoObject.GetUserCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while getting the user collection.\n\tError: %s", err.Error())})
			logger.Log.Fatalf("Error: Problem while getting the user collection.\n\tError: %s", err.Error())
		}

		filter = bson.M{"_id": userId}

		var foundUser models.UserDataServer

		err = userCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: Problem while finding the user data to whom the note is being shared.\n\tError: %s", err.Error())})
			logger.Log.Fatalf("Error: Problem while finding the user data to whom the note is being shared.\n\tError: %s", err.Error())
		}

		insertNote.Email = foundUser.Email

		insertNote.Data = foundNote.Data

		sharable := false
		insertNote.Sharable = &sharable

		createdAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			logger.Log.Fatalf("Error: Problem while storing the creation time tamp for the user.\n\tError: %s", err.Error())
		}

		updatedAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			logger.Log.Fatalf("Error: Problem while storing the update time tamp for the user.\n\tError: %s", err.Error())
		}

		insertNote.Created_At = createdAt
		insertNote.Updated_At = updatedAt

		_, err = noteCollection.InsertOne(database.MongoObject.Ctx, insertNote)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Fatalf("Error: Problem while inserting new data into the notes collection.\n\tError: %s", err.Error())
		}

		// Find the inserted note.
		var noteInserted models.NoteData

		filter = bson.M{"_id": insertNote.ID}

		err = noteCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&noteInserted)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while trying the find the inserted document in the DB."})
			logger.Log.Fatal("Error: Problem while trying the find the inserted document in the DB.")
		}

		c.JSON(http.StatusOK, noteInserted)
		logger.Log.Printf("Message: Successful creation of new note document with note id: %s", noteInserted.ID)
	}
}

// GET /api/search?q=:query: search for notes based on keywords for the authenticated user.

func SearchNotesByKeywords() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user id from the authentication.
		userIdAny, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error: No user id given for authentication."})
			logger.Log.Fatal("Error: No user id given for authentication.")
		}

		userId, ok := userIdAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while converting user id from any to string."})
			logger.Log.Fatal("Error: Problem while converting user id from any to string.")
		}

		// Get access to the user collection.
		noteCollection, err := database.MongoObject.GetNoteCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while getting the note collection.\n\tError: %s", err.Error())})
			logger.Log.Fatalf("Error: Problem while getting the note collection.\n\tError: %s", err.Error())
		}

		// Get the searchword.
		query := c.Query("q")

		// Create a index.
		_, err = noteCollection.Indexes().CreateOne(
			database.MongoObject.Ctx,
			mongo.IndexModel{
				Keys: bson.D{{Key: "notesData", Value: "text"}},
			},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while trying to create a index.\n\tError: %s", err.Error())})
			logger.Log.Fatalf("Error: Problem while trying to create a index.\n\tError: %s", err.Error())
		}

		// Create a filter.
		filter := bson.M{
			"$text": bson.M{
				"$search": query,
			},
		}

		// Create the cursor.
		cursor, err := noteCollection.Find(database.MongoObject.Ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while creating the cursor.\n\tError: %s", err.Error())})
			logger.Log.Fatalf("Error: Problem while creating the cursor.\n\tError: %s", err.Error())
		}
		defer cursor.Close(database.MongoObject.Ctx)

		// Declare the list of notes matching the criteria.
		var foundNotes []models.NoteData

		for cursor.Next(database.MongoObject.Ctx) {
			var foundNote models.NoteData

			err := cursor.Decode(&foundNote)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while decoding the found note.\n\tError: %s", err.Error())})
				logger.Log.Fatalf("Error: Problem while decoding the found note.\n\tError: %s", err.Error())
			}

			if *foundNote.User_Id == userId || *foundNote.Sharable {
				foundNotes = append(foundNotes, foundNote)
			}
		}

		// Send status ok with the list of found notes.
		c.JSON(http.StatusOK, foundNotes)
		logger.Log.Printf("Message: Successfully find all the notes with the keyword: %s", query)
	}
}
