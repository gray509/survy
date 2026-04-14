package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/survy/internal/auth"
	"github.com/gray509/survy/internal/database"
	"github.com/gray509/survy/internal/querieutils"
)

// "POST /v0/survey"
func (cfg *apiConfig) CreateSurvey(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Title          string    `json:"title"`
		ExpirationTime time.Time `json:"expiration_time"`
		Identified     bool      `json:"identified"`
		MaxResponse    int       `json:"max_response"`
		Questions      []struct {
			Title      string        `json:"title"`
			Types      QuestionTypes `json:"types"`
			IsRequired bool          `json:"required"`
			Options    struct {
			} `json:"options,omitempty"`
		} `json:"questions"`
	}

	type response struct {
		Surveyid uuid.UUID `json:"survey_id"`
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

	now := time.Now()
	timestamptz := querieutils.Time(&now)
	surveyId, err := cfg.db.CreateSurvey(r.Context(), database.CreateSurveyParams{
		ID:             uuid.New(),
		CreatedAt:      timestamptz,
		UpdatedAt:      timestamptz,
		Title:          params.Title,
		UserID:         userId,
		ExpirationTime: querieutils.Time(&params.ExpirationTime),
		Indentified:    params.Identified,
		MaxResponse:    querieutils.Int4(&params.MaxResponse),
	})
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "couldn't save survey to db", err)
		return
	}
	questions := make([]database.QuestionsBulkInsertParams, 0)
	for _, q := range params.Questions {
		var options *json.RawMessage
		switch q.Types {
		case (MultiChoice), SingleChoice, Rating, Ranking:
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
			CreatedAt:  timestamptz,
			UpdatedAt:  timestamptz,
			Title:      q.Title,
			Types:      string(q.Types),
			IsRequired: q.IsRequired,
			SurveysID:  surveyId,
			Options:    options,
		})
	}

	_, err = cfg.db.QuestionsBulkInsert(context.Background(), questions)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "couldn't save questions to db", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Surveyid: surveyId,
	})
}

// GET /v0/survey/{surveyId}
func (cfg *apiConfig) ServeSurvey(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Survey
		questions []Questions
	}
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
	urlSurveyId, err := uuid.Parse(r.PathValue("surveyId"))
	if err != nil {
		resWithErr(w, http.StatusBadRequest, "err with uuid", err)
		return
	}

	survey, err := cfg.db.GetSurveyByIdUserId(r.Context(), database.GetSurveyByIdUserIdParams{ID: urlSurveyId, UserID: userId})
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error retrieving survey", err)
		return
	}

	questions, err := cfg.db.GetQuestionBySurveyId(r.Context(), survey.ID)
	var responseQuestions []Questions
	var options map[string]interface{}
	for _, q := range questions {
		err = json.Unmarshal(*q.Options, &options)
		if err != nil {
			resWithErr(w, http.StatusInternalServerError, "Couldn't parse options to json", err)
			return
		}
		responseQuestions = append(responseQuestions, Questions{
			Title:      q.Title,
			Types:      QuestionTypes(q.Types),
			IsRequired: q.IsRequired,
			Options:    options,
		})
	}
	respondWithJSON(w, http.StatusOK, response{
		Survey: Survey{
			Id:             survey.ID,
			CreatedAt:      survey.CreatedAt.Time,
			UpdatedAt:      survey.ExpirationTime.Time,
			Title:          survey.Title,
			ExpirationTime: survey.ExpirationTime.Time,
			Identified:     survey.Indentified,
			MaxResponse:    int(survey.MaxResponse.Int32),
		},
	})
}
