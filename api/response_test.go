package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gray509/survy/server/internal/database"
	"github.com/gray509/survy/server/internal/querieutils"
)

func TestRequestAnswerValidation(t *testing.T) {
	ids := make([]uuid.UUID, 0)
	for i := 0; i < 8; i++ {
		ids = append(ids, uuid.New())
	}
	now := time.Now()
	timestamptz := querieutils.Time(&now)
	type testCases struct {
		title       string
		rAnswers    string
		err         error
		dbQuestions []database.GetQuestionsFromSurveyIdRow
		dbAnswers   map[string]interface{}
	}
	runCases := []testCases{
		{
			"0: valid, all questions are required",
			fmt.Sprintf(`{
			"survey_id": "%s",
			"%s": [0,2],
			"%s": 1,
			"%s": 5,
			"%s": [2,0,1],
			"%s": "this the answers"
		}`, ids[7].String(), ids[0].String(), ids[1].String(), ids[2].String(), ids[3].String(), ids[4].String()),
			nil,
			[]database.GetQuestionsFromSurveyIdRow{
				{ID: ids[0], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "0:checkbox", QuestionType: Checkbox, Choice: []string{"red", "blue", "green"}, IsRequired: true},
				{ID: ids[1], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "0:radio", QuestionType: Radio, Choice: []string{"red", "blue", "green"}, IsRequired: true},
				{ID: ids[2], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "0:rating", QuestionType: Rating, Choice: []string{"0", "5"}, IsRequired: true},
				{ID: ids[3], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "0:ranking", QuestionType: Ranking, Choice: []string{"red", "blue", "green"}, IsRequired: true},
				{ID: ids[4], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "0:opentext", QuestionType: OpenText, Choice: nil, IsRequired: true},
			},
			map[string]interface{}{
				ids[0].String(): []int{0, 2},
				ids[1].String(): int(1),
				ids[2].String(): int(5),
				ids[3].String(): []int{2, 0, 1},
				ids[4].String(): string("this the answers"),
			},
		},
		{
			"1: valid, some question are not required",
			fmt.Sprintf(`{
			"survey_id": "%s",
			"%s": [0,2],
			"%s": 1,
			"%s": 5,
			"%s": [2,0,1],
			"%s": "this the answers",
			"%s": [1,2],
			"%s": 2
		}`, ids[7].String(), ids[0].String(), ids[1].String(), ids[2].String(), ids[3].String(), ids[4].String(), ids[5].String(), ids[6].String()),
			nil,
			[]database.GetQuestionsFromSurveyIdRow{
				{ID: ids[0], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "1:checkbox", QuestionType: Checkbox, Choice: []string{"red", "blue", "green"}, IsRequired: true},
				{ID: ids[1], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "1:radio", QuestionType: Radio, Choice: []string{"red", "blue", "green"}, IsRequired: true},
				{ID: ids[2], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "1:rating", QuestionType: Rating, Choice: []string{"0", "5"}, IsRequired: true},
				{ID: ids[3], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "1:ranking", QuestionType: Ranking, Choice: []string{"red", "blue", "green"}, IsRequired: true},
				{ID: ids[4], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "1:opentext", QuestionType: OpenText, Choice: nil, IsRequired: true},
				{ID: ids[5], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "1:checkbox2", QuestionType: Checkbox, Choice: []string{"redf", "bluef", "greenf"}, IsRequired: false},
				{ID: ids[6], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "1:radio2", QuestionType: Radio, Choice: []string{"red", "blue", "green"}, IsRequired: false},
			},
			map[string]interface{}{
				ids[0].String(): []int{0, 2},
				ids[1].String(): int(1),
				ids[2].String(): int(5),
				ids[3].String(): []int{2, 0, 1},
				ids[4].String(): string("this the answers"),
				ids[5].String(): []int{1, 2},
				ids[6].String(): int(2),
			},
		},
		{
			"2: question is required but answer not exist",
			fmt.Sprintf(`{
			"survey_id": "%s",
			"%s": 1,
			"%s": 5,
			"%s": [2,0,1],
			"%s": "this the answers"
		}`, ids[7].String(), ids[1].String(), ids[2].String(), ids[3].String(), ids[4].String()),
			ErrMissingRequiredQuest,
			[]database.GetQuestionsFromSurveyIdRow{
				{ID: ids[0], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "2:checkbox", QuestionType: Checkbox, Choice: []string{"red", "blue", "green"}, IsRequired: true},
				{ID: ids[1], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "2:radio", QuestionType: Radio, Choice: []string{"red", "blue", "green"}, IsRequired: true},
				{ID: ids[2], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "2:rating", QuestionType: Rating, Choice: []string{"0", "5"}, IsRequired: true},
				{ID: ids[3], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "2:ranking", QuestionType: Ranking, Choice: []string{"red", "blue", "green"}, IsRequired: true},
				{ID: ids[4], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "2:opentext", QuestionType: OpenText, Choice: nil, IsRequired: true},
			},
			nil,
		},
		{
			"3: invalid checkbox answer (value does not exist)",
			fmt.Sprintf(`{
			"survey_id": "%s",
			"%s": [0,5]
		}`, ids[7].String(), ids[0].String()),
			ErrAccessingNonExisting,
			[]database.GetQuestionsFromSurveyIdRow{
				{ID: ids[0], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "3:checkbox", QuestionType: Checkbox, Choice: []string{"red", "blue", "green"}, IsRequired: true},
			},
			nil,
		},
		{
			"4: invalid radio answer (value does not exist)",
			fmt.Sprintf(`{
			"survey_id": "%s",
			"%s": 5
		}`, ids[7].String(), ids[0].String()),
			ErrAccessingNonExisting,
			[]database.GetQuestionsFromSurveyIdRow{
				{ID: ids[0], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "4:radio", QuestionType: Radio, Choice: []string{"red", "blue", "green"}, IsRequired: true},
			},
			nil,
		},
		{
			"5: invalid ranking answer (number of choice not consistant with db)",
			fmt.Sprintf(`{
			"survey_id": "%s",
			"%s": [2,1]
		}`, ids[7].String(), ids[0].String()),
			fmt.Errorf("request and db choices are not consistant"),
			[]database.GetQuestionsFromSurveyIdRow{
				{ID: ids[0], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "5:ranking", QuestionType: Ranking, Choice: []string{"red", "blue", "green"}, IsRequired: true},
			},
			nil,
		},
		{
			"6: invalid ranking answer (value does not exist)",
			fmt.Sprintf(`{
			"survey_id": "%s",
			"%s": [2,1,45]
		}`, ids[7].String(), ids[0].String()),
			ErrAccessingNonExisting,
			[]database.GetQuestionsFromSurveyIdRow{
				{ID: ids[0], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "6:ranking", QuestionType: Ranking, Choice: []string{"red", "blue", "green"}, IsRequired: true},
			},
			nil,
		},
		{
			"7: invalid answer type (open text)",
			fmt.Sprintf(`{
			"survey_id": "%s",
			"%s": [2,1,45]
		}`, ids[7].String(), ids[0].String()),
			ErrMissingRequiredQuest,
			[]database.GetQuestionsFromSurveyIdRow{
				{ID: ids[0], CreatedAt: timestamptz, UpdatedAt: timestamptz, Title: "7:opentext", QuestionType: OpenText, Choice: nil, IsRequired: true},
			},
			nil,
		},
	}

	for _, tt := range runCases {
		t.Run(tt.title, func(t *testing.T) {
			var requestAnswers map[string]interface{}
			err := json.Unmarshal([]byte(tt.rAnswers), &requestAnswers)
			if err != nil {
				t.Fatal(err)
			}
			testdbAnswers, testErr := requestAnwserValidation(tt.dbQuestions, requestAnswers)
			if !errors.Is(testErr, tt.err) {
				t.Fatalf("\nerrors dont match\ntestERR: %s\nexpected error: %s", testErr, tt.err)
			}

			if !checkIfMapEqual(testdbAnswers, tt.dbAnswers) {
				t.Fatalf("\nanswers dont match\ntest answer:     %v\nexpected answer: %v", testdbAnswers, tt.dbAnswers)
			}

		})
	}

}

func checkIfMapEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		fmt.Println(k, v, b[k])
		fmt.Println(reflect.DeepEqual(v, b[k]))
		if !reflect.DeepEqual(v, b[k]) {
			return false
		}
	}
	return true
}
