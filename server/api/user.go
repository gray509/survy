package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/survy/server/internal/auth"
	"github.com/gray509/survy/server/internal/database"
	"github.com/gray509/survy/server/internal/querieutils"
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
		resWithErr(w, http.StatusInternalServerError, "Couldn't decode request Json // POST /v0/signup", err)
		return
	}
	match, err := cfg.q.UserExist(r.Context(), emailPass.Email)
	if match {
		resWithErr(w, http.StatusUnauthorized, "user already exists // POST /v0/signup", err)
		return
	}
	hashPass, err := auth.Hash(emailPass.Password)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't hash password // POST /v0/signup", err)
		return
	}
	now := time.Now()
	timestamptz := querieutils.Time(&now)
	_, err = cfg.q.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: timestamptz,
		UpdatedAt: timestamptz,
		Email:     emailPass.Email,
		Password:  hashPass})

	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't save user // POST /v0/signup", err)
		return
	}

	//http.SetCookie(w, )
	respondWithJSON(w, http.StatusCreated, nil)

}
