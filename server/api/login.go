package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/survy/server/internal/auth"
	"github.com/gray509/survy/server/internal/database"
	"github.com/gray509/survy/server/internal/querieutils"
)

// "POST /v0/login"
func (cfg *apiConfig) Login(w http.ResponseWriter, r *http.Request) {
	type r_email_pass struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
		AccessToken string `json:"access_token"`
	}

	decoder := json.NewDecoder(r.Body)
	emailPass := r_email_pass{}
	err := decoder.Decode(&emailPass)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't decode request Json // POST /v0/login ", err)
		return
	}
	q := database.New(cfg.db)
	//GET USER FROM DB
	user, err := q.GetUserByEmail(r.Context(), emailPass.Email)
	if err != nil {
		log.Println(emailPass.Email)
		resWithErr(w, http.StatusUnauthorized, "Incorrect email // POST /v0/login", err)
		return
	}

	match, err := auth.CheckPasswordHash(emailPass.Password, user.Password)
	if err != nil || !match {
		resWithErr(w, http.StatusUnauthorized, "Incorrect password // POST /v0/login", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Could't make access token // POST /v0/login", err)
	}

	refreshToken, hash, create_at, expires_at, err := auth.MakeRefreshToken()
	err = q.AddRefreshToken(r.Context(), database.AddRefreshTokenParams{
		ID:        uuid.New(),
		TokenHash: hash,
		UserID:    user.ID,
		CreatedAt: querieutils.Time(create_at),
		ExpiresAt: querieutils.Time(expires_at),
	})

	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Could't save refresh token // POST /v0/login", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   false, // use false only in local dev if needed
		SameSite: http.SameSiteStrictMode,
		Expires:  *expires_at,
	})
	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt.Time,
			UpdatedAt: user.UpdatedAt.Time,
			Email:     user.Email,
		},
		AccessToken: accessToken,
	})
}
