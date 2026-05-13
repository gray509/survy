package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/gray509/survy/server/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QuestionTypes string
type QuestionsMap map[string]interface{}

const (
	Checkbox QuestionTypes = "checkbox"
	Radio    QuestionTypes = "radio"
	Rating   QuestionTypes = "rating"
	YesNo    QuestionTypes = "yes/no"
	Ranking  QuestionTypes = "ranking"
	OpenText QuestionTypes = "open"
)

type apiConfig struct {
	db        *pgxpool.Pool
	port      string
	platform  string
	jwtSecret string
	q         *database.Queries
}

func NewConfig(db *pgxpool.Pool, port, platform, jwtSecret string) *apiConfig {
	return &apiConfig{
		db:        db,
		port:      port,
		platform:  platform,
		jwtSecret: jwtSecret,
		q:         database.New(db),
	}
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Survey struct {
	Id             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Title          string    `json:"title"`
	ExpirationTime time.Time `json:"expiration_time"`
	Identified     bool      `json:"identified"`
	MaxResponse    int       `json:"max_response"`
}
