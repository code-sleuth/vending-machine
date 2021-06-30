package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/code-sleuth/vending-machine/helpers"
	"github.com/code-sleuth/vending-machine/models"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

// CreateUser function
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User

	email, ok := CheckIfUserSessionIsActive(w, r)
	if !ok {
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	if r.ContentLength == 0 {
		helpers.ErrorResponse(w, http.StatusBadRequest, "empty json body")
		return
	}

	if len(user.Username) == 0 || len(user.Password) == 0 || len(user.Role) == 0 {
		helpers.ErrorResponse(w, http.StatusBadRequest, "username, role and password should not be empty")
		return
	}

	_, ok = models.UserCanCRUD(email)
	if !ok {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to create user or you're logged out")
		return
	}

	u, err := user.CreateUser(user.Username, user.Password, string(user.Role), user.Deposit)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "unable to create user "+err.Error())
		return
	}

	helpers.JSONResponse(w, http.StatusCreated, u)
}

// GetUsers function
func GetUsers(w http.ResponseWriter, r *http.Request) {
	var u models.User

	email, ok := CheckIfUserSessionIsActive(w, r)
	if !ok {
		return
	}

	_, ok = models.UserCanCRUD(email)
	if !ok {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to list users")
		return
	}

	userList, err := u.GetUsers()
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNoContent, "could not get users from database | "+err.Error())
		return
	}

	helpers.JSONResponse(w, http.StatusOK, userList)
}

// GetUser by id
func GetUser(w http.ResponseWriter, r *http.Request) {
	var u models.User

	email, ok := CheckIfUserSessionIsActive(w, r)
	if !ok {
		return
	}

	params := mux.Vars(r)

	uid, err := helpers.ConvertStringToUint(params["id"])
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNotFound, "invalid character in route:"+err.Error())
		return
	}

	id, ok := models.UserCanCRUD(email)
	if !ok && uid != id {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to get user")
		return
	}

	user, err := u.GetUser(uid)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	helpers.JSONResponse(w, http.StatusOK, user)
}

// UpdateUser function
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	params := mux.Vars(r)

	email, ok := CheckIfUserSessionIsActive(w, r)
	if !ok {
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	uid, err := helpers.ConvertStringToUint(params["id"])
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNotFound, "invalid character in route:"+err.Error())
		return
	}

	id, ok := models.UserCanCRUD(email)
	if !ok && uid != id {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to update user")
		return
	}

	u, err := user.UpdateUser(uid, user.Username, string(user.Role), user.Deposit)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	helpers.JSONResponse(w, http.StatusAccepted, u)
}

// ChangePassword function
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	var user models.User
	var changePassword models.ChangePassword
	params := mux.Vars(r)
	// dataFromRequest := getDataFromRequest()

	email, ok := CheckIfUserSessionIsActive(w, r)
	if !ok {
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&changePassword); err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	uid, err := helpers.ConvertStringToUint(params["id"])
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNotFound, "invalid character in route:"+err.Error())
		return
	}

	id, ok := models.UserCanCRUD(email)
	if !ok && uid != id {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to update user")
		return
	}

	u, err := user.ChangePassword(uid, changePassword.OldPassword, changePassword.NewPassword, changePassword.ConfirmPassword)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	helpers.JSONResponse(w, http.StatusAccepted, u)
}

// DeleteUser function
func DeleteUser(w http.ResponseWriter, r *http.Request) {

	email, ok := CheckIfUserSessionIsActive(w, r)
	if !ok {
		return
	}

	if _, ok := models.UserCanCRUD(email); !ok {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to delete user")
		return
	}
	var user models.User

	params := mux.Vars(r)

	uid, err := helpers.ConvertStringToUint(params["id"])
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNotFound, "invalid character in route:"+err.Error())
		return
	}

	d, err := user.DeleteUser(uid)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	helpers.JSONResponse(w, http.StatusAccepted, map[string]string{"success": d})
}

// Login function
func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User

	if r.Method == "OPTIONS" {
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	token, err := user.Login(user.Username, user.Password)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusForbidden, err.Error())
		return
	}

	// Create a new random session token
	sessionToken := uuid.NewV4().String()
	// Set the token in the cache, along with the user whom it represents
	// The token has an expiry time of 7200 seconds (120 minutes or 2 hours)
	_, err = cache.Do("SETEX", sessionToken, "7200", user.Username)
	if err != nil {
		// If there is an error in setting the cache, return an internal server error
		helpers.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Finally, we set the client cookie for "session_token" as the session token we just generated
	// we also set an expiry time of 120 minutes, the same as the cache
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: time.Now().Add(120 * time.Minute),
		Path:    "/",
	})

	u, err := user.GetUserByUsername(user.Username)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	successMap := map[string]interface{}{
		"success":  "logged in successfully",
		"username": u.Username,
		"deposit":  u.Deposit,
		"role":     string(u.Role),
		"Token":    token,
	}

	helpers.JSONResponse(w, http.StatusOK, successMap)
}

// Store the redis connection as a package level variable
var cache redis.Conn

// InitCache function
func InitCache() {
	// Initialize the redis connection to a redis instance running on your local machine
	conn, err := redis.DialURL(helpers.GetEnv("REDIS_URL", "redis://localhost"))
	if err != nil {
		log.Fatalf("failed to start redis server: %+v", err)
	}
	// Assign the connection to the package level `cache` variable
	cache = conn
}

// CheckIfUserSessionIsActive function
func CheckIfUserSessionIsActive(w http.ResponseWriter, r *http.Request) (string, bool) {
	// We can obtain the session token from the requests cookies, which come with every request
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			helpers.ErrorResponse(w, http.StatusUnauthorized, "Not Authorized, please login")
			return "", false
		}
		// For any other type of error, return a bad request status
		helpers.ErrorResponse(w, http.StatusBadRequest, "Bad request, please login")
		return "", false
	}
	sessionToken := c.Value

	// We then get the email of the user from our cache, where we set the session token
	email, err := cache.Do("GET", sessionToken)
	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		helpers.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error, please login")
		return "", false
	}
	if email == nil {
		// If the session token is not present in cache, return an unauthorized error
		helpers.ErrorResponse(w, http.StatusUnauthorized, "Not Authorized, please login")
		return "", false
	}

	// Return true if user has active session
	return fmt.Sprintf("%s", email), true
}
