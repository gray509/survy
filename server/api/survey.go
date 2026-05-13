package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/survy/server/internal/auth"
	"github.com/gray509/survy/server/internal/database"
	"github.com/gray509/survy/server/internal/querieutils"
)

// "POST /v0/survey"
func (cfg *apiConfig) CreateSurvey(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Title          string    `json:"title"`
		ExpirationTime time.Time `json:"expiration_time"`
		Identified     bool      `json:"identified"`
		MaxResponse    int       `json:"max_response"`
		Questions      []struct {
			QuestionId uuid.UUID     `json:"question_id,omitempty"`
			Title      string        `json:"title"`
			Types      QuestionTypes `json:"types"`
			IsRequired bool          `json:"required"`
			Options    []string      `json:"options,omitempty"`
		} `json:"questions"`
	}

	type response struct {
		Surveyid uuid.UUID `json:"survey_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't decode parameters // POST /v0/survey", err)
		return
	}

	// authorization
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error getting header jwt token // POST /v0/survey", err)
		return
	}
	userId, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error validating token // POST /v0/survey", err)
		return
	}
	if len(params.Questions) < 1 {
		resWithErr(w, http.StatusBadRequest, "Empty question list // POST /v0/survey", fmt.Errorf("Empty question list"))
		return
	}

	// proocessing to db
	for i := range params.Questions {
		q := &params.Questions[i]
		if q.Types != Checkbox && q.Types != Radio && q.Types != Rating && q.Types != YesNo && q.Types != Ranking && q.Types != OpenText {
			resWithErr(w, http.StatusBadRequest, "Unreconized question type // POST /v0/survey", fmt.Errorf("Unknow question type %s", q.Types))
			return
		}
		q.QuestionId = uuid.New()
	}
	questionsJson, err := json.Marshal(params.Questions)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't marshal questions // POST /v0/survey", err)
		return
	}
	now := time.Now()
	timestamptz := querieutils.Time(&now)

	surveyId, err := cfg.q.CreateSurvey(r.Context(), database.CreateSurveyParams{
		ID:             uuid.New(),
		CreatedAt:      timestamptz,
		UpdatedAt:      timestamptz,
		Title:          params.Title,
		UserID:         userId,
		ExpirationTime: querieutils.Time(&params.ExpirationTime),
		MaxResponse:    querieutils.Int4(&params.MaxResponse),
		Questions:      (json.RawMessage)(questionsJson),
	})
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "couldn't save survey to db // POST /v0/survey", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Surveyid: surveyId,
	})
}

// GET /v0/survey/{surveyId}
func (cfg *apiConfig) GetSurvey(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Survey
		questions QuestionsMap
	}
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error getting header jwt token // GET /v0/survey/{surveyId}", err)
		return
	}
	userId, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error validating token // GET /v0/survey/{surveyId}", err)
		return
	}
	urlSurveyId, err := uuid.Parse(r.PathValue("surveyId"))
	if err != nil {
		resWithErr(w, http.StatusNotFound, "err with uuid // GET /v0/survey/{surveyId}", err)
		return
	}

	survey, err := cfg.q.GetSurveyByIdUserId(r.Context(), database.GetSurveyByIdUserIdParams{ID: urlSurveyId, UserID: userId})
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error retrieving survey // GET /v0/survey/{surveyId}", err)
		return
	}

	var questions QuestionsMap
	err = json.Unmarshal(survey.Questions, &questions)

	respondWithJSON(w, http.StatusOK, response{
		Survey: Survey{
			Id:             survey.ID,
			CreatedAt:      survey.CreatedAt.Time,
			UpdatedAt:      survey.ExpirationTime.Time,
			Title:          survey.Title,
			ExpirationTime: survey.ExpirationTime.Time,
			MaxResponse:    int(survey.MaxResponse.Int32),
		},
		questions: questions,
	})
}

// GET /v0/survey/
func (cfg *apiConfig) GetUserSurveys(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	sortOrder := r.URL.Query().Get("sort")
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error getting header jwt token // GET /v0/survey/", err)
		return
	}
	userId, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error validating token // GET /v0/survey", err)
		return
	}

	surveys, err := cfg.q.GetAllUserSurveys(r.Context(), userId)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error getting surveys // GET GET /v0/survey/", err)
		return
	}

	response := make([]Survey, 0)
	for _, s := range surveys {
		response = append(response, Survey{
			Id:             s.ID,
			CreatedAt:      s.CreatedAt.Time,
			UpdatedAt:      s.UpdatedAt.Time,
			Title:          s.Title,
			ExpirationTime: s.ExpirationTime.Time,
			MaxResponse:    int(s.MaxResponse.Int32),
		})
	}
	if sortOrder == "desc" {
		sort.Slice(response, func(i, j int) bool { return response[i].CreatedAt.After(response[j].UpdatedAt) })
	}
	respondWithJSON(w, http.StatusOK, response)
}

// POST /v0/publish/
func (cfg *apiConfig) PublishSurvey(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		SurveyId uuid.UUID `json:"survey_id"`
		Enable   bool      `json:"enable"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't decode parameters // POST /v0/publish/", err)
		return
	}

	// authorization
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error getting header jwt token // POST /v0/publish/", err)
		return
	}
	userId, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error validating token // POST /v0/publish/", err)
		return
	}
	log.Println(userId.String(), "//publish")

	row, err := cfg.q.SetIsPublish(r.Context(), database.SetIsPublishParams{UserID: userId, ID: params.SurveyId, IsPublished: params.Enable})
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Error with db conn  // POST /v0/publish/", err)
		return
	}
	if row.RowsAffected() == 0 {
		resWithErr(w, http.StatusUnauthorized, "Error setting publish flag  // POST /v0/publish/", fmt.Errorf("no rows were affected"))
		return
	}

	surveyUrl := ""
	if params.Enable == true {
		surveyUrl = fmt.Sprintf("http://localhost:8080/v0/%s/serve", params.SurveyId.String())
	}
	type response struct {
		SurveyUrl string `json:"survey_url"`
	}
	respondWithJSON(w, http.StatusOK, response{SurveyUrl: surveyUrl})
}

// GET /v0/survey/{surveyId}/serve
func (cfg *apiConfig) ServeSurvey(w http.ResponseWriter, r *http.Request) {
	surveyId, err := uuid.Parse(r.PathValue("surveyId"))
	if err != nil {
		resWithErr(w, http.StatusBadRequest, "couldn't parse uuid", err)
		return
	}

	voterId := uuid.New()
	expires := time.Now().Add(time.Minute * 10)
	http.SetCookie(w, &http.Cookie{
		Name:     "voter_id",
		Value:    voterId.String(),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  expires,
	})

}
