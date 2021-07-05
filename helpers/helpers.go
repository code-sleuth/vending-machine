package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// ErrorResponse function
func ErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	JSONResponse(w, statusCode, map[string]string{"error": message})
}

// JSONResponse function
func JSONResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	resp, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(resp); err != nil {
		log.Println(err)
	}
}

// ConvertStringToUint function
func ConvertStringToUint(param string) (uint, error) {
	u64, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		return 0, err
	}
	uid := uint(u64)
	return uid, nil
}

// ConvertStringToInt function
func ConvertStringToInt(param string) (int, error) {
	u64, err := strconv.ParseInt(param, 10, 32)
	if err != nil {
		return 0, err
	}
	uid := int(u64)
	return uid, nil
}

// GetEnv function
func GetEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// GenerateJWT function
func GenerateJWT(uuid string, username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["uuid"] = uuid
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Minute * 120).Unix()

	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		err := fmt.Errorf("failed to generate token: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}

var signingKey = []byte(GetEnv("JWT_SECRET", ""))

// IsAuthorized function
func IsAuthorized(endpoint func(http.ResponseWriter, *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Header["Token"] != nil {

			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					ErrorResponse(w, http.StatusForbidden, "there was an error")
					return nil, fmt.Errorf("there was an error")
				}
				return signingKey, nil
			})

			if err != nil {
				ErrorResponse(w, http.StatusForbidden, "invalid token: "+err.Error())
			}

			if token != nil {
				if token.Valid {
					endpoint(w, r)
				}
			}
		} else {
			ErrorResponse(w, http.StatusForbidden, "Not Authorized")
		}
	}
}
