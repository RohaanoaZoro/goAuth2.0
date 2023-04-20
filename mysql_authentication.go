package main

import (
	"log"
)

type UserDetails struct {
	Client_ID  string
	User_Email string
	Password   string
}

func GetUserDetails(Email string) (UserDetails, error) {

	msDB := MySQLConnect()
	defer msDB.Close()

	var userDetails UserDetails
	var Query string = "SELECT `ClientId`, `email`, `password` FROM `Oauth2`.`Users` WHERE email='" + Email + "'"
	err := msDB.QueryRow(Query).Scan(&userDetails.Client_ID, &userDetails.User_Email, &userDetails.Password)
	if err != nil {
		log.Println("Error in SQL Login : ", err)
		return userDetails, err
	}

	return userDetails, nil
}
