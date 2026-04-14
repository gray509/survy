package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
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
	} `json:"client_create_Survey"`
}

type r_email_pass struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type login_response struct {
	User
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
type responseSurveyId struct {
	Surveyid uuid.UUID `json:"survey_id"`
}

func getJsonTest() ([]byte, error) {
	data, err := os.ReadFile("../test.json")
	if err != nil {
		return nil, err
	}
	return data, err
}

// func getQueries() (*database.Queries, error) {
// 	godotenv.Load(".env.test")
// 	dbURL := os.Getenv("DB_URL")
// 	db, err := pgx.Connect(context.Background(), dbURL)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return database.New(db), nil
// }

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
func sendRequest(req *http.Request) (*http.Response, int, []byte, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, nil, err
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, nil, err
	}
	resp.Body.Close()

	return resp, resp.StatusCode, respBody, nil
}
func TestUserFlow(t *testing.T) {
	user := User{}
	loginResp := login_response{}
	clientUserCreateRequest := r_email_pass{
		Email:    "test@test.com",
		Password: "pass",
	}
	surveyID := responseSurveyId{}
	data, err := getJsonTest()
	clientCreateSurveyRequest := testJson{}
	if err = json.Unmarshal(data, &clientCreateSurveyRequest); err != nil {
		t.Fatal(err)
	}

	// reset the db
	t.Run("reset the db", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/admin/reset", nil)
		if err != nil {
			t.Fatal(err)
		}
		_, respStatusCode, respBody, err := sendRequest(req)
		if err != nil {
			t.Fatal(err)
		}
		checkingExpectedStatusCode(respStatusCode, respBody, http.StatusOK, t)
	})

	// create the user
	t.Run("create the user", func(t *testing.T) {
		body, err := json.Marshal(clientUserCreateRequest)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v0/signup", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}
		_, respStatusCode, respBody, err := sendRequest(req)
		if err != nil {
			t.Fatal(err)
		}
		checkingExpectedStatusCode(respStatusCode, respBody, http.StatusCreated, t)

		if err = json.Unmarshal(respBody, &user); err != nil {
			t.Fatal(err)
		}

		if user.Email != clientUserCreateRequest.Email {
			t.Fatal("user email does not match")
		}
	})

	// login with bad pass
	t.Run("login with bad pass", func(t *testing.T) {
		clientUserCreateRequest.Password = "bad"
		body, err := json.Marshal(clientUserCreateRequest)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v0/login", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}
		_, respStatusCode, respBody, err := sendRequest(req)
		if err != nil {
			t.Fatal(err)
		}
		checkingExpectedStatusCode(respStatusCode, respBody, http.StatusUnauthorized, t)
	})

	// login with bad email
	t.Run("login with bad email", func(t *testing.T) {
		clientUserCreateRequest.Email = "bad@test.com"
		clientUserCreateRequest.Password = "pass"
		body, err := json.Marshal(clientUserCreateRequest)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v0/login", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}
		_, respStatusCode, respBody, err := sendRequest(req)
		if err != nil {
			t.Fatal(err)
		}
		checkingExpectedStatusCode(respStatusCode, respBody, http.StatusUnauthorized, t)
	})

	// login
	t.Run("login", func(t *testing.T) {
		clientUserCreateRequest.Email = "test@test.com"
		clientUserCreateRequest.Password = "pass"
		body, err := json.Marshal(clientUserCreateRequest)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v0/login", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}
		_, respStatusCode, respBody, err := sendRequest(req)
		if err != nil {
			t.Fatal(err)
		}
		checkingExpectedStatusCode(respStatusCode, respBody, http.StatusOK, t)

		if err = json.Unmarshal(respBody, &loginResp); err != nil {
			t.Fatal(err)
		}
		if loginResp.User.ID != user.ID {
			t.Fatal("login in the wrong user account")
		}
		if loginResp.AccessToken == "" || loginResp.RefreshToken == "" {
			t.Fatal("something wrong with auth tokens")
		}
	})

	// create survey with no access token
	t.Run("create survey w/no access token", func(t *testing.T) {
		body, err := json.Marshal(clientCreateSurveyRequest.ClientCreateSurvey)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v0/survey", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}
		_, respStatusCode, respBody, err := sendRequest(req)
		if err != nil {
			t.Fatal(err)
		}
		checkingExpectedStatusCode(respStatusCode, respBody, http.StatusUnauthorized, t)
	})

	//create survey
	t.Run("create survey", func(t *testing.T) {
		body, err := json.Marshal(clientCreateSurveyRequest.ClientCreateSurvey)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v0/survey", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", loginResp.AccessToken))
		_, respStatusCode, respBody, err := sendRequest(req)
		if err != nil {
			t.Fatal(err)
		}
		checkingExpectedStatusCode(respStatusCode, respBody, http.StatusOK, t)
		if err = json.Unmarshal(respBody, &surveyID); err != nil {
			t.Fatal(err)
		}
	})

	//serve survey
	t.Run("serve survey", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:8080/v0/survey/%s", surveyID.Surveyid), nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", loginResp.AccessToken))
		_, respStatusCode, respBody, err := sendRequest(req)
		if err != nil {
			t.Fatal(err)
		}
		checkingExpectedStatusCode(respStatusCode, respBody, http.StatusOK, t)
	})
}
