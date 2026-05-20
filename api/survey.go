package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
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
			Title        string   `json:"title"`
			QuestionType string   `json:"question_type"`
			IsRequired   bool     `json:"required"`
			Choices      []string `json:"choices,omitempty"`
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
	now := time.Now()
	timestamptz := querieutils.Time(&now)
	surveyId := uuid.New()
	questions := make([]database.BulkEnterQuestionsParams, 0)
	for _, q := range params.Questions {
		if q.QuestionType != Checkbox && q.QuestionType != Radio && q.QuestionType != Rating && q.QuestionType != Ranking && q.QuestionType != OpenText {
			resWithErr(w, http.StatusBadRequest, "Unreconized question type // POST /v0/survey", fmt.Errorf("Unknow question type %s", q.QuestionType))
			return
		}
		questions = append(questions, database.BulkEnterQuestionsParams{
			ID:           uuid.New(),
			CreatedAt:    timestamptz,
			UpdatedAt:    timestamptz,
			Title:        q.Title,
			QuestionType: string(q.QuestionType),
			Choice:       q.Choices,
			SurveyID:     surveyId,
		})
	}

	_, err = cfg.q.CreateSurvey(r.Context(), database.CreateSurveyParams{
		ID:             surveyId,
		CreatedAt:      timestamptz,
		UpdatedAt:      timestamptz,
		Title:          params.Title,
		UserID:         userId,
		ExpirationTime: querieutils.Time(&params.ExpirationTime),
		MaxResponse:    querieutils.Int4(&params.MaxResponse),
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
func (cfg *apiConfig) GetSurveyByID(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Survey    Survey      `json:"survey"`
		Questions []Questions `json:"Questions"`
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

	dbquestions, err := cfg.q.GetQuestionsFromSurveyId(r.Context(), survey.ID)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error retrieving questions // GET /v0/survey/{surveyId}", err)
		return
	}
	questions := make([]Questions, 0)
	for _, q := range dbquestions {
		questions = append(questions, Questions{
			QuestionId:   q.ID,
			CreatedAt:    q.CreatedAt.Time,
			UpdatedAt:    q.UpdatedAt.Time,
			Title:        q.Title,
			QuestionType: q.QuestionType,
			IsRequired:   q.IsRequired,
			Choice:       q.Choice,
		})
	}
	respondWithJSON(w, http.StatusOK, response{
		Survey: Survey{
			Id:             survey.ID,
			CreatedAt:      survey.CreatedAt.Time,
			UpdatedAt:      survey.ExpirationTime.Time,
			Title:          survey.Title,
			ExpirationTime: survey.ExpirationTime.Time,
			MaxResponse:    int(survey.MaxResponse.Int32),
		},
		Questions: questions,
	})
}

// GET /v0/surveys/
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
	type response struct {
		Survey    Survey      `json:"survey"`
		Questions []Questions `json:"Questions"`
		token     string
	}
	surveyId, err := uuid.Parse(r.PathValue("surveyId"))
	if err != nil {
		resWithErr(w, http.StatusBadRequest, "couldn't parse uuid", err)
		return
	}
	survey, err := cfg.q.GetSurveyByIdIsPublish(r.Context(), surveyId)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "couldn't get survey  // GET /v0/survey/{surveyId}/serve", err)
		return
	}
	voterId := uuid.New()
	jwt, err := auth.MakeJWT(voterId, cfg.jwtSecret, time.Minute*10)

	dbquestions, err := cfg.q.GetQuestionsFromSurveyId(r.Context(), survey.ID)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error retrieving questions // GET /v0/survey/{surveyId}/serve", err)
		return
	}
	questions := make([]Questions, 0)
	for _, q := range dbquestions {
		questions = append(questions, Questions{
			QuestionId:   q.ID,
			CreatedAt:    q.CreatedAt.Time,
			UpdatedAt:    q.UpdatedAt.Time,
			Title:        q.Title,
			QuestionType: q.QuestionType,
			IsRequired:   q.IsRequired,
			Choice:       q.Choice,
		})
	}
	respondWithJSON(w, http.StatusOK, response{
		Survey: Survey{
			Id:             survey.ID,
			CreatedAt:      survey.CreatedAt.Time,
			UpdatedAt:      survey.ExpirationTime.Time,
			Title:          survey.Title,
			ExpirationTime: survey.ExpirationTime.Time,
			MaxResponse:    int(survey.MaxResponse.Int32),
		},
		Questions: questions,
		token:     jwt,
	})
}

// POST /v0/survey/{surveyId}/collect
func (cfg *apiConfig) CollectSurvey(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	voterId, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Error validating token // POST /v0/survey/{surveyId}/collect", err)
		return
	}

	// josn processing
	data := make([]byte, 0)
	var requestAnswers map[string]interface{}
	dbAnswers := make(map[string]interface{})

	_, err = r.Body.Read(data)
	if err != nil {
		resWithErr(w, http.StatusUnauthorized, "Couldn't read request body // POST /v0/survey/{surveyId}/collect", err)
		return
	}
	err = json.Unmarshal(data, &requestAnswers)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't unmarshall request body // POST /v0/survey/{surveyId}/collect", err)
		return
	}

	// id processing for db search
	id, ok := requestAnswers["survey_id"].(string)
	if !ok {
		resWithErr(w, http.StatusBadRequest, "Missing survey_id // POST /v0/survey/{surveyId}/collect", fmt.Errorf("survey_id is miising from request"))
		return
	}
	surveyId, err := uuid.Parse(id)
	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't parse uuid // POST /v0/survey/{surveyId}/collect", err)
		return
	}
	dbQuestions, err := cfg.q.GetQuestionsFromSurveyId(r.Context(), surveyId)

	if err != nil {
		resWithErr(w, http.StatusInternalServerError, "Couldn't get questions from db // POST /v0/survey/{surveyId}/collect", err)
		return
	}

	// validating and processing request (choices)
	for _, dbq := range dbQuestions {
		questionId := dbq.ID.String()

		switch dbq.QuestionType {
		case Checkbox:
			choice, ok := requestAnswers[questionId].([]interface{})
			con, err := IsAnswerRequiredValid(ok, dbq.IsRequired)
			if con {
				continue
			}
			if err != nil {
				resWithErr(w, http.StatusBadRequest, "Answer not valid", err)
				return
			}
			indexes, err := converSliceInterfaceToSliceInt(choice)
			if err != nil {
				resWithErr(w, http.StatusBadRequest, "Couldn't convert to []int", err)
				return
			}
			maxValue := len(dbq.Choice) - 1
			for _, n := range indexes {
				if n < 0 || n > maxValue {
					resWithErr(w, http.StatusBadRequest, "Accessing a non-exitant choice // POST /v0/survey/{surveyId}/collect", fmt.Errorf("Accessing a non-exitant choice"))
					return
				}
			}
			dbAnswers[questionId] = choice
		case Radio:
			choice, ok := requestAnswers[questionId].(int)
			con, err := IsAnswerRequiredValid(ok, dbq.IsRequired)
			if con {
				continue
			}
			if err != nil {
				resWithErr(w, http.StatusBadRequest, "Answer not valid", err)
				return
			}
			maxValue := len(dbq.Choice) - 1
			if choice < 0 || choice > maxValue {
				resWithErr(w, http.StatusBadRequest, "Accessing a non-exitant choice // POST /v0/survey/{surveyId}/collect", fmt.Errorf("Accessing a non-exitant choice"))
				return
			}
			dbAnswers[questionId] = choice
		case Rating:
			choice, ok := requestAnswers[questionId].(int)
			con, err := IsAnswerRequiredValid(ok, dbq.IsRequired)
			if con {
				continue
			}
			if err != nil {
				resWithErr(w, http.StatusBadRequest, "Answer not valid", err)
				return
			}
			maxValue, err := strconv.Atoi(dbq.Choice[1])
			if err != nil {
				resWithErr(w, http.StatusBadRequest, fmt.Sprintf("Error with Rating question value: questionid(%s) // POST /v0/survey/{surveyId}/collect", questionId), err)
				return
			}
			if choice < 0 || choice > maxValue {
				resWithErr(w, http.StatusBadRequest, "Accessing a non-exitant choice // POST /v0/survey/{surveyId}/collect", fmt.Errorf("Accessing a non-exitant choice"))
				return
			}
			dbAnswers[questionId] = choice
		case Ranking:
			choice, ok := requestAnswers[questionId].([]interface{})
			con, err := IsAnswerRequiredValid(ok, dbq.IsRequired)
			if con {
				continue
			}
			if err != nil {
				resWithErr(w, http.StatusBadRequest, "Answer not valid", err)
				return
			}
			if len(choice) != len(dbq.Choice) {
				resWithErr(w, http.StatusBadRequest, "request choice length and the db's dont match // POST /v0/survey/{surveyId}/collect", fmt.Errorf("request and db choice are not consistant"))
				return
			}
			indexes, err := converSliceInterfaceToSliceInt(choice)
			if err != nil {
				resWithErr(w, http.StatusBadRequest, "Couldn't convert to []int", err)
				return
			}
			maxValue := len(dbq.Choice) - 1
			for _, n := range indexes {
				if n < 0 || n > maxValue {
					resWithErr(w, http.StatusBadRequest, "Accessing a non-exitant choice // POST /v0/survey/{surveyId}/collect", fmt.Errorf("Accessing a non-exitant choice"))
					return
				}
			}
			dbAnswers[questionId] = choice
		case OpenText:
			choice, ok := requestAnswers[questionId].(string)
			con, err := IsAnswerRequiredValid(ok, dbq.IsRequired)
			if con {
				continue
			}
			if err != nil {
				resWithErr(w, http.StatusBadRequest, "Answer not valid", err)
				return
			}
			dbAnswers[questionId] = choice
		}
	}
	now := time.Now()
	timestamptz := querieutils.Time(&now)
	resonse, err := json.Marshal(dbAnswers)
	if err != nil {
		resWithErr(w, http.StatusBadRequest, "Couldn't Marsall response", err)
		return
	}
	err = cfg.q.CreateResponse(r.Context(), database.CreateResponseParams{
		ID:        uuid.New(),
		CreatedAt: timestamptz,
		UpdatedAt: timestamptz,
		Response:  resonse,
		SurveyID:  surveyId,
		VoterID:   voterId,
	})
	respondWithJSON(w, http.StatusOK, nil)
}

func IsAnswerRequiredValid(exits, isRequired bool) (bool, error) {
	if exits == false && isRequired == true {
		return false, fmt.Errorf("Missing a required question")
	}
	if exits == false && isRequired == false {
		return true, nil
	}
	return false, nil
}
func converSliceInterfaceToSliceInt(rawNums []interface{}) ([]int, error) {
	nums := make([]int, len(rawNums))
	for i, v := range rawNums {
		f, ok := v.(int)
		if !ok {
			return nil, fmt.Errorf("nums[%d] must be number", i)
		}

		nums[i] = f
	}
	return nums, nil
}
