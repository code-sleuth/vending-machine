package controllers

import "github.com/gorilla/mux"

// StartUp function registers all routes
func StartUp(mux *mux.Router) {
	UserController{mux}.registerRoutes()
	TransactionController{mux}.registerRoutes()
}
