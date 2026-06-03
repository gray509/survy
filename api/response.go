package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/survy/server/internal/auth"
	"github.com/gray509/survy/server/internal/database"
	"github.com/gray509/survy/server/internal/querieutils"
)

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

	dbAnswers, err := requestAnwserValidation(dbQuestions, requestAnswers)
	if err != nil {
		resWithErr(w, http.StatusBadRequest, "Answers was not validated // POST /v0/survey/{surveyId}/collect", err)
		return
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

var ErrMissingRequiredQuest = errors.New("Missing a required question")
var ErrMustbeNum = errors.New("error converting inter to int")
var ErrAccessingNonExisting = errors.New("Accessing a non-exitant choice")
var ErrNotConsistantWithDb = errors.New("request and db choices are not consistant")

func isAnswerRequiredValid(exits, isRequired bool) (bool, error) {
	if exits == false && isRequired == true {
		return false, ErrMissingRequiredQuest
	}
	if exits == false && isRequired == false {
		return true, nil
	}
	return false, nil
}
func converSliceInterfaceToSliceInt(rawNums []interface{}) ([]int, error) {
	nums := make([]int, len(rawNums))
	for i, v := range rawNums {
		f, ok := v.(float64)
		if !ok {
			return nil, fmt.Errorf("nums[%d] must be number: %e", i, ErrMissingRequiredQuest)
		}

		nums[i] = int(f)
	}
	return nums, nil
}
func requestAnwserValidation(dbQuestions []database.GetQuestionsFromSurveyIdRow, requestAnswers map[string]interface{}) (map[string]interface{}, error) {
	dbAnswers := make(map[string]interface{})
	for _, dbq := range dbQuestions {
		questionId := dbq.ID.String()
		fmt.Println(dbq.Title)
		switch dbq.QuestionType {
		case Checkbox:
			choice, ok := requestAnswers[questionId].([]interface{})
			con, err := isAnswerRequiredValid(ok, dbq.IsRequired)
			if con {
				continue
			}
			if err != nil {
				return nil, err
			}
			indexes, err := converSliceInterfaceToSliceInt(choice)
			if err != nil {
				return nil, err
			}
			maxValue := len(dbq.Choice) - 1
			for _, n := range indexes {
				if n < 0 || n > maxValue {
					return nil, ErrAccessingNonExisting
				}
			}
			dbAnswers[questionId] = indexes
		case Radio:
			choice, ok := requestAnswers[questionId].(float64)
			con, err := isAnswerRequiredValid(ok, dbq.IsRequired)
			if con {
				continue
			}
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			maxValue := len(dbq.Choice) - 1
			c := int(choice)
			if c < 0 || c > maxValue {
				return nil, ErrAccessingNonExisting
			}
			dbAnswers[questionId] = c
		case Rating:
			choice, ok := requestAnswers[questionId].(float64)
			con, err := isAnswerRequiredValid(ok, dbq.IsRequired)
			if con {
				continue
			}
			if err != nil {
				return nil, err
			}
			maxValue, err := strconv.Atoi(dbq.Choice[1])
			if err != nil {
				return nil, err
			}
			c := int(choice)
			if c < 0 || c > maxValue {
				return nil, ErrAccessingNonExisting
			}
			dbAnswers[questionId] = c
		case Ranking:
			choice, ok := requestAnswers[questionId].([]interface{})
			con, err := isAnswerRequiredValid(ok, dbq.IsRequired)
			if con {
				continue
			}
			if err != nil {
				return nil, err
			}
			if len(choice) != len(dbq.Choice) {
				return nil, ErrNotConsistantWithDb
			}
			indexes, err := converSliceInterfaceToSliceInt(choice)
			if err != nil {
				return nil, err
			}
			maxValue := len(dbq.Choice) - 1
			for _, n := range indexes {
				if n < 0 || n > maxValue {
					return nil, ErrAccessingNonExisting
				}
			}
			dbAnswers[questionId] = indexes
		case OpenText:
			choice, ok := requestAnswers[questionId].(string)
			con, err := isAnswerRequiredValid(ok, dbq.IsRequired)
			if con {
				continue
			}
			if err != nil {
				return nil, err
			}
			dbAnswers[questionId] = choice
		}
	}
	return dbAnswers, nil
}
