package controllers

import (
	"github.com/code-sleuth/vending-machine/handlers"
	"github.com/code-sleuth/vending-machine/helpers"
	"github.com/gorilla/mux"
)

// UserController struct
type UserController struct {
	Router *mux.Router
}

// registerRoutes registers the user routes
func (controller UserController) registerRoutes() {
	controller.Router.HandleFunc("/api/users/login", handlers.Login).Methods("POST", "OPTIONS")
	controller.Router.HandleFunc("/api/users", handlers.CreateUser).Methods("POST")
	controller.Router.HandleFunc("/api/users", helpers.IsAuthorized(handlers.GetUsers)).Methods("GET")
	controller.Router.HandleFunc("/api/users/{id}", helpers.IsAuthorized(handlers.GetUser)).Methods("GET")
	controller.Router.HandleFunc("/api/users/{id}", helpers.IsAuthorized(handlers.UpdateUser)).Methods("PUT")
	controller.Router.HandleFunc("/api/users/{id}/change_password", helpers.IsAuthorized(handlers.ChangePassword)).Methods("PUT")
	controller.Router.HandleFunc("/api/users/{id}", helpers.IsAuthorized(handlers.DeleteUser)).Methods("DELETE")
}
