package api

import (
	"net/http"
	"time"

	"github.com/gray509/survy/internal/auth"
)

// "POST /v0/refresh"
func (cfg *apiConfig) Refresh(w http.ResponseWriter, r *http.Request) {
	client_refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		resWithErr(w, http.StatusBadRequest, "could get refresh token", err)
		return
	}

	token_hash, err := auth.Hash(client_refresh_token)
	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), token_hash)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "error retrieving refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "error couldn't make access token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, accessToken)
}

// "POST /v0/revoke"
func (cfg *apiConfig) Revoke(w http.ResponseWriter, r *http.Request) {

	client_refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		resWithErr(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}
	token_hash, err := auth.Hash(client_refresh_token)
	err = cfg.db.RevokeRefreshToken(r.Context(), token_hash)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't revoke session", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
