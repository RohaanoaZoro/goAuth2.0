package main

import (
	"log"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Create the route
	// router.HandleFunc("/credslogin", CredsLogin)
	router.HandleFunc("/authenticate", Authentication)
	router.HandleFunc("/authorize", Authorization)
	router.HandleFunc("/verifytoken", VerifyTokenAPI)
	router.HandleFunc("/refreshtoken", RefreshTokenAPI)
	// // router.HandleFunc("/refresh", Refresh)
	// router.HandleFunc("/test", Test)
	router.HandleFunc("/registerApp", RegisterApplication)
	router.HandleFunc("/registerClient", RegisterClient)

	return router
}

func main() {

	log.Println("Go auth Started")

	JwtKeyLocker = mysql_getActiveSessions()

	router := NewRouter()
	if err := http.ListenAndServe(":2011", router); err != nil {
		log.Fatal("ListenAndServe Error: ", err)
	}
}

func sendRes(w http.ResponseWriter, jsonData *simplejson.Json) {

	// JSON encode jsonData
	payload, err := jsonData.MarshalJSON()
	if err != nil {
		log.Println(err, "\tstatus_code: 992")
		http.Error(w, "Internal Error", http.StatusMethodNotAllowed)
		return
	}

	// Return response JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}
