package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/code-sleuth/vending-machine/db"
	"github.com/code-sleuth/vending-machine/helpers"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

type Service interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
	DepositAmount(w http.ResponseWriter, r *http.Request)
	Buy(w http.ResponseWriter, r *http.Request)
	Reset(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)

	CreateProduct(w http.ResponseWriter, r *http.Request)
	GetProduct(w http.ResponseWriter, r *http.Request)
	UpdateProduct(w http.ResponseWriter, r *http.Request)
	DeleteProductHandler(w http.ResponseWriter, r *http.Request)
}

type service struct {
	db db.Service
}

func New(db db.Service) Service {
	return &service{
		db: db,
	}
}

// CreateUser handler
func (s *service) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user db.User

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

	u, err := s.db.CreateUser(&user)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "unable to create user "+err.Error())
		return
	}
	helpers.JSONResponse(w, http.StatusCreated, u)
}

// GetUser handler
func (s *service) GetUser(w http.ResponseWriter, r *http.Request) {

	_, ok := s.CheckIfUserSessionIsActive(w, r)
	if !ok {
		return
	}

	params := mux.Vars(r)

	user, err := s.db.GetUser(params["id"])
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	if user.UUID != params["id"] {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to get user")
		return
	}

	helpers.JSONResponse(w, http.StatusOK, user)
}

// UpdateUser handler
func (s *service) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var user db.User
	params := mux.Vars(r)

	username, ok := s.CheckIfUserSessionIsActive(w, r)
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

	if username != user.Username {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to update user")
		return
	}

	userUUID := params["id"]
	usr, err := s.db.GetUser(userUUID)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	usr.Deposit = user.Deposit

	u, err := s.db.UpdateUser(usr)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	helpers.JSONResponse(w, http.StatusAccepted, u)
}

// DeleteUser handler
func (s *service) DeleteUser(w http.ResponseWriter, r *http.Request) {

	username, ok := s.CheckIfUserSessionIsActive(w, r)
	if !ok {
		return
	}

	params := mux.Vars(r)
	userUUID := params["id"]

	user, err := s.db.GetUser(userUUID)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}

	if username != user.Username {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to delete user")
		return
	}

	err = s.db.DeleteUser(userUUID)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	d := fmt.Sprintf("user with id: %+v deleted", params["id"])

	helpers.JSONResponse(w, http.StatusAccepted, map[string]string{"success": d})
}

// DepositAmount handler
func (s *service) DepositAmount(w http.ResponseWriter, r *http.Request) {
	var user db.User
	params := mux.Vars(r)

	username, ok := s.CheckIfUserSessionIsActive(w, r)
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

	if username != user.Username {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to DepositAmount")
		return
	}

	amount, err := helpers.ConvertStringToInt(params["amount"])
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNotFound, "invalid character in route for amount:"+err.Error())
		return
	}

	u, err := s.db.Deposit(params["id"], amount)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	helpers.JSONResponse(w, http.StatusOK, u)
}

// Buy handler
func (s *service) Buy(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	username, ok := s.CheckIfUserSessionIsActive(w, r)
	if !ok {
		return
	}
	userUUID := params["id"]
	productUUID := params["productId"]

	user, err := s.db.GetUser(userUUID)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}

	if username != user.Username {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to make purchase")
		return
	}

	if user.Role != "buyer" {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to make purchase, make sure user is a buyer")
		return
	}

	amountOfProducts, err := helpers.ConvertStringToInt(params["amountOfProducts"])
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNotFound, "invalid character in route for amount:"+err.Error())
		return
	}

	u, err := s.db.Buy(userUUID, productUUID, amountOfProducts)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	helpers.JSONResponse(w, http.StatusOK, u)
}

// Reset handler
func (s *service) Reset(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	username, ok := s.CheckIfUserSessionIsActive(w, r)
	if !ok {
		return
	}

	userUUID := params["id"]
	user, err := s.db.GetUser(userUUID)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}

	if username != user.Username {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to reset deposit")
		return
	}

	if user.Role != "buyer" {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to reset deposit, make sure user is a buyer")
		return
	}

	u, err := s.db.Reset(userUUID)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	helpers.JSONResponse(w, http.StatusAccepted, u)
}

// Login function
func (s *service) Login(w http.ResponseWriter, r *http.Request) {
	var user db.User

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

	u, err := s.db.Login(user.Username, user.Password)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusUnauthorized, err.Error())
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

	token, err := helpers.GenerateJWT(user.UUID, user.Username)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	successMap := map[string]interface{}{
		"success": "logged in successfully",
		"user":    u,
		"token":   token,
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

// CheckIfUserSessionIsActive handler
func (s *service) CheckIfUserSessionIsActive(w http.ResponseWriter, r *http.Request) (string, bool) {
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

	// We then get the username of the user from our cache, where we set the session token
	username, err := cache.Do("GET", sessionToken)
	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		helpers.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error, please login")
		return "", false
	}
	if username == nil {
		// If the session token is not present in cache, return an unauthorized error
		helpers.ErrorResponse(w, http.StatusUnauthorized, "Not Authorized, please login")
		return "", false
	}

	// Return true if user has active session
	return fmt.Sprintf("%s", username), true
}

// CreateProduct handler
func (s *service) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product db.Product

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
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

	if len(product.ProductName) == 0 {
		helpers.ErrorResponse(w, http.StatusBadRequest, "product name should not be empty")
		return
	}

	p, err := s.db.CreateProduct(&product)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "unable to create product "+err.Error())
		return
	}
	helpers.JSONResponse(w, http.StatusCreated, p)
}

// GetProducts handler
//func (s *service) GetProducts(w http.ResponseWriter, r *http.Request) {
//	var p db.Product
//
//	productsList, err := s.db.GetProducts()
//	if err != nil {
//		helpers.ErrorResponse(w, http.StatusNoContent, "could not get products from database | "+err.Error())
//		return
//	}
//
//	helpers.JSONResponse(w, http.StatusOK, productsList)
//}

// GetProduct by id handler
func (s *service) GetProduct(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	uid := params["id"]
	product, err := s.db.GetProduct(uid)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	helpers.JSONResponse(w, http.StatusOK, product)
}

// UpdateProduct handler
func (s *service) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	var product db.Product
	params := mux.Vars(r)
	sellerUUID := params["id"]

	username, ok := s.CheckIfUserSessionIsActive(w, r)
	if !ok {
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	p, err := s.db.GetProduct(sellerUUID)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}

	user, err := s.db.GetUser(product.SellerID)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}

	if username != user.Username {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to update product")
		return
	}

	if user.Role != "seller" {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to update product, make sure user is a seller")
		return
	}

	p.ProductName = product.ProductName
	p.Cost = product.Cost
	p.AmountAvailable = product.AmountAvailable

	u, err := s.db.UpdateProduct(p)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	helpers.JSONResponse(w, http.StatusAccepted, u)
}

// DeleteProductHandler handler
func (s *service) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := s.CheckIfUserSessionIsActive(w, r)
	if !ok {
		return
	}

	params := mux.Vars(r)
	productUUID := params["id"]
	userUUID := params["userId"]

	user, err := s.db.GetUser(userUUID)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "bad request: [DeleteProductHandler] "+err.Error())
		return
	}

	if username != user.Username {
		helpers.ErrorResponse(w, http.StatusForbidden, "[DeleteProductHandler] insufficient rights to delete product")
		return
	}

	if user.Role != "seller" {
		helpers.ErrorResponse(w, http.StatusForbidden, "[DeleteProductHandler] insufficient rights to delete product, make sure user is a seller and owns the product")
		return
	}

	err = s.db.DeleteProduct(productUUID)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	d := fmt.Sprintf("product with id: %+v deleted", productUUID)

	helpers.JSONResponse(w, http.StatusAccepted, map[string]string{"success": d})
}
