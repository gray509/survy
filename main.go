package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gray509/survy/api"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")

	db, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	apicfg := api.NewConfig(db, port, platform, jwtSecret)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /admin/reset", apicfg.Reset)

	mux.HandleFunc("POST /v0/login", apicfg.Login)
	mux.HandleFunc("POST /v0/refresh", apicfg.Refresh)
	mux.HandleFunc("POST /v0/revoke", apicfg.Revoke)

	mux.HandleFunc("POST /v0/signup", apicfg.CreateUser)
	mux.HandleFunc("POST /v0/survey", apicfg.CreateSurvey)
	mux.HandleFunc("GET /v0/survey/{surveyId}", apicfg.ServeSurvey)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on: http://localhost:%s/\n", port)
	log.Fatal(srv.ListenAndServe())
}
