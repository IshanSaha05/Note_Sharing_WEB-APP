package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserDataClient struct {
	ID         primitive.ObjectID `bson:"_id"`
	First_Name *string            `json:"firstName"`
	Last_Name  *string            `json:"lastName"`
	Password   *string            `json:"password"`
	Email      *string            `json:"email"`
	Token      *string            `json:"token"`
	UserID     string             `json:"userId"`
}

type UserDataServer struct {
	ID            primitive.ObjectID `bson:"_id"`
	First_Name    *string            `json:"firstName"`
	Last_Name     *string            `json:"lastName"`
	Password      *string            `json:"password"`
	Email         *string            `json:"email"`
	Created_At    time.Time          `json:"createdAt"`
	Updated_At    time.Time          `json:"updatedAt"`
	Last_Login    time.Time          `json:"lastLogin"`
	Refresh_Token *string            `json:"refreshToken"`
	UserID        string             `json:"userId"`
}
