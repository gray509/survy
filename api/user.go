package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/survy/internal/auth"
	"github.com/gray509/survy/internal/database"
	"github.com/gray509/survy/internal/querieutils"
)

// "POST /v0/signup"
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
	timestamptz := querieutils.Time(&now)
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: timestamptz,
		UpdatedAt: timestamptz,
		Email:     emailPass.Email,
		Password:  hashPass})

	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't save user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        user.ID,
		CreatedAt: now,
		UpdatedAt: now,
		Email:     user.Email,
	})

}
