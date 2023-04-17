package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/google/uuid"
)

var JwtKeyLocker map[string]string = make(map[string]string)

func RegisterApplication(w http.ResponseWriter, r *http.Request) {

	// Create empty return JSON
	jsonData := simplejson.New()

	//Get the Application Name
	type ApplicationStruct struct {
		ApplicationName string `json:"appname"`
	}
	var AppDetails ApplicationStruct
	err := json.NewDecoder(r.Body).Decode(&AppDetails)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "401")
		jsonData.Set("message", "Cannot Decode Json")
		return
	}

	ApplicationId := uuid.New().String()
	jwtKey := uuid.New().String()

	isRegisterSuccess, errMsg := mysql_registerApplication(ApplicationId, AppDetails.ApplicationName)
	if !isRegisterSuccess {
		log.Println("Error into inserting Appliaction")
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "401")
		jsonData.Set("message", "App Registration failed. "+errMsg)
		sendRes(w, jsonData)
		return
	}

	isRegisterSuccess, errMsg = mysql_addJwtKey(ApplicationId, jwtKey)
	if !isRegisterSuccess {
		log.Println("Error into inserting App Auth")
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "401")
		jsonData.Set("message", "App Registration failed. "+errMsg)
		sendRes(w, jsonData)
		return
	}

	data := map[string]string{
		"ApplicationId":   ApplicationId,
		"ApplicationName": AppDetails.ApplicationName,
	}
	jsonData.Set("status", "Success")
	jsonData.Set("status_code", "200")
	jsonData.Set("message", "Application Registration Successful")
	jsonData.Set("data", data)
	sendRes(w, jsonData)
	return
}

func RegisterClient(w http.ResponseWriter, r *http.Request) {

	// Create empty return JSON
	jsonData := simplejson.New()

	type ClientStruct struct {
		ClientId      string `json:"clientid"`
		ClientSecret  string `json:"clientsecret"`
		ApplicationId string `json:"appid"`
	}

	//Decode json into structure
	var ClientDetails ClientStruct
	err := json.NewDecoder(r.Body).Decode(&ClientDetails)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "401")
		jsonData.Set("message", "Cannot Decode Json")
		sendRes(w, jsonData)
		return
	}

	//Generate new Client Id if client id is missing
	if ClientDetails.ClientId == "" {
		ClientDetails.ClientId = uuid.New().String()
	}

	//Generate new Client Secret if client secret is missing
	if ClientDetails.ClientSecret == "" {
		ClientDetails.ClientSecret = uuid.New().String()
	}

	//Check if ApplicationId is present
	if ClientDetails.ApplicationId == "" {
		w.WriteHeader(http.StatusBadRequest)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "401")
		jsonData.Set("message", "Application Id missing")
		sendRes(w, jsonData)
		return
	}

	isRegisterSuccess, errMsg := mysql_registerClient(ClientDetails.ClientId, ClientDetails.ClientSecret, ClientDetails.ApplicationId)
	if !isRegisterSuccess {
		log.Println("Error into inserting App Auth")
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "401")
		jsonData.Set("message", "App Registration failed. "+errMsg)
		sendRes(w, jsonData)
		return
	}

	data := map[string]string{
		"ClientId": ClientDetails.ClientId,
		"AppId":    ClientDetails.ApplicationId,
	}
	jsonData.Set("status", "Success")
	jsonData.Set("status_code", "200")
	jsonData.Set("message", "Client Registration Successful")
	jsonData.Set("data", data)

	sendRes(w, jsonData)
	return
}

func Authentication(w http.ResponseWriter, r *http.Request) {

	log.Println("/Authentication")

	jsonData := simplejson.New()

	type Credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var credentials Credentials
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clientid, authStatus := LanquillAuthentication(credentials.Username, credentials.Password)
	if !authStatus {
		w.WriteHeader(http.StatusUnauthorized)

		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "401")
		jsonData.Set("message", "Authentication Failed")
		sendRes(w, jsonData)
		return
	}

	clientSecret, _, clientSecretExists := mysql_getClientSecret(clientid)
	if !clientSecretExists {
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Client verification failed. Client may not exist.")
		sendRes(w, jsonData)
		return
	}

	jsonData.Set("status", "Success")
	jsonData.Set("status_code", "200")
	jsonData.Set("message", "Authentication Successful")
	jsonData.Set("clientid", clientid)
	jsonData.Set("clientsecret", clientSecret)

	sendRes(w, jsonData)
}

func Authorization(w http.ResponseWriter, r *http.Request) {

	log.Println("/Authorization")

	jsonData := simplejson.New()

	type ClientDetailsStruct struct {
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		GrantType    string `json:"grant_type"`
		Scope        string `json:"scope"`
	}
	var credentials ClientDetailsStruct

	//Read the Body of the request
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Unable to read body of request.")
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Unable to read body of request.")
		sendRes(w, jsonData)
		return
	}

	u, err := url.Parse(r.URL.Host + "?" + string(bodyBytes))
	if err != nil {
		log.Println("Unable to parse body of request.")
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Unable to parse body of request.")
		sendRes(w, jsonData)
		return
	}
	queryParams := u.Query()

	credentials.ClientId = queryParams["client_id"][0]
	credentials.ClientSecret = queryParams["client_secret"][0]
	credentials.GrantType = queryParams["grant_type"][0]
	credentials.Scope = queryParams["scope"][0]

	var ApplicationId string = r.URL.Query()["Application-Id"][0]

	clientSecret, jwtkey, clientSecretExists := mysql_getClientSecret(credentials.ClientId)
	if !clientSecretExists {
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Client verification failed. Client may not exist.")
		sendRes(w, jsonData)
		return
	}

	if clientSecret != credentials.ClientSecret {

		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Client verification failed. Client Secret Mismatch.")
		sendRes(w, jsonData)
		return
	}

	TotalSessionArr := mysql_allSessionsUser(credentials.ClientId)
	if len(TotalSessionArr) > 0 {
		// log.Println("Too Many Sessions. Deleting Oldest Session", TotalSessionArr[0])
		// mysql_deleteSession(TotalSessionArr[0])

		// w.WriteHeader(http.StatusInternalServerError)
		// jsonData.Set("status", "Failed")
		// jsonData.Set("status_code", "400")
		// jsonData.Set("message", "Session Already Exists ")
		// sendRes(w, jsonData)
		// return
	}

	sessionId := uuid.New().String()

	Token, isTokenGenerated := GenerateToken(credentials.ClientId, sessionId, jwtkey)
	if !isTokenGenerated {
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Cookie Creation/ Token Generation Failed")
		sendRes(w, jsonData)
		return
	}

	RefreshCookie, isValidCookie := GenerateRefreshToken(credentials.ClientId, clientSecret, sessionId, ApplicationId, jwtkey)
	if !isValidCookie {
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Cookie Creation/ Token Generation Failed")
		sendRes(w, jsonData)
		return
	}

	sessionCreated, errMsg := mysql_createSession(sessionId, credentials.ClientId, Token, RefreshCookie.Value)
	if !sessionCreated {
		log.Println("Error into Creating Session")
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Session Creation Failed "+errMsg)
		sendRes(w, jsonData)
		return
	}

	JwtKeyLocker[ApplicationId] = jwtkey

	http.SetCookie(w, &RefreshCookie)
	jsonData.Set("status", "Success")
	jsonData.Set("status_code", "200")
	jsonData.Set("message", "Authentication Successful")
	jsonData.Set("ClientId", credentials.ClientId)
	jsonData.Set("clientsecret", clientSecret)
	jsonData.Set("access_token", Token)

	sendRes(w, jsonData)
}

func VerifyTokenAPI(w http.ResponseWriter, r *http.Request) {

	log.Println("/Verify Token")
	jsonData := simplejson.New()

	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]

	log.Println("reqToken", reqToken)

	var ApplicationId string = r.URL.Query()["Application-Id"][0]

	var signatureKey string = JwtKeyLocker[ApplicationId]
	if len(signatureKey) == 0 {
		log.Println("Session Not available.")
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Session Not available.")
		sendRes(w, jsonData)
		return
	}

	isTokenVerified, errorMsg, _ := VerifyToken(reqToken, signatureKey)
	if !isTokenVerified {
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Token Verification Failed. "+errorMsg)
		sendRes(w, jsonData)
		return
	}

	jsonData.Set("status", "Success")
	jsonData.Set("status_code", "200")
	jsonData.Set("message", "Token Verified. Session is Active.")

	sendRes(w, jsonData)
}

func RefreshTokenAPI(w http.ResponseWriter, r *http.Request) {

	log.Println("/Refresh Token")

	jsonData := simplejson.New()

	var ApplicationId string = r.URL.Query()["Application-Id"][0]

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Failed to view Cookie. "+err.Error())
		sendRes(w, jsonData)
		return
	}

	RefreshToken := cookie.Value

	var signatureKey string = JwtKeyLocker[ApplicationId]
	if len(signatureKey) == 0 {
		log.Println("Session Not available.")
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Session Not available.")
		sendRes(w, jsonData)
		return
	}

	isTokenVerified, errorMsg, ClientId := VerifyToken(RefreshToken, signatureKey)
	if !isTokenVerified {
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Token Verification Failed. "+errorMsg)
		sendRes(w, jsonData)
		return
	}

	TotalSessionArr := mysql_allSessionsUser(ClientId)
	if len(TotalSessionArr) > 0 {
		log.Println("Too Many Sessions. Deleting Oldest Session", TotalSessionArr[0])
		mysql_deleteSession(TotalSessionArr[0])

		// w.WriteHeader(http.StatusInternalServerError)
		// jsonData.Set("status", "Failed")
		// jsonData.Set("status_code", "400")
		// jsonData.Set("message", "Session Already Exists ")
		// sendRes(w, jsonData)
		// return
	}

	sessionId := uuid.New().String()

	Token, isTokenGenerated := GenerateToken(ClientId, sessionId, signatureKey)
	if !isTokenGenerated {
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Cookie Creation/ Token Generation Failed")
		sendRes(w, jsonData)
		return
	}

	sessionCreated, errMsg := mysql_createSession(sessionId, ClientId, Token, RefreshToken)
	if !sessionCreated {
		log.Println("Error into Creating Session")
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("status_code", "400")
		jsonData.Set("message", "Session Creation Failed "+errMsg)
		sendRes(w, jsonData)
		return
	}

	jsonData.Set("status", "Success")
	jsonData.Set("status_code", "200")
	jsonData.Set("message", "Refresh Token Verified. New Token and Session Created.")
	jsonData.Set("access_token", Token)

	sendRes(w, jsonData)
}
