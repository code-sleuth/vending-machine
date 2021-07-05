package controllers

import (
	"github.com/code-sleuth/vending-machine/handlers"
	"github.com/gorilla/mux"
)

type Service interface {
	StartUp()

	registerUserRoutes()
	registerProductRoutes()
}

type service struct {
	handlers          handlers.Service
	userController    UserController
	productController ProductController
}

// New creates new instance of the handlers
func New(handlers handlers.Service, mux *mux.Router) Service {
	return &service{
		handlers:          handlers,
		userController:    UserController{mux},
		productController: ProductController{mux},
	}
}

// StartUp function registers all routes
func (s *service) StartUp() {
	s.registerUserRoutes()
	s.registerProductRoutes()
}
