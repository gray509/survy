package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/polls/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
)

type testJson struct {
	ClientCreatePoll struct {
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
	} `json:"client_create_poll"`
	CreateUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	} `json:"create_user"`
}

func getJsonTest() ([]byte, error) {
	data, err := os.ReadFile("/home/ddori/workspace/github/polls/test.json")
	if err != nil {
		return nil, err
	}
	return data, err
}

func checkingExpectedStatusCode(statusCode int, respBody []byte, expectedCode int, t *testing.T) {
	if statusCode != expectedCode {
		type errorResponse struct {
			Error string `json:"error"`
		}

		var respErr errorResponse
		if err := json.Unmarshal(respBody, &respErr); err != nil {
			t.Fatal(err)
		}

		t.Fatalf("expected status %d, got %d, with err message %s", expectedCode, statusCode, respErr.Error)
	}
}

func createTestUser() (string, error) {
	type createUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	email := "_@test.com"
	pass := "thisapassword"
	jsonBody := createUser{
		Email:    email,
		Password: pass,
	}
	body, err := json.Marshal(jsonBody)
	if err != nil {
		return "", err
	}

	//setting request
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v0/signup", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	//sending resquest
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// processing response
	respBody, err := io.ReadAll(resp.Body)
	type User struct {
		ID        uuid.UUID        `json:"id"`
		CreatedAt pgtype.Timestamp `json:"created_at"`
		UpdatedAt pgtype.Timestamp `json:"updated_at"`
		Email     string           `json:"email"`
	}
	var respUser User
	if err = json.Unmarshal(respBody, &respUser); err != nil {
		return "", err
	}

	return respUser.ID.String(), nil
}

func getQueries() (*database.Queries, error) {
	godotenv.Load(".env.test")
	dbURL := os.Getenv("DB_URL")
	db, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	return database.New(db), nil
}

func TestUserCreation(t *testing.T) {
	type createUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	email := "user@test.com"
	pass := "thisapassword"
	jsonBody := createUser{
		Email:    email,
		Password: pass,
	}
	body, err := json.Marshal(jsonBody)
	if err != nil {
		t.Fatal(err)
	}

	//setting request
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v0/signup", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	//sending resquest
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// processing response
	respBody, err := io.ReadAll(resp.Body)

	checkingExpectedStatusCode(resp.StatusCode, respBody, http.StatusCreated, t)

	type User struct {
		ID        uuid.UUID        `json:"id"`
		CreatedAt pgtype.Timestamp `json:"created_at"`
		UpdatedAt pgtype.Timestamp `json:"updated_at"`
		Email     string           `json:"email"`
	}
	var respUser User

	if err = json.Unmarshal(respBody, &respUser); err != nil {
		t.Fatal(err)
	}
	if respUser.Email != email {
		t.Fail()
	}

}

func TestPollsCreation(t *testing.T) {
	// setting request
	data, err := getJsonTest()
	if err != nil {
		t.Error(err)
	}
	var requestEx testJson
	if err = json.Unmarshal(data, &requestEx); err != nil {
		t.Error(err)
	}
	userId, err := createTestUser()
	if err != nil {
		t.Error(err, errors.New("could created new test user for its uuid"))
	}
	requestEx.ClientCreatePoll.UserID = userId
	body, err := json.Marshal(requestEx.ClientCreatePoll)
	if err != nil {
		t.Error(err)
	}
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v0/poll", bytes.NewBuffer(body))
	if err != nil {
		t.Error(err)
	}

	// processing resonse
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	checkingExpectedStatusCode(resp.StatusCode, respBody, http.StatusOK, t)

	if len(respBody) == 0 {
		t.Fatal("empty response body")
	}

	type response struct {
		Pollid string `json:"poll_id"`
	}
	// unmarshal RESPONSE (not input data)
	var pollid response
	if err := json.Unmarshal(respBody, &pollid); err != nil {
		t.Fatal(err)
	}

	if pollid.Pollid == "" {
		t.Fail()
	}

	q, err := getQueries()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	pollID, err := uuid.Parse(pollid.Pollid)
	if err != nil {
		t.Fatal(err)
	}
	questions, err := q.GetQuestionsWithPollid(ctx, pollID)
	if err != nil {
		t.Fatal(err)
	}
	if len(questions) != 6 {
		t.Fatal()
	}
	for _, question := range questions {
		fmt.Printf("Title: %s\n", question.Title)
		fmt.Printf("Types: %s\n", question.Types)
		fmt.Printf("Title: %t\n", question.IsRequired)
	}

}
