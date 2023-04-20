package main

import (
	"encoding/json"
	"log"
	"net/http"
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

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	// Create empty return JSON
	jsonData := simplejson.New()

	type ClientStruct struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		ClientId string `json:"clientid"`
	}

	//Decode json into structure
	var ClientDetails ClientStruct
	err := json.NewDecoder(r.Body).Decode(&ClientDetails)
	if err != nil {
		jsonData.Set("status", "Failed")
		jsonData.Set("message", "Cannot Decode Json")
		w.WriteHeader(http.StatusBadRequest)
		sendRes(w, jsonData)
		return
	}

	//Email is missing
	if ClientDetails.Email == "" {
		jsonData.Set("status", "Failed")
		jsonData.Set("message", "Email missing")
		w.WriteHeader(http.StatusBadRequest)
		sendRes(w, jsonData)
		return
	}

	//Password is missing
	if ClientDetails.Password == "" {
		jsonData.Set("status", "Failed")
		jsonData.Set("message", "Password missing")
		w.WriteHeader(http.StatusBadRequest)
		sendRes(w, jsonData)
		return
	}

	//Check if Client ID is missing
	if ClientDetails.ClientId == "" {
		jsonData.Set("status", "Failed")
		jsonData.Set("message", "Client Id missing")
		w.WriteHeader(http.StatusBadRequest)
		sendRes(w, jsonData)
		return
	}

	err = mysql_registerUser(ClientDetails.Email, ClientDetails.Password, ClientDetails.ClientId)
	if err != nil {
		log.Println("Error into inserting User Table")
		w.WriteHeader(http.StatusInternalServerError)
		jsonData.Set("status", "Failed")
		jsonData.Set("message", "Coulf not insert into database: "+err.Error())
		sendRes(w, jsonData)
		return
	}

	data := map[string]string{
		"ClientId": ClientDetails.ClientId,
	}
	jsonData.Set("status", "Success")
	jsonData.Set("status_code", "200")
	jsonData.Set("message", "User Registration Successful")
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

func AuthenticationAPI(w http.ResponseWriter, r *http.Request) {

	log.Println("/Authentication")

	jsonData := simplejson.New()

	type Credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var credentials Credentials
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clientid, authStatus := Authentication(credentials.Email, credentials.Password)
	if !authStatus {
		w.WriteHeader(http.StatusUnauthorized)

		jsonData.Set("status", "Failed")
		jsonData.Set("message", "Authentication Failed")
		sendRes(w, jsonData)
		return
	}

	clientSecret, _, _, clientSecretExists := mysql_getClientSecret(clientid)
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
		ClientId     string `json:"clientid"`
		ClientSecret string `json:"clientsecret"`
		GrantType    string `json:"grant_type"`
		Scope        string `json:"scope"`
	}

	queryParams := r.URL.Query()

	var credentials ClientDetailsStruct
	credentials.ClientId = queryParams.Get("clientid")
	credentials.ClientSecret = queryParams.Get("clientsecret")
	credentials.GrantType = queryParams.Get("grant_type")
	credentials.Scope = queryParams.Get("scope")

	clientSecret, jwtkey, ApplicationId, clientSecretExists := mysql_getClientSecret(credentials.ClientId)
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
