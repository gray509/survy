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

// "POST /v0/login"
func (cfg *apiConfig) Login(w http.ResponseWriter, r *http.Request) {
	type r_email_pass struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	emailPass := r_email_pass{}
	err := decoder.Decode(&emailPass)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't decode rquest Json", err)
		return
	}

	//GET USER FROM DB
	user, err := cfg.db.GetUserByEmail(r.Context(), emailPass.Email)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	match, err := auth.CheckPasswordHash(emailPass.Password, user.Password)
	if err != nil || !match {
		resWithErr(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Could't make access token", err)
	}

	refreshToken, hash, create_at, expires_at, err := auth.MakeRefreshToken()
	err = cfg.db.AddRefreshToken(r.Context(), database.AddRefreshTokenParams{
		ID:        uuid.New(),
		TokenHash: hash,
		UserID:    user.ID,
		CreatedAt: querieutils.Time(create_at),
		ExpiresAt: querieutils.Time(expires_at),
	})

	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Could't save refresh token", err)
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt.Time,
			UpdatedAt: user.UpdatedAt.Time,
			Email:     user.Email,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
