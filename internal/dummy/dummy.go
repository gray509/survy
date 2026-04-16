package dummy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/survy/internal/auth"
	"github.com/gray509/survy/internal/database"
	"github.com/gray509/survy/internal/querieutils"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type QuestionTypes string
type Options map[string]interface{}

const (
	MultiChoice  QuestionTypes = "multi-choice"
	SingleChoice QuestionTypes = "single-choice"
	Rating       QuestionTypes = "rating"
	YesNo        QuestionTypes = "yes/no"
	Ranking      QuestionTypes = "ranking"
	OpenText     QuestionTypes = "open"
)

type testJson struct {
	ClientCreateSurvey struct {
		Title          string    `json:"title"`
		ExpirationTime time.Time `json:"expiration_time"`
		Identified     bool      `json:"identified"`
		MaxResponse    int       `json:"max_response"`
		Questions      []struct {
			Title      string `json:"title"`
			Types      string `json:"types"`
			IsRequired bool   `json:"required"`
			Options    struct {
				Answers []string `json:"answers"`
			} `json:"options,omitempty"`
		} `json:"questions"`
	} `json:"r_create_Survey"`
}

func GetJsonTest() ([]byte, error) {
	data, err := os.ReadFile("../test.json")
	if err != nil {
		return nil, err
	}
	return data, err
}

func GetDbConn() (*pgx.Conn, error) {
	godotenv.Load("../.env")
	dbURL := os.Getenv("DB_URL")
	db, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	return db, nil
}

// user pass and email are the same
func CreateUsers(qtx *database.Queries, count int, t *testing.T, rawPassword string) ([]database.BulkCreateUserParams, error) {
	var user []database.BulkCreateUserParams
	now := time.Now()
	timetz := querieutils.Time(&now)

	pass, err := auth.Hash(rawPassword)
	if err != nil {
		return nil, err
	}
	for i := 0; i < count; i++ {
		email := fmt.Sprintf("user%d@testsurvy.com", i)
		user = append(user, database.BulkCreateUserParams{
			ID:        uuid.New(),
			CreatedAt: timetz,
			UpdatedAt: timetz,
			Email:     email,
			Password:  pass,
		})
	}
	_, err = qtx.BulkCreateUser(t.Context(), user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func LoginUser(email, pass string) (string, string, error) {
	type r_email_pass struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type resp_login struct {
		Id           uuid.UUID `json:"id"`
		CreateAt     time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
	}
	body, err := json.Marshal(r_email_pass{Email: email, Password: pass})
	if err != nil {
		return "", "", err
	}
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v0/login", bytes.NewBuffer(body))
	if err != nil {
		return "", "", err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var loginResp resp_login
	if err = json.Unmarshal(respBody, &loginResp); err != nil {
		return "", "", err
	}
	return loginResp.AccessToken, loginResp.RefreshToken, nil
}

func CreateSurvey(qtx *database.Queries, users []database.BulkCreateUserParams, t *testing.T) ([]uuid.UUID, error) {
	// json request loaded
	data, err := GetJsonTest()
	if err != nil {
		t.Fatal(err)
	}
	var testjson testJson
	err = json.Unmarshal(data, &testjson)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	timestamptz := querieutils.Time(&now)
	surveyIds := make([]uuid.UUID, 0)
	surveys := make([]database.BulkCreateSurveyParams, 0)
	questions := make([]database.QuestionsBulkInsertParams, 0)
	for i, user := range users {
		surveyIds = append(surveyIds, uuid.New())
		surveys = append(surveys, database.BulkCreateSurveyParams{
			ID:             surveyIds[i],
			CreatedAt:      timestamptz,
			UpdatedAt:      timestamptz,
			Title:          testjson.ClientCreateSurvey.Title,
			UserID:         user.ID,
			ExpirationTime: querieutils.Time(&testjson.ClientCreateSurvey.ExpirationTime),
			Indentified:    testjson.ClientCreateSurvey.Identified,
			MaxResponse:    querieutils.Int4(&testjson.ClientCreateSurvey.MaxResponse),
		})

		for _, q := range testjson.ClientCreateSurvey.Questions {
			var options *json.RawMessage
			switch QuestionTypes(q.Types) {
			case MultiChoice, SingleChoice, Rating, Ranking:
				rawJson, err := json.Marshal(q.Options)
				if err != nil {
					return nil, err
				}
				options = (*json.RawMessage)(&rawJson)

			case YesNo, OpenText:
				// keeps options nil
			default:
				return nil, fmt.Errorf("bad type")
			}
			questions = append(questions, database.QuestionsBulkInsertParams{
				ID:         uuid.New(),
				CreatedAt:  timestamptz,
				UpdatedAt:  timestamptz,
				Title:      q.Title,
				Types:      string(q.Types),
				IsRequired: q.IsRequired,
				SurveysID:  surveyIds[i],
				Options:    options,
			})
		}
	}
	_, err = qtx.BulkCreateSurvey(t.Context(), surveys)
	if err != nil {
		return nil, err
	}

	_, err = qtx.QuestionsBulkInsert(t.Context(), questions)
	if err != nil {
		return nil, err
	}
	return surveyIds, nil
}
