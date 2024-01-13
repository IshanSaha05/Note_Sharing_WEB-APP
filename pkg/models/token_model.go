package models

import "github.com/dgrijalva/jwt-go"

type SignedDetails struct {
	Email      string
	First_Name string
	Last_Name  string
	User_Id    string
	jwt.StandardClaims
}
