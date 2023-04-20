package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var MYSQLHOST = "127.0.0.1"
var MYSQLPORT = "3306"
var MYSQLUSER = "root"
var MYSQLPASS = "Zxcvbnm@2"
var MYSQLDB = "Oauth2"

func MySQLConnect() *sql.DB {

	mySQLdb, err := sql.Open("mysql", MYSQLUSER+":"+MYSQLPASS+"@tcp("+MYSQLHOST+":"+MYSQLPORT+")/"+MYSQLDB+"?parseTime=true")
	if err != nil {
		log.Fatal("Unable to Connect to SQL:", err)
	}

	return mySQLdb
}
