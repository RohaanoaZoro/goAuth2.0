package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

func LanquillAuthentication(username string, password string) (string, bool) {

	UserDetails := GetUserDetails(username)

	if EncryptedPassword(password) != UserDetails.Password {
		log.Println("Epass Not Same")
		return "", false
	}

	UserType := strings.Compare(UserDetails.User_Type, "2")
	if UserType != 0 {
		log.Println("User_Type Not Same")
		return "", false
	}

	UserStatus := strings.Compare(UserDetails.User_Status, "Active")
	if UserStatus != 0 {
		return "", false
	}

	return strconv.Itoa(UserDetails.User_ID), true
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
