package dummy

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/survy/internal/auth"
	"github.com/gray509/survy/internal/database"
	"github.com/gray509/survy/internal/querieutils"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func getJsonTest() ([]byte, error) {
	data, err := os.ReadFile("../test.json")
	if err != nil {
		return nil, err
	}
	return data, err
}

func GetDbConn() (*pgx.Conn, error) {
	godotenv.Load("../.env")
	dbURL := os.Getenv("DB_URL")
	db, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	return db, nil
}

// user pass and email are the same
func CreateUsers(qtx *database.Queries, count int, t *testing.T) ([]database.BulkCreateUserParams, error) {
	var user []database.BulkCreateUserParams
	now := time.Now()
	timetz := querieutils.Time(&now)

	for i := 0; i < count; i++ {
		email := fmt.Sprintf("user%d@testsurvy.com", i)
		pass, err := auth.Hash(email)
		if err != nil {
			return nil, err
		}
		user = append(user, database.BulkCreateUserParams{
			ID:        uuid.New(),
			CreatedAt: timetz,
			UpdatedAt: timetz,
			Email:     email,
			Password:  pass,
		})
	}
	_, err := qtx.BulkCreateUser(t.Context(), user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
