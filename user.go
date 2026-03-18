package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/polls/internal/auth"
	"github.com/gray509/polls/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {

	type r_email_pass struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	emailPass := r_email_pass{}

	err := decoder.Decode(&emailPass)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't decode rquest Json", err)
		return
	}

	hashPass, err := auth.HashPassword(emailPass.Password)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{Email: emailPass.Email, Password: hashPass})
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't save user", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})

}
