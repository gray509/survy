package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type QuestionTypes string
type Options map[string]interface{}

const (
	MultiChoice  QuestionTypes = "multi-choice"
	SingleChoice QuestionTypes = "single-choice"
	Rating       QuestionTypes = "rating"
	YesNo        QuestionTypes = "yes/no"
	Ranking      QuestionTypes = "ranking"
	OpenText     QuestionTypes = "open"
)

type apiConfig struct {
	db        *pgx.Conn
	port      string
	platform  string
	jwtSecret string
}

func NewConfig(db *pgx.Conn, port, platform, jwtSecret string) *apiConfig {
	return &apiConfig{
		db:        db,
		port:      port,
		platform:  platform,
		jwtSecret: jwtSecret,
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
	Title      string        `json:"title"`
	Types      QuestionTypes `json:"types"`
	IsRequired bool          `json:"is_required"`
	Options    Options       `json:"options"`
}
