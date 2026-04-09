package api

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

	"github.com/gray509/polls/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type testJson struct {
	ClientCreatePoll struct {
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

type r_email_pass struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type login_response struct {
	User
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func getJsonTest() ([]byte, error) {
	data, err := os.ReadFile("/home/ddori/workspace/github/polls/test.json")
	if err != nil {
		return nil, err
	}
	return data, err
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

	// create poll with no access token
	t.Run("create poll w/no access token", func(t *testing.T) {
		data, err := getJsonTest()
		clientCreatePollRequest := testJson{}
		if err = json.Unmarshal(data, &clientCreatePollRequest); err != nil {
			t.Fatal(err)
		}
		body, err := json.Marshal(clientCreatePollRequest.ClientCreatePoll)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v0/poll", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}
		_, respStatusCode, respBody, err := sendRequest(req)
		if err != nil {
			t.Fatal(err)
		}
		checkingExpectedStatusCode(respStatusCode, respBody, http.StatusUnauthorized, t)
	})

	//create poll
	t.Run("create poll", func(t *testing.T) {
		data, err := getJsonTest()
		clientCreatePollRequest := testJson{}
		if err = json.Unmarshal(data, &clientCreatePollRequest); err != nil {
			t.Fatal(err)
		}
		body, err := json.Marshal(clientCreatePollRequest.ClientCreatePoll)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v0/poll", bytes.NewBuffer(body))
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
