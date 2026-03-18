package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gray509/polls/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	db        *database.Queries
	port      string
	platform  string
	jwtSecret string
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	apicfg := &apiConfig{
		db:        database.New(db),
		port:      port,
		platform:  platform,
		jwtSecret: jwtSecret,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v0/signup", apicfg.createUser)
	mux.HandleFunc("POST /v0/login", apicfg.login)
	mux.HandleFunc("POST /v0/reset", apicfg.reset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on: http://localhost:%s/app/\n", port)
	log.Fatal(srv.ListenAndServe())
}
