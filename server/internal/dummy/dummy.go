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
	"github.com/gray509/survy/server/internal/auth"
	"github.com/gray509/survy/server/internal/database"
	"github.com/gray509/survy/server/internal/querieutils"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

const (
	Checkbox string = "checkbox"
	Radio    string = "radio"
	Rating   string = "rating"
	Ranking  string = "ranking"
	OpenText string = "open"
)

type testJson struct {
	ClientCreateSurvey struct {
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
	questions := make([]database.BulkEnterQuestionsParams, 0)

	for i, user := range users {
		surveyIds = append(surveyIds, uuid.New())
		surveys = append(surveys, database.BulkCreateSurveyParams{
			ID:             surveyIds[i],
			CreatedAt:      timestamptz,
			UpdatedAt:      timestamptz,
			Title:          testjson.ClientCreateSurvey.Title,
			UserID:         user.ID,
			ExpirationTime: querieutils.Time(&testjson.ClientCreateSurvey.ExpirationTime),
			MaxResponse:    querieutils.Int4(&testjson.ClientCreateSurvey.MaxResponse),
		})

		for _, q := range testjson.ClientCreateSurvey.Questions {
			if q.QuestionType != Checkbox && q.QuestionType != Radio && q.QuestionType != Rating && q.QuestionType != Ranking && q.QuestionType != OpenText {
				t.Fatal(fmt.Errorf("unknown question type, %s", q.QuestionType))
			}
			questions = append(questions, database.BulkEnterQuestionsParams{
				ID:           uuid.New(),
				CreatedAt:    timestamptz,
				UpdatedAt:    timestamptz,
				Title:        q.Title,
				QuestionType: string(q.QuestionType),
				Choice:       q.Choices,
				SurveyID:     surveyIds[i],
			})
		}
	}

	_, err = qtx.BulkCreateSurvey(t.Context(), surveys)
	if err != nil {
		return nil, err
	}

	return surveyIds, nil
}
