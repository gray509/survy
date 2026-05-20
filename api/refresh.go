package api

import (
	"net/http"
	"time"

	"github.com/gray509/survy/server/internal/auth"
)

// "POST /v0/refresh"
func (cfg *apiConfig) Refresh(w http.ResponseWriter, r *http.Request) {
	type respnse struct {
		AccessToken string `json:"access_token"`
	}
	cook, err := r.Cookie("refresh_token")
	if err != nil {
		resWithErr(w, http.StatusBadRequest, "refresh token cookie missing", err)
		return
	}

	token_hash, err := auth.DeterHash(cook.Value)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "error couldn't verify token", err)
		return
	}
	userId, err := cfg.q.GetUserFromRefreshToken(r.Context(), token_hash)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "error retrieving refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(userId, cfg.jwtSecret, time.Hour)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "error couldn't make access token", err)
		return
	}
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	respondWithJSON(w, http.StatusOK, respnse{
		AccessToken: accessToken,
	})
}

// "POST /v0/revoke"
func (cfg *apiConfig) Revoke(w http.ResponseWriter, r *http.Request) {

	client_refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		resWithErr(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}
	token_hash, err := auth.Hash(client_refresh_token)

	err = cfg.q.RevokeRefreshToken(r.Context(), token_hash)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't revoke session", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
