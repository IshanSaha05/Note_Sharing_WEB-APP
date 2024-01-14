package models

import "github.com/dgrijalva/jwt-go"

type SignedDetails struct {
	Email         string `json:"email"`
	First_Name    string `json:"firstName"`
	Last_Name     string `json:"lastName"`
	User_Id       string `json:"userID"`
	Refresh_Token string `json:"refreshToken"`
	jwt.StandardClaims
}
