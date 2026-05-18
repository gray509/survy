package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/gray509/survy/server/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	Checkbox string = "checkbox"
	Radio    string = "radio"
	Rating   string = "rating"
	Ranking  string = "ranking"
	OpenText string = "open"
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

type Questions struct {
	QuestionId   uuid.UUID `json:"question_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Title        string    `json:"title"`
	QuestionType string    `json:"types"`
	IsRequired   bool      `json:"required"`
	Choice       []string  `json:"options,omitempty"`
}
