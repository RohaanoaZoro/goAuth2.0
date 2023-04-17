package main

import (
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type RefreshTokenClaims struct {
	ClientId      string `json:"clientid"`
	ClientSecret  string `json:"clientsecret"`
	SessionId     string `json:"sessionid"`
	ApplicationId string `json:"appId"`
	jwt.StandardClaims
}

type Claims struct {
	ClientId  string `json:"clientid"`
	SessionId string `json:"seesionid"`
	jwt.StandardClaims
}

func GenerateToken(clientId string, sessionId string, SignatureKey string) (string, bool) {

	expirationTime := time.Now().Add(time.Minute * 15)

	claims := Claims{
		ClientId:  clientId,
		SessionId: sessionId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	var jwtKey = []byte(SignatureKey)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Println("token err", err)
		return "", false
	}

	return tokenString, true
}

func GenerateRefreshToken(clientId string, clientSecret string, sessionId string, applicationId string, signatureKey string) (http.Cookie, bool) {

	expirationTime := time.Now().Add(time.Hour * 24)

	claims := RefreshTokenClaims{
		ClientId:      clientId,
		SessionId:     sessionId,
		ApplicationId: applicationId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	var jwtKey = []byte(signatureKey)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Println("token err", err)
		return http.Cookie{}, false
	}

	return http.Cookie{
		Name:     "refresh_token",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Hour * 24),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}, true
}
