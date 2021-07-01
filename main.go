package main

import (
	"log"
	"net/http"

	"github.com/code-sleuth/vending-machine/controllers"
	"github.com/code-sleuth/vending-machine/handlers"
	"github.com/code-sleuth/vending-machine/models"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// run db adapter
	models.Adapter()

	// instantiate multiplexer/router
	mux := mux.NewRouter()

	// initialize cache (redis)
	handlers.InitCache()

	// register routes
	controllers.StartUp(mux)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost"},
		AllowCredentials: true,
		Debug:            true,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowedHeaders: []string{"*"},
	})
	handler := c.Handler(mux)

	log.Println("listening on 3333")
	if err := http.ListenAndServe(":3333", handler); err != nil {
		log.Fatal(err)
	}
}
