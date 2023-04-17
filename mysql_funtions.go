package main

import (
	"log"
	"time"
)

func mysql_registerApplication(ApplicationId string, ApplicationName string) (bool, string) {

	mysqlDB := MySQLConnect()
	defer mysqlDB.Close()

	Query := "INSERT INTO `Outh2`.`Application` (`ApplicationId`, `ApplicationName`) VALUES (?,?)"
	InsertQuery, err := mysqlDB.Prepare(Query)
	if err != nil {
		return false, err.Error()
	}
	InsertQuery.Exec(ApplicationId, ApplicationName)

	return true, ""
}

func mysql_registerClient(ClientId string, ClientSecret string, ApplicationId string) (bool, string) {

	mysqlDB := MySQLConnect()
	defer mysqlDB.Close()

	Query := "INSERT INTO `Outh2`.`Client` (`ClientId`, `ClientSecret`, `ApplicationId`) VALUES (?,?,?)"
	InsertQuery, err := mysqlDB.Prepare(Query)
	if err != nil {
		return false, err.Error()
	}
	InsertQuery.Exec(ClientId, ClientSecret, ApplicationId)

	return true, ""
}

func mysql_addJwtKey(ApplicationId string, JWTKey string) (bool, string) {

	mysqlDB := MySQLConnect()
	defer mysqlDB.Close()

	Query := "INSERT INTO `Outh2`.`JWTKeys` (`ApplicationId`, `JWTKey`) VALUES (?,?)"
	InsertQuery, err := mysqlDB.Prepare(Query)
	if err != nil {
		return false, err.Error()
	}
	InsertQuery.Exec(ApplicationId, JWTKey)

	return true, ""
}

func mysql_allSessionsUser(clientid string) []string {

	db := MySQLConnect()
	defer db.Close()

	var Query string = "SELECT SessionId FROM Outh2.Sessions WHERE ClientId='" + clientid + "'"
	rows, err := db.Query(Query)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var SessionArr []string

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var SessionId string
		if err := rows.Scan(&SessionId); err != nil {
			return SessionArr
		}
		SessionArr = append(SessionArr, SessionId)
	}

	if err = rows.Err(); err != nil {
		return SessionArr
	}

	return SessionArr
}

func mysql_deleteSession(SessionId string) (bool, string) {

	mysqlDB := MySQLConnect()
	defer mysqlDB.Close()

	//Creates a new Session
	Query := "DELETE FROM `Oauth2`.`Sessions` WHERE SessionId='" + SessionId + "'"
	InsertQuery, err := mysqlDB.Prepare(Query)
	if err != nil {
		log.Println("Error into deleting Session", SessionId)
		return false, err.Error()
	}
	InsertQuery.Exec()

	return true, ""
}

func mysql_getClientSecret(clientid string) (string, string, bool) {

	msDB := MySQLConnect()
	defer msDB.Close()

	type ClientInfo struct {
		ClientSecret string
		JWTKey       string
	}
	var clientInfo ClientInfo
	var Query string = "SELECT ClientSecret, JWTKey FROM Outh2.Client JOIN Outh2.JWTKeys WHERE Outh2.Client.ApplicationId=Outh2.JWTKeys.ApplicationId AND clientId='" + clientid + "'"
	err := msDB.QueryRow(Query).Scan(&clientInfo.ClientSecret, &clientInfo.JWTKey)
	if err != nil {
		log.Println("Error in Getting Client Secret : ", err)
		return "", "", false
	}

	return clientInfo.ClientSecret, clientInfo.JWTKey, true
}

func mysql_getJwtKey(applicationId string) (string, bool) {

	msDB := MySQLConnect()
	defer msDB.Close()

	var JWTKey string
	var Query string = "SELECT JWTKey FROM Outh2.JWTKeys WHERE ApplicationId='" + applicationId + "'"
	err := msDB.QueryRow(Query).Scan(&JWTKey)
	if err != nil {
		log.Println("Error in Getting Client Secret : ", err)
		return JWTKey, false
	}

	return JWTKey, true
}

func mysql_createSession(SessionId string, ClientId string, Token string, RefreshToken string) (bool, string) {

	mysqlDB := MySQLConnect()
	defer mysqlDB.Close()

	t := time.Now()
	SessionStartTime := t.Format("2006-01-02 15:04:05")

	//Creates a new Session
	Query := "INSERT INTO `Outh2`.`Sessions` (`SessionId`, `ClientId`, `Token`, `RefreshToken`, `SessionStartTime`, `SessionDuration`) VALUES (?,?,?,?,?,?)"
	InsertQuery, err := mysqlDB.Prepare(Query)
	if err != nil {
		log.Println("Error creating Session")
		return false, err.Error()
	}
	InsertQuery.Exec(SessionId, ClientId, Token, RefreshToken, SessionStartTime, "15")

	return true, ""
}

func mysql_verifyClientInfo(clientid string, clientSecret string) bool {

	msDB := MySQLConnect()
	defer msDB.Close()

	var clientInfo int
	var Query string = "SELECT Count(*) FROM Outh2.Client WHERE clientId='" + clientid + "'"
	err := msDB.QueryRow(Query).Scan(&clientInfo)
	if err != nil {
		log.Println("Error in Getting Client Secret : ", err)
		return false
	}

	if clientInfo > 0 {
		return true
	}

	return false
}

func mysql_getActiveSessions() map[string]string {

	db := MySQLConnect()
	defer db.Close()

	var Query string = "SELECT ApplicationId,JWTKey FROM Outh2.JWTKeys"
	rows, err := db.Query(Query)
	if err != nil {
		return map[string]string{}
	}
	defer rows.Close()

	var tempSessionArr map[string]string = make(map[string]string)

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var ApplicationId string
		var JwtKey string
		if err := rows.Scan(&ApplicationId, &JwtKey); err != nil {
			return tempSessionArr
		}
		tempSessionArr[ApplicationId] = JwtKey
	}

	if err = rows.Err(); err != nil {
		return tempSessionArr
	}

	return tempSessionArr
}
