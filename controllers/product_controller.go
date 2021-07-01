package controllers

import (
	"github.com/code-sleuth/vending-machine/handlers"
	"github.com/code-sleuth/vending-machine/helpers"
	"github.com/gorilla/mux"
)

// ProductController struct
type ProductController struct {
	Router *mux.Router
}

// registerRoutes registers the user routes
func (controller ProductController) registerRoutes() {
	controller.Router.HandleFunc("/api/products", helpers.IsAuthorized(handlers.CreateProduct)).Methods("POST")
	controller.Router.HandleFunc("/api/products", handlers.GetProducts).Methods("GET")
	controller.Router.HandleFunc("/api/products/{id}", handlers.GetProduct).Methods("GET")
	controller.Router.HandleFunc("/api/products/{id}", helpers.IsAuthorized(handlers.UpdateProduct)).Methods("PUT")
	controller.Router.HandleFunc("/api/products/{id}", helpers.IsAuthorized(handlers.DeleteProduct)).Methods("DELETE")
}
