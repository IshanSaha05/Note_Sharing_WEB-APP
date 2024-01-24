package models

import "github.com/dgrijalva/jwt-go"

type SignedDetails struct {
	Email         string `json:"email" bson:"email"`
	First_Name    string `json:"firstName" bson:"firstName"`
	Last_Name     string `json:"lastName" bson:"lastName"`
	User_Id       string `json:"userId" bson:"userId"`
	Refresh_Token string `json:"refreshToken" bson:"refreshToken"`
	jwt.StandardClaims
}
