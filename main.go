package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gray509/polls/api"
	"github.com/gray509/polls/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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
	apicfg := api.NewConfig(database.New(db), port, platform, jwtSecret)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v0/signup", apicfg.CreateUser)
	mux.HandleFunc("POST /v0/login", apicfg.Login)
	mux.HandleFunc("POST /v0/reset", apicfg.Reset)
	mux.HandleFunc("POST /v0/poll", apicfg.CreatePoll)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on: http://localhost:%s/\n", port)
	log.Fatal(srv.ListenAndServe())
}
