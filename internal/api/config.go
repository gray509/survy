package api

import "github.com/gray509/polls/internal/database"

const (
	MultiChoice  = "multi-choice"
	SingleChoice = "single-choice"
	Rating       = "rating"
	YesNo        = "yes/no"
	Ranking      = "ranking"
	OpenText     = "open"
)

type apiConfig struct {
	db        *database.Queries
	port      string
	platform  string
	jwtSecret string
}

func NewConfig(db *database.Queries, port, platform, jwtSecret string) *apiConfig {
	return &apiConfig{
		db:        db,
		port:      port,
		platform:  platform,
		jwtSecret: jwtSecret,
	}
}
