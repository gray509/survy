package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/polls/internal/auth"
	"github.com/gray509/polls/internal/database"
	"github.com/gray509/polls/internal/querieutils"
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID        uuid.UUID          `json:"id"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
	Email     string             `json:"email"`
}

func (cfg *apiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	type r_email_pass struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	emailPass := r_email_pass{}

	err := decoder.Decode(&emailPass)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't decode request Json", err)
		return
	}

	hashPass, err := auth.Hash(emailPass.Password)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}
	now := time.Now()
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: querieutils.Time(&now),
		UpdatedAt: querieutils.Time(&now),
		Email:     emailPass.Email,
		Password:  hashPass})

	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't save user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        user.ID,
		CreatedAt: querieutils.Time(&now),
		UpdatedAt: querieutils.Time(&now),
		Email:     user.Email,
	})

}
