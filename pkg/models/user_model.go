package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserData struct {
	ID            primitive.ObjectID `bson:"_id"`
	First_Name    *string            `json:"firstName"`
	Last_Name     *string            `json:"lastName"`
	Password      *string            `json:"password"`
	Email         *string            `json:"email"`
	Token         *string            `json:"token"`
	Refresh_Token *string            `json:"refreshToken"`
	Created_At    time.Time          `json:"createdAt"`
	Updated_At    time.Time          `json:"updatedAt"`
	UserID        string             `json:"userId"`
}
