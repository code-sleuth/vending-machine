package controllers

import (
	"github.com/code-sleuth/vending-machine/helpers"
	"github.com/gorilla/mux"
)

// ProductController struct
type ProductController struct {
	Router *mux.Router
}

// registerRoutes registers the user routes
func (s *service) registerProductRoutes() {
	s.productController.Router.HandleFunc("/api/products", helpers.IsAuthorized(s.handlers.CreateProduct)).Methods("POST")
	//s.productController.Router.HandleFunc("/api/products", s.handlers.GetProducts).Methods("GET")
	s.productController.Router.HandleFunc("/api/products/{id}", s.handlers.GetProduct).Methods("GET")
	s.productController.Router.HandleFunc("/api/products/{id}", helpers.IsAuthorized(s.handlers.UpdateProduct)).Methods("PUT")
	s.productController.Router.HandleFunc("/api/products/{id}/{userId}", helpers.IsAuthorized(s.handlers.DeleteProductHandler)).Methods("DELETE")
}
