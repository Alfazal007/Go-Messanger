package main

import (
	"database/sql"
	"log"
	"messager/controllers"
	"messager/internal/database"
	"messager/routes"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("There was an error loading the environment variables")
	}
	// chi router env variables cors
	portNumber := os.Getenv("PORT")
	if portNumber == "" {
		log.Fatal("Error loading the dot env file")
	}

	// get the database credentiantials
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB url is not found in env variables")
	}
	conn, err := sql.Open("postgres", dbURL)
	apiCfg := controllers.ApiConfig{DB: database.New(conn)}
	if err != nil {
		log.Fatal("Cannot connect to the database", err)
	}
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Mount("/user", routes.UserRouter(&apiCfg))
	srv := &http.Server{
		Addr:    ":" + portNumber,
		Handler: r,
	}
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal("There was an error with the server", err)
	}
}
