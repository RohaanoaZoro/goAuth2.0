package main

import (
	"log"

	"github.com/dgrijalva/jwt-go"
)

func Authentication(username string, password string) (string, bool) {

	UserDetails, err := GetUserDetails(username)
	if err != nil {
		return err.Error(), false
	}

	// if EncryptedPassword(password) != UserDetails.Password {
	// 	log.Println("Epass Not Same")
	// 	return "", false
	// }

	if password != UserDetails.Password {
		log.Println("Epass Not Same")
		return "password mismatch", false
	}

	return UserDetails.Client_ID, true
}

func VerifyToken(token string, signKey string) (bool, string, string) {

	var jwtKey = []byte(signKey)
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(token, claims,
		func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return false, err.Error(), ""
		}
		return false, err.Error(), ""
	}

	if !tkn.Valid {
		return false, "token invalid", ""
	}

	log.Println("Claims", claims.ClientId)

	return true, "", claims.ClientId
}
