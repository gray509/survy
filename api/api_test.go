package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/survy/server/internal/database"
	"github.com/gray509/survy/server/internal/dummy"
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
func TestUserCreation(t *testing.T) {
	type r_email_pass struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	clientUserCreateRequest := r_email_pass{
		Email:    "user-1@testsurvy.com",
		Password: "pass",
	}

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
	type r_email_pass struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type resp_login struct {
		Id          uuid.UUID `json:"id"`
		CreateAt    time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		AccessToken string    `json:"access_token"`
	}

	db, err := dummy.GetDbConn()
	if err != nil {
		t.Fatal()
	}
	qtx := database.New(db)
	defer qtx.DeleteTestUsers(t.Context())
	usersPassword := "pass"
	users, err := dummy.CreateUsers(qtx, 1, t, usersPassword)
	if err != nil {
		t.Fatal(err)
	}
	type testCases struct {
		title      string
		statusCode int
		login      r_email_pass
		getResult  bool
	}

	runCases := []testCases{
		{"login with bad pass", http.StatusUnauthorized, r_email_pass{Email: users[0].Email, Password: "bad"}, false},
		{"login with bad email", http.StatusUnauthorized, r_email_pass{Email: "bad", Password: usersPassword}, false},
		{"login", http.StatusOK, r_email_pass{Email: users[0].Email, Password: usersPassword}, true},
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
				if tt.login.Email != loginResp.Email {
					t.Fatalf("user email does not match, test_Email:%s != response_Email:%v", tt.login.Email, loginResp.Email)
				}
				if loginResp.AccessToken == "" {
					t.Fatal("nil access token")
				}
				if loginResp.Id != users[0].ID {
					t.Fatal("uuid ids does not match")
				}
			}
		})
	}
}

func TestSurveyCreation(t *testing.T) {
	//db conn
	db, err := dummy.GetDbConn()
	if err != nil {
		t.Fatal()
	}
	qtx := database.New(db)
	defer qtx.DeleteTestUsers(t.Context())
	//users created
	usersPassword := "pass"
	users, err := dummy.CreateUsers(qtx, 1, t, usersPassword)
	if err != nil {
		t.Fatal(err)
	}
	// json request loaded
	data, err := dummy.GetJsonTest()
	if err != nil {
		t.Fatal(err)
	}
	var testjson testJson
	err = json.Unmarshal(data, &testjson)
	if err != nil {
		t.Fatal(err)
	}
	//login user
	accessToken, _, err := dummy.LoginUser(users[0].Email, usersPassword)
	if err != nil {
		t.Fatal(err)
	}
	type testCases struct {
		title          string
		statusCode     int
		setAccessToken bool
		badTypes       bool
		setNoQuestions bool
	}
	runCases := []testCases{
		{"create survey w/no access token", http.StatusUnauthorized, false, false, false},
		{"create survey", http.StatusOK, true, false, false},
		{"create survey bad types", http.StatusBadRequest, true, true, false},
		{"create survey with no questions", http.StatusBadRequest, true, false, true},
	}

	for _, tt := range runCases {
		t.Run(tt.title, func(t *testing.T) {
			if tt.badTypes {
				testjson.ClientCreateSurvey.Questions[0].QuestionType = "bad"
			}
			if tt.setNoQuestions {
				testjson.ClientCreateSurvey.Questions = nil
			}
			body, err := json.Marshal(testjson.ClientCreateSurvey)
			if err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v0/survey", bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			if tt.setAccessToken {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
			}
			_, respStatusCode, respBody, err := sendRequest(req)
			if err != nil {
				t.Fatal(err)
			}
			checkingExpectedStatusCode(respStatusCode, respBody, tt.statusCode, t)

		})
	}
}

func TestGettingSurveyByID(t *testing.T) {
	//db conn
	db, err := dummy.GetDbConn()
	if err != nil {
		t.Fatal()
	}
	qtx := database.New(db)
	defer qtx.DeleteTestUsers(t.Context())
	//users created
	usersPassword := "psssaass"
	users, err := dummy.CreateUsers(qtx, 2, t, usersPassword)
	if err != nil {
		t.Fatal(err)
	}
	//login user
	accessToken0, _, err := dummy.LoginUser(users[0].Email, usersPassword)
	if err != nil {
		t.Fatal(err)
	}
	accessToken1, _, err := dummy.LoginUser(users[1].Email, usersPassword)
	if err != nil {
		t.Fatal(err)
	}

	// create surveys
	surveyIds, err := dummy.CreateSurvey(qtx, users, t)
	if err != nil {
		t.Fatal(err)
	}
	type testCases struct {
		title      string
		statusCode int
		token      string
		surveyId   string
	}
	runCases := []testCases{
		{"no token", http.StatusUnauthorized, "", surveyIds[0].String()},
		{"bad id", http.StatusNotFound, accessToken0, "f45646546218"},
		{"no id", http.StatusNotFound, accessToken0, ""},
		{"wrong user", http.StatusUnauthorized, accessToken1, surveyIds[0].String()},
		{"wrong user 2", http.StatusUnauthorized, accessToken0, surveyIds[1].String()},
		{"good", http.StatusOK, accessToken0, surveyIds[0].String()},
	}

	for _, tt := range runCases {
		t.Run(tt.title, func(t *testing.T) {
			url := fmt.Sprintf("http://localhost:8080/v0/survey/%s", tt.surveyId)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatal(err)
			}
			if tt.token != "" {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tt.token))
			}
			_, respStatusCode, respBody, err := sendRequest(req)
			if err != nil {
				t.Fatal(err)
			}
			checkingExpectedStatusCode(respStatusCode, respBody, tt.statusCode, t)
		})
	}
}

func TestSetPublish(t *testing.T) {
	//db conn
	db, err := dummy.GetDbConn()
	if err != nil {
		t.Fatal()
	}
	qtx := database.New(db)
	//defer qtx.DeleteTestUsers(t.Context())
	//users created
	usersPassword := "psssaass"
	users, err := dummy.CreateUsers(qtx, 2, t, usersPassword)
	if err != nil {
		t.Fatal(err)
	}
	//login user
	accessToken0, _, err := dummy.LoginUser(users[0].Email, usersPassword)
	if err != nil {
		t.Fatal(err)
	}
	accessToken1, _, err := dummy.LoginUser(users[1].Email, usersPassword)
	if err != nil {
		t.Fatal(err)
	}
	// create surveys
	surveyIds, err := dummy.CreateSurvey(qtx, users, t)
	if err != nil {
		t.Fatal(err)
	}
	type parameters struct {
		SurveyId uuid.UUID `json:"survey_id"`
		Enable   bool      `json:"enable"`
	}
	type testCases struct {
		title         string
		statusCode    int
		token         string
		publish       parameters
		surveyUrl     string
		checkResponse bool
	}
	type response struct {
		SurveyUrl string `json:"survey_url"`
	}
	var url response
	runCases := []testCases{
		{"publish true", http.StatusOK, accessToken0, parameters{SurveyId: surveyIds[0], Enable: true}, fmt.Sprintf("http://localhost:8080/v0/%s/serve", surveyIds[0].String()), true},
		{"publish false", http.StatusOK, accessToken0, parameters{SurveyId: surveyIds[0], Enable: false}, "", false},
		{"no token", http.StatusUnauthorized, "", parameters{SurveyId: surveyIds[0], Enable: true}, fmt.Sprintf("http://localhost:8080/v0/%s/serve", surveyIds[0].String()), true},
		{"wrong token", http.StatusUnauthorized, accessToken1, parameters{SurveyId: surveyIds[0], Enable: true}, fmt.Sprintf("http://localhost:8080/v0/%s/serve", surveyIds[0].String()), true},
		{"publish false 2", http.StatusOK, accessToken1, parameters{SurveyId: surveyIds[1], Enable: false}, "", false},
		{"publish true 2", http.StatusOK, accessToken1, parameters{SurveyId: surveyIds[1], Enable: true}, fmt.Sprintf("http://localhost:8080/v0/%s/serve", surveyIds[1].String()), true},
	}
	for _, tt := range runCases {
		t.Run(tt.title, func(t *testing.T) {
			body, err := json.Marshal(tt.publish)
			if err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v0/publish", bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tt.token))
			_, respStatusCode, respBody, err := sendRequest(req)
			if err != nil {
				t.Fatal(err)
			}

			checkingExpectedStatusCode(respStatusCode, respBody, tt.statusCode, t)
			if tt.checkResponse {
				err = json.Unmarshal(respBody, &url)
				if err != nil {
					t.Fatal(err)
				}

				if url.SurveyUrl != tt.surveyUrl {
					t.Fatalf("response does not match")
				}
			}
		})
	}
}
