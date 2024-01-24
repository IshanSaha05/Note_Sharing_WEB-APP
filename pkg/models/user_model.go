package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserDataClient struct {
	ID         primitive.ObjectID `bson:"_id"`
	First_Name *string            `json:"firstName" bson:"firstName"`
	Last_Name  *string            `json:"lastName" bson:"lastName"`
	Password   *string            `json:"password" bson:"password"`
	Email      *string            `json:"email" bson:"email"`
	Token      *string            `json:"token" bson:"token"`
	UserID     string             `json:"userId" bson:"userId"`
}

type UserDataServer struct {
	ID            primitive.ObjectID `bson:"_id"`
	First_Name    *string            `json:"firstName" bson:"firstName"`
	Last_Name     *string            `json:"lastName" bson:"lastName"`
	Password      *string            `json:"password" bson:"password"`
	Email         *string            `json:"email" bson:"email"`
	Created_At    time.Time          `json:"createdAt" bson:"createdAt"`
	Updated_At    time.Time          `json:"updatedAt" bson:"updatedAt"`
	Last_Login    time.Time          `json:"lastLogin" bson:"lastLogin"`
	Refresh_Token *string            `json:"refreshToken" bson:"refreshToken"`
	UserID        string             `json:"userId" bson:"userId"`
}
