package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/survy/dummy"
	"github.com/gray509/survy/internal/database"
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
type resp_login struct {
	User struct {
		Id        uuid.UUID `json:"id"`
		CreateAt  time.Time `json:"create_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	} `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
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

		t.Fatalf("expected status %d, got %d, with err message: %s", expectedCode, statusCode, respErr.Error)
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
func TestUserCreationFlow(t *testing.T) {
	clientUserCreateRequest := r_email_pass{
		Email:    "test@test.com",
		Password: "pass",
	}
	user := User{}
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

	// create the user again
	t.Run("create the user again", func(t *testing.T) {
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
		checkingExpectedStatusCode(respStatusCode, respBody, http.StatusUnauthorized, t)
	})

}

func TestLoginFlow(t *testing.T) {
	db, err := dummy.GetDbConn()
	if err != nil {
		t.Fatal()
	}
	// tx, err := db.Begin(t.Context())
	// if err != nil {
	// 	t.Fatal()
	// }
	qtx := database.New(db) //.WithTx(tx)
	users, err := dummy.CreateUsers(qtx, 1, t)
	//defer tx.Rollback(t.Context())

	type testCases struct {
		title      string
		statusCode int
		login      r_email_pass
		getResult  bool
	}

	runCases := []testCases{
		{"login with bad pass", http.StatusUnauthorized, r_email_pass{Email: users[0].Email, Password: "bad"}, false},
		//{"login with bad email", http.StatusUnauthorized, r_email_pass{Email: "bad", Password: users[0].Password}, false},
		//{"login", http.StatusOK, r_email_pass{Email: users[0].Email, Password: users[0].Password}, true},
	}
	for _, tt := range runCases {
		t.Run(tt.title, func(t *testing.T) {
			body, err := json.Marshal(tt.login)
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
			checkingExpectedStatusCode(respStatusCode, respBody, tt.statusCode, t)
			if tt.getResult {
				var loginResp resp_login
				if err = json.Unmarshal(respBody, &loginResp); err != nil {
					t.Fatal(err)
				}
				if tt.login.Email != loginResp.User.Email {
					t.Fatal("user email does not match")
				}
				if loginResp.AccessToken == "" {
					t.Fatal("nil access token")
				}
				if loginResp.RefreshToken == "" {
					t.Fatal("nil refresh token")
				}
				if loginResp.User.Id != users[0].ID {
					t.Fatal("uuid ids does not match")
				}
			}
		})
	}
}
