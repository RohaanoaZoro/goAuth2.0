package main

import "log"

type UserDetails struct {
	User_ID     int
	User_Status string
	User_Type   string
	Password    string
}

func GetUserDetails(Email string) UserDetails {

	msDB := MySQLConnect()
	defer msDB.Close()

	var userDetails UserDetails
	var Query string = "SELECT User_ID, User_Status, User_Type, Password FROM sententia.user WHERE User_Email='" + Email + "'"
	err := msDB.QueryRow(Query).Scan(&userDetails.User_ID, &userDetails.User_Status, &userDetails.User_Type, &userDetails.Password)
	if err != nil {
		log.Println("Error in SQL Login : ", err)
	}

	return userDetails
}
