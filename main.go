package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	_, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
}
