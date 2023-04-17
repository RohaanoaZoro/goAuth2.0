package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var MYSQLHOST = os.Getenv("MYSQL_HOST")
var MYSQLPORT = os.Getenv("MYSQL_PORT")
var MYSQLUSER = os.Getenv("MYSQL_USER")
var MYSQLPASS = os.Getenv("MYSQL_PASS")
var MYSQLDB = os.Getenv("MYSQL_AUTHDB")

func MySQLConnect() *sql.DB {

	mySQLdb, err := sql.Open("mysql", MYSQLUSER+":"+MYSQLPASS+"@tcp("+MYSQLHOST+":"+MYSQLPORT+")/"+MYSQLDB+"?parseTime=true")
	if err != nil {
		log.Fatal("Unable to Connect to SQL:", err)
	}

	return mySQLdb
}
