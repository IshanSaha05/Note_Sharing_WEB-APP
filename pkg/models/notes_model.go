package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NoteData struct {
	ID            primitive.ObjectID `bson:"_id"`                              // will be created
	User_Id       *string            `json:"userId" bson:"userId"`             // will be taken from middleware
	Header        *string            `json:"header" bson:"header"`             // will be provided in request
	Unique_Header *string            `json:"uniqueHeader" bson:"uniqueHeader"` // will be created
	Email         *string            `json:"email" bson:"email"`               // will be taken from middleware
	Data          *string            `json:"notesData" bson:"notesData"`       // will be provided in request
	Sharable      *bool              `json:"sharable" bson:"sharable"`         // will be provided in request
	Created_At    time.Time          `json:"createdAt" bson:"createdAt"`       // will be created
	Updated_At    time.Time          `json:"updatedAt" bson:"updatedAt"`       // will be created
}
