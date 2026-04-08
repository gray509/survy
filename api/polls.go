package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gray509/polls/internal/database"
)

func (cfg *apiConfig) CreatePoll(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		UserID string `json:"user_id"`
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
	config, err := json.Marshal(params.Config)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't marshall config", err)
		return
	}
	userId, err := uuid.Parse(params.UserID)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't parse uuid", err)
		return
	}
	pollId, err := cfg.db.CreatePoll(r.Context(), database.CreatePollParams{
		Title:  params.Title,
		UserID: userId,
		Config: config,
	})

	for _, q := range params.Questions {
		var options *json.RawMessage
		switch q.Types {
		case MultiChoice, SingleChoice, Rating, Ranking:
			rawJson, err := json.Marshal(q.Options)
			options = (*json.RawMessage)(&rawJson)
			if err != nil {
				resWithErr(w, http.StatusInternalServerError, "Couldn't marshal options", err)
				return
			}
		case YesNo, OpenText:
			// keeps options nil
		default:
			resWithErr(w, http.StatusBadRequest, "Couldn't recognize question type", fmt.Errorf("unknown question type: %v", q.Types))
		}
		_, err = cfg.db.CreateQuestion(r.Context(), database.CreateQuestionParams{
			Title:      q.Title,
			Types:      q.Types,
			IsRequired: q.Required,
			PollsID:    pollId,
			Options:    options,
		})
		if err != nil {
			resWithErr(w, http.StatusInternalServerError, "Couldn't create quetions table", err)
			return
		}
	}

	respondWithJSON(w, http.StatusOK, response{
		Pollid: pollId.String(),
	})
}
