package controllers

import (
	"github.com/code-sleuth/vending-machine/helpers"
	"github.com/gorilla/mux"
)

// UserController struct
type UserController struct {
	Router *mux.Router
}

// registerRoutes registers the user routes
func (s *service) registerUserRoutes() {
	s.userController.Router.HandleFunc("/api/users/login", s.handlers.Login).Methods("POST", "OPTIONS")
	s.userController.Router.HandleFunc("/api/users", s.handlers.CreateUser).Methods("POST")
	//s.userController.Router.HandleFunc("/api/users", helpers.IsAuthorized(s.handlers.GetUsers)).Methods("GET")
	s.userController.Router.HandleFunc("/api/users/{id}", helpers.IsAuthorized(s.handlers.GetUser)).Methods("GET")
	s.userController.Router.HandleFunc("/api/users/{id}", helpers.IsAuthorized(s.handlers.UpdateUser)).Methods("PUT")
	//s.userController.Router.HandleFunc("/api/users/{id}/change_password", helpers.IsAuthorized(handlers.ChangePassword)).Methods("PUT")
	s.userController.Router.HandleFunc("/api/users/{id}", helpers.IsAuthorized(s.handlers.DeleteUser)).Methods("DELETE")
	s.userController.Router.HandleFunc("/api/users/deposit/{id}/{amount}", helpers.IsAuthorized(s.handlers.DepositAmount)).Methods("POST")
	s.userController.Router.HandleFunc("/api/users/buy/{id}/{productId}/{amountOfProducts}", helpers.IsAuthorized(s.handlers.Buy)).Methods("POST")
	s.userController.Router.HandleFunc("/api/users/reset/{id}", helpers.IsAuthorized(s.handlers.Reset)).Methods("POST")
}
