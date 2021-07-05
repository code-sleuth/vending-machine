package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/code-sleuth/vending-machine/controllers"
	"github.com/code-sleuth/vending-machine/db"
	"github.com/code-sleuth/vending-machine/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

func main() {
	// instantiate multiplexer/router
	mux := mux.NewRouter()

	// initialize database
	database := initDB()

	// initialize db service
	dbService := db.New(database)

	// initialize handlerService
	handlerService := handlers.New(dbService)

	// initialize cache (redis)
	handlers.InitCache()

	// register routes
	controllerService := controllers.New(handlerService, mux)
	controllerService.StartUp()

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

// initDB function
func initDB() *sqlx.DB {
	dbURL := os.Getenv("DB_URL")
	dbConn, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Panicf("unable to connect to database: %v", err)
	}
	dbSchema := loadDBScript()
	dbConn.MustExec(dbSchema)

	dbConn.SetConnMaxLifetime(60 * time.Second)
	log.Println("database initialized successfully")
	return dbConn
}

func loadDBScript() string {
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	file := fmt.Sprintf("%s/%s", currentDirectory, "db/sql/init_schema.sql")
	fileBites, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	return string(fileBites)
}
