package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NoteData struct {
	ID            primitive.ObjectID `bson:"_id"`
	User_Id       *string            `json:"userId"`
	Header        *string            `json:"header"`
	Unique_Header *string            `json:"uniqueHeader"`
	Email         *string            `json:"email"`
	Data          *string            `json:"notesData"`
	Sharable      *bool              `json:"sharable"`
	Created_At    time.Time          `json:"createdAt"`
	Updated_At    time.Time          `json:"updatedAt"`
}
