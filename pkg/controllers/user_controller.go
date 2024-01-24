package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/IshanSaha05/jwt_authentication_rest_api/logger"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/database"
	"github.com/IshanSaha05/jwt_authentication_rest_api/pkg/models"
	"github.com/gin-gonic/gin"
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
			logger.Log.Print("Error: No user id present.")
			c.Abort()
			return
		}
		userId, ok := userIdAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error:": "Error: Problem while converting user id to string."})
			logger.Log.Print("Error: Problem while converting user if to string.")
			c.Abort()
			return
		}

		// Get the note collection.
		noteCollection, err := database.MongoObject.GetNoteCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while getting the note collection.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: Problem while getting the note collection.\n\tError: %s", err)
			c.Abort()
			return
		}

		// Make the filter for search --> use the note id.
		noteIdPrimitive, err := primitive.ObjectIDFromHex(notesId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while converting notes id to primitive object.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: Problem while converting notes id to primitve object.\n\tError: %s", err.Error())
			c.Abort()
			return
		}
		filter := bson.D{{Key: "_id", Value: noteIdPrimitive}}

		// Decode and bind it in a note user.
		var note models.NoteData

		err = noteCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&note)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: No such note present. Problem while decoding the note.\n\tError: %s", err.Error()), "data": notesId})
			logger.Log.Printf("Error: No such note present. Problem while decoding the note.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Check whether user id in the decoded version is same as the authenticated user id.
		// If yes, then send status ok and send the data with header field.
		// If no, then check if sharable is true.
		// If yes, then send status ok and send the data with header field.
		// Otherwise, send bad status.
		if userId == *note.User_Id {
			c.JSON(http.StatusOK, note)
			logger.Log.Printf("Message: Note belongs to user with user id: %s and present in the db with notes id: %s. It is successfully shared.", userId, notesId)
			c.Abort()
			return
		} else {
			if *note.Sharable {
				c.JSON(http.StatusOK, note)
				logger.Log.Printf("Message: Note does not belong to user with user with user id: %s, but sharable and thus successfully shared notes with notes id: %s", userId, notesId)
				c.Abort()
				return
			} else {
				c.JSON(http.StatusBadRequest, nil)
				logger.Log.Printf("Message: Note does not belong to user with user id: %s and it is also not sharable publically. Hence, it cannot be shared.", userId)
				c.Abort()
				return
			}
		}
	}
}

// POST /api/notes: create a new note for the authenticated user.
func CreateNotes() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bind the note sent in the request body.
		var note models.NoteData

		err := c.BindJSON(&note)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while binding note from json.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: Problem while binding note from json.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Get the user id of the authenticated user from the middleware.
		userIdAny, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error: Problem while getting user id from the middleware."})
			logger.Log.Printf("Error: Problem while getting user id from the middleware.")
			c.Abort()
			return
		}
		userId, ok := userIdAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while converting user id to string."})
			logger.Log.Printf("Error: Problem while converting user id to string.")
			c.Abort()
			return
		}

		// Make the unqiue header.
		unqiueHeader := fmt.Sprintf("%s%s", userId, *note.Header)

		// Open the note collection.
		noteCollection, err := database.MongoObject.GetNoteCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while trying to get the note collection.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: Problem while trying to get the note collection.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Make the filter.
		filter := bson.D{{Key: "uniqueHeader", Value: unqiueHeader}}

		// Find the note if already present with the same unique header.
		findErr := noteCollection.FindOne(database.MongoObject.Ctx, filter)

		// If no document found, can create the note.
		if findErr.Err() == mongo.ErrNoDocuments {
			note.ID = primitive.NewObjectID()
			note.User_Id = &userId
			note.Unique_Header = &unqiueHeader

			emailAny, exists := c.Get("email")
			if !exists {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Error: No email id provided for the user."})
				logger.Log.Printf("Error: No email id provided for the user.")
				c.Abort()
				return
			}
			email, ok := emailAny.(string)
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while converting email to string."})
				logger.Log.Printf("Error: Problem while converting email to string.")
				c.Abort()
				return
			}
			note.Email = &email

			createdAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
				logger.Log.Printf("\nError: Problem while storing the creation time tamp for the user.\n\tError: %s", err.Error())
				c.Abort()
				return
			}
			note.Created_At = createdAt

			updatedAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
				logger.Log.Printf("\nError: Problem while storing the update time tamp for the user.\n\tError: %s", err.Error())
				c.Abort()
				return
			}
			note.Updated_At = updatedAt

			// Insert the document.
			_, err = noteCollection.InsertOne(database.MongoObject.Ctx, note)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while inserting the new document.\n\tError: %s", err.Error())})
				logger.Log.Printf("Error: Problem while inserting the new document.\n\tError: %s", err.Error())
				c.Abort()
				return
			}

			// Find the inserted document.
			var foundNote models.NoteData
			err = noteCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&foundNote)

			// If no such note found, then data not inserted.
			if err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: No such note yet has been created.\n\tError: %s", err.Error())})
					logger.Log.Printf("Error: No such note yet has been created.\n\tError: %s", err.Error())
					c.Abort()
					return
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while decoding the found note.\n\tError: %s", err.Error())})
					logger.Log.Printf("Error: Problem while decoding the found note.\n\tError: %s", err.Error())
					c.Abort()
					return
				}
			}

			// Note has been found, send status ok and send it.
			c.JSON(http.StatusOK, gin.H{"data": foundNote, "message": fmt.Sprintf("Message: Successfully created new note with note id: %s and uniquq header: %s", foundNote.ID, *foundNote.Unique_Header)})
			logger.Log.Printf("Message:  Successfully created new note with note id: %s and unique header: %s", foundNote.ID, *foundNote.Unique_Header)
			return
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error: Same note already exists in the database."})
			logger.Log.Printf("Error: Same note already exists in the database.")
			c.Abort()
			return
		}
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
			logger.Log.Print("Error: The user-id is not passed from the middleware the create handler through context.")
			c.Abort()
			return
		}
		userId, ok := userIdAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while converting any to string."})
			logger.Log.Print("Error: Problem while converting any to string.")
			c.Abort()
			return
		}

		// Get the notes collection.
		notesCollection, err := database.MongoObject.GetNoteCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Printf("Error: Problem while trying to open notes collection.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Find whether there is any notes present with the passed notes id in the url.
		noteIdPrimitive, err := primitive.ObjectIDFromHex(notesId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while converting notes id to primitive object.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: Problem while converting notes id to primitve object.\n\tError: %s", err.Error())
			c.Abort()
			return
		}
		filter := bson.D{{Key: "_id", Value: noteIdPrimitive}}

		var foundNotes models.NoteData

		err = notesCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&foundNotes)

		// If no notes present, send bad request.
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: No notes with id: %s is present in the database.", notesId)})
			logger.Log.Printf("Error: No notes with id: %s is present in the database.", notesId)
			c.Abort()
			return
		}

		// If authenticator user id is not same as document user id, send status bad request.
		if *foundNotes.User_Id != userId {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: Notes is not accessible to the user with user id: %s", userId)})
			logger.Log.Printf("Error: Notes is not accessible to the user with user id: %s", userId)
			c.Abort()
			return
		}

		// If same, then the user has access to the note document and can update.
		// Bind the json and extract the data into the note struct.
		var note *models.NoteData

		err = c.BindJSON(&note)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Printf("Error: Problem while binding the data to the note struct.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Fill in the other fields in the note which needs to be changed.
		updatedAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			logger.Log.Printf("Error: Problem while storing the update time tamp for the user.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		note.Updated_At = updatedAt

		// Make the update object.
		var updateMiniObj primitive.D

		if note.Header != nil {
			updateMiniObj = append(updateMiniObj, bson.E{Key: "header", Value: note.Header})

			uniqueHeader := fmt.Sprintf("%s%s", userId, *note.Header)
			updateMiniObj = append(updateMiniObj, bson.E{Key: "uniqueHeader", Value: uniqueHeader})
		}

		if note.Data != nil {
			updateMiniObj = append(updateMiniObj, bson.E{Key: "notesData", Value: note.Data})
		}

		if note.Sharable != nil {
			updateMiniObj = append(updateMiniObj, bson.E{Key: "sharable", Value: note.Sharable})
		}

		updateMiniObj = append(updateMiniObj, bson.E{Key: "updatedAt", Value: note.Updated_At})

		// Make the options for the find and update operation.
		options := options.FindOneAndUpdate().SetReturnDocument(options.After)

		// Make the update.
		var updateNote models.NoteData

		updateObj := primitive.D{
			{
				Key: "$set", Value: updateMiniObj,
			},
		}

		err = notesCollection.FindOneAndUpdate(database.MongoObject.Ctx, filter, updateObj, options).Decode(&updateNote)

		// If could not, send bad request.
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while upating data.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: Problem while upating data.\n\tError: %s", err.Error())
			c.Abort()
			return
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
		noteIdPrimitive, err := primitive.ObjectIDFromHex(noteId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while converting notes id to primitive object.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: Problem while converting notes id to primitve object.\n\tError: %s", err.Error())
			c.Abort()
			return
		}
		filter := bson.D{{Key: "_id", Value: noteIdPrimitive}}

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
		// Get the note id from the url.
		noteId := c.Param("id")

		// Get the sender user id.
		senderUserIdAny, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error: No user id given for authentication."})
			logger.Log.Print("Error: No user id given for authentication.")
		}
		senderUserId, ok := senderUserIdAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while converting user id from any to string."})
			logger.Log.Print("Error: Problem while converting user id from any to string.")
		}

		// Get the receiver user id.
		type receiver struct {
			UserId string `json:"userId"`
		}
		var receiverObj receiver

		err := c.BindJSON(&receiverObj)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: No user id with whom the data is to be shared is given./Problem while trying to bind the data.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: No user id with whom the data is to be shared is given./Problem while trying to bind the data.\n\tError: %s", err.Error())
		}

		receiverUserId := receiverObj.UserId
		fmt.Println(senderUserId, receiverUserId)

		// If the sender and receiver is same, sent bad status request.
		if senderUserId == receiverUserId {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error: Sender and receiver cannot be same."})
			logger.Log.Print("Error: Sender and receiver cannot be same.")
			c.Abort()
			return
		}

		// Get the note collection.
		noteCollection, err := database.MongoObject.GetNoteCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while getting the note collection.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: Problem while getting the note collection.\n\tError: %s", err.Error())
		}

		// Make the filter to use note id to search in the note database.
		noteIdPrimitve, err := primitive.ObjectIDFromHex(noteId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while trying to convert note id from string to primitive object.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: Problem while trying to convert note id from string to primitive object.\n\tError: %s", err.Error())
			c.Abort()
			return
		}
		filter := bson.D{{Key: "_id", Value: noteIdPrimitve}}

		// Find whether the note present in the database or not.
		var foundNote models.NoteData

		err = noteCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&foundNote)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: Problem while finding the note with note id: %s.\n\tError: %s", noteId, err.Error())})
			logger.Log.Printf("Error: Problem while finding the note with note id: %s.\n\tError: %s", noteId, err.Error())
			c.Abort()
			return
		}

		// Get the necessary receiver user details from the database.
		userCollection, err := database.MongoObject.GetUserCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while getting the user collection.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: Problem while getting the user collection.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		filter = bson.D{{Key: "userId", Value: receiverUserId}}

		var receiverUser models.UserDataServer

		err = userCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&receiverUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: Receiver does not exist in the database./Problem while decoding the found data.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: Receiver does not exist in the database./Problem while decoding the found data.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Create the insert note data.
		var insertNote models.NoteData
		insertNote.ID = primitive.NewObjectID()
		insertNote.User_Id = &receiverUserId
		insertNote.Header = foundNote.Header

		uniqueHeader := fmt.Sprintf("%s%s", *insertNote.User_Id, *insertNote.Header)
		insertNote.Unique_Header = &uniqueHeader

		// Check whether the any notes already present in the database with the same unique header by the receiver user.
		filter = bson.D{{Key: "userId", Value: insertNote.User_Id}, {Key: "uniqueHeader", Value: insertNote.Unique_Header}}

		findResult := noteCollection.FindOne(database.MongoObject.Ctx, filter)
		if findResult.Err() != mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error: User: %s already got a note with header: %s", *insertNote.User_Id, *insertNote.Header)})
			logger.Log.Printf("Error: User: %s already got a note with header: %s", *insertNote.User_Id, *insertNote.Header)
			c.Abort()
			return
		}

		insertNote.Email = receiverUser.Email
		insertNote.Data = foundNote.Data

		sharable := false
		insertNote.Sharable = &sharable

		createdAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			logger.Log.Printf("Error: Problem while storing the creation time tamp for the user.\n\tError: %s", err.Error())
			c.Abort()
			return
		}
		insertNote.Created_At = createdAt

		updatedAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			logger.Log.Printf("Error: Problem while storing the update time tamp for the user.\n\tError: %s", err.Error())
			c.Abort()
			return
		}
		insertNote.Updated_At = updatedAt

		_, err = noteCollection.InsertOne(database.MongoObject.Ctx, insertNote)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			logger.Log.Printf("Error: Problem while inserting new data into the notes collection.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Make the filter to find the inserted note.
		filter = bson.D{{Key: "_id", Value: insertNote.ID}}

		var insertedFoundNote models.NoteData

		err = noteCollection.FindOne(database.MongoObject.Ctx, filter).Decode(&insertedFoundNote)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: No note found./Problem while decoding the found data.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: No note found./Problem while decoding the found data.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, insertedFoundNote)
		logger.Log.Printf("Message: Successful creation of new note document with note id: %s", insertedFoundNote.ID)
	}
}

// GET /api/search?q=:query: search for notes based on keywords for the authenticated user.

func SearchNotesByKeywords() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user id.
		userIdAny, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error: No user id given for authentication."})
			logger.Log.Print("Error: No user id given for authentication.")
			c.Abort()
			return
		}
		userId, ok := userIdAny.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Problem while converting user id from any to string."})
			logger.Log.Print("Error: Problem while converting user id from any to string.")
			c.Abort()
			return
		}

		// Get the query.
		query := c.Query("q")

		// Get the note collection.
		noteCollection, err := database.MongoObject.GetNoteCollection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while getting the note collection.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: Problem while getting the note collection.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Create a index.
		_, err = noteCollection.Indexes().CreateOne(
			database.MongoObject.Ctx,
			mongo.IndexModel{
				Keys: bson.D{{Key: "notesData", Value: 1}},
			},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while trying to create a index.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: Problem while trying to create a index.\n\tError: %s", err.Error())
			c.Abort()
			return
		}

		// Create a filter.
		filter := bson.D{
			{Key: "$and", Value: bson.A{
				bson.D{
					{Key: "$or", Value: bson.A{
						bson.D{{Key: "userId", Value: userId}},
						bson.D{{Key: "sharable", Value: true}},
					}},
				},
				bson.D{{
					Key: "notesData", Value: bson.D{{
						Key: "$regex", Value: query,
					}},
				}},
			}},
		}

		// Get the cursor.
		cursor, err := noteCollection.Find(database.MongoObject.Ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while creating the cursor.\n\tError: %s", err.Error())})
			logger.Log.Printf("Error: Problem while creating the cursor.\n\tError: %s", err.Error())
			c.Abort()
			return
		}
		defer cursor.Close(database.MongoObject.Ctx)

		// Declare the list of notes.
		var foundNotes []models.NoteData

		// Iterate through the cursor
		for cursor.Next(database.MongoObject.Ctx) {
			var foundNote models.NoteData

			err := cursor.Decode(&foundNote)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: Problem while decoding the found note.\n\tError: %s", err.Error())})
				logger.Log.Fatalf("Error: Problem while decoding the found note.\n\tError: %s", err.Error())
			}

			// Append each found note.
			foundNotes = append(foundNotes, foundNote)
		}

		if foundNotes == nil {
			c.JSON(http.StatusOK, gin.H{"error": "Error: No such note contained the passed keywords in it."})
			logger.Log.Println("Error: No such note contained the passed keywords in it.")
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, foundNotes)
		logger.Log.Printf("Message: Successfully find all the notes with the keyword: %s", query)
	}
}

/*
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
		filter := bson.D{{Key: "$text", Value: bson.D{{Key: "$search", Value: query}}}}

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

			//if *foundNote.User_Id == userId || *foundNote.Sharable {
			foundNotes = append(foundNotes, foundNote)
			fmt.Println(foundNote)
			//}
		}

		// Send status ok with the list of found notes.
		fmt.Println(foundNotes)
		c.JSON(http.StatusOK, foundNotes)
		logger.Log.Printf("Message: Successfully find all the notes with the keyword: %s", query)
	}
}
*/
