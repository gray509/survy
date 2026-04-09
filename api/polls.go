package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/polls/internal/auth"
	"github.com/gray509/polls/internal/database"
	"github.com/gray509/polls/internal/querieutils"
)

func (cfg *apiConfig) CreatePoll(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Title  string `json:"title"`
		Config struct {
			ExpirationTime string `json:"expiration_time"`
			Identified     bool   `json:"identified"`
			MaxResponse    int    `json:"max_response"`
		} `json:"config"`
		Questions []struct {
			Title    string `json:"title"`
			Types    string `json:"types"`
			Required bool   `json:"required"`
			Options  struct {
				Answers []string `json:"answers"`
			} `json:"options,omitempty"`
		} `json:"questions"`
	}

	type response struct {
		Pollid string `json:"poll_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// authorization
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error getting header jwt token", err)
		return
	}
	userId, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error validating token", err)
		return
	}

	// proocessing to db
	config, err := json.Marshal(params.Config)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't marshall config", err)
		return
	}
	now := time.Now()
	pollId, err := cfg.db.CreatePoll(r.Context(), database.CreatePollParams{
		ID:        uuid.New(),
		CreatedAt: querieutils.Time(&now),
		UpdatedAt: querieutils.Time(&now),
		Title:     params.Title,
		UserID:    userId,
		Config:    config,
	})
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "couldn't save poll to db", err)
		return
	}
	questions := make([]database.QuestionsBulkInsertParams, 0)
	for _, q := range params.Questions {
		var options *json.RawMessage
		switch q.Types {
		case MultiChoice, SingleChoice, Rating, Ranking:
			rawJson, err := json.Marshal(q.Options)
			if err != nil {
				resWithErr(w, http.StatusInternalServerError, "Couldn't marshal options", err)
				return
			}
			options = (*json.RawMessage)(&rawJson)

		case YesNo, OpenText:
			// keeps options nil
		default:
			resWithErr(w, http.StatusBadRequest, "Couldn't recognize question type", fmt.Errorf("unknown question type: %v", q.Types))
		}
		questions = append(questions, database.QuestionsBulkInsertParams{
			ID:         uuid.New(),
			CreatedAt:  querieutils.Time(&now),
			UpdatedAt:  querieutils.Time(&now),
			Title:      q.Title,
			Types:      q.Types,
			IsRequired: q.Required,
			PollsID:    pollId,
			Options:    options,
		})
	}

	_, err = cfg.db.QuestionsBulkInsert(context.Background(), questions)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "couldn't save questions to db", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Pollid: pollId.String(),
	})
}
