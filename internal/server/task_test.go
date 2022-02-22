//go:build unit

package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/teitiago/task-manager-poc/pkg/models"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// ==================
// Service
// ==================

// mockTaskService Mocks the Task Service logic
type mockTaskService struct {
	returnValue models.Task
	returnArray []*models.Task
	returnError error
}

func (m *mockTaskService) Create(ctx context.Context, task models.Task) (models.Task, error) {
	if m.returnValue != (models.Task{}) {
		return m.returnValue, nil
	}
	return models.Task{}, m.returnError
}

func (m *mockTaskService) Delete(ctx context.Context, id uuid.UUID) error {
	return m.returnError
}

func (m *mockTaskService) Patch(ctx context.Context, id uuid.UUID, task models.Task) (models.Task, error) {
	if m.returnValue != (models.Task{}) {
		return m.returnValue, nil
	}
	return models.Task{}, m.returnError
}

func (m *mockTaskService) Get(ctx context.Context, id uuid.UUID) (models.Task, error) {
	if m.returnValue != (models.Task{}) {
		return m.returnValue, nil
	}
	return models.Task{}, m.returnError
}

func (m *mockTaskService) Filter(ctx context.Context, filter map[string]interface{}, page models.Pagination) ([]*models.Task, error) {
	if m.returnArray != nil {
		return m.returnArray, nil
	}
	return nil, m.returnError
}

// ==================
// Tests
// ==================

// TestCreateTask validates the correct behavior of CreateTask
func TestCreateTask(t *testing.T) {

	testingMap := []struct {
		name         string
		returnValue  models.Task
		returnError  error
		inputBody    string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "invalid body",
			returnValue:  models.Task{},
			returnError:  nil,
			inputBody:    `{"firstname":"test","email":"test@test.com"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"status":"error while binding request"}`,
		},
		{
			name:         "invalid summary len",
			returnValue:  models.Task{},
			returnError:  nil,
			inputBody:    fmt.Sprintf(`{"summary":"%v"}`, randStringRunes(3000)),
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"status":"error while binding request"}`,
		},
		{
			name:         "service error",
			returnValue:  models.Task{},
			returnError:  errors.New("failed"),
			inputBody:    fmt.Sprintf(`{"summary":"%v"}`, randStringRunes(10)),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"status":"task creation failed"}`,
		},
		{
			name:         "valid request",
			returnValue:  models.Task{ID: uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466")},
			returnError:  errors.New("failed"),
			inputBody:    fmt.Sprintf(`{"summary":"%v"}`, randStringRunes(10)),
			expectedCode: http.StatusCreated,
			expectedBody: `{"id":"74531653-252b-48c7-b562-63e82f5e3466", "status":"task created"}`,
		},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {
			// Given
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := &http.Request{
				URL:    &url.URL{},
				Header: make(http.Header),
				Body:   io.NopCloser(strings.NewReader(string(test.inputBody))),
			}

			c.Request = req

			// When
			handler := NewTaskHandler(&mockTaskService{returnValue: test.returnValue, returnError: test.returnError})
			handler.CreateTask(c)

			// Then
			if w.Code != test.expectedCode {
				t.Errorf("expected %v got: %v", test.expectedCode, w.Code)
			}

			var expectedBody interface{}
			var gotBody interface{}
			json.Unmarshal([]byte(test.expectedBody), &expectedBody)
			json.Unmarshal([]byte(w.Body.String()), &gotBody)

			if !reflect.DeepEqual(expectedBody, gotBody) {
				t.Errorf("expected %v got: %v", expectedBody, gotBody)
			}
		})

	}
}

// TestDeleteTask validates the correct behavior of DeleteTask
func TestDeleteTask(t *testing.T) {

	testingMap := []struct {
		name         string
		inputUUID    string
		returnError  error
		expectedCode int
	}{
		{name: "invalid uuid", inputUUID: "invalid-uuid", returnError: nil, expectedCode: http.StatusNotFound},
		{name: "ise", inputUUID: "6dafac59-c6a3-473c-8056-19e03620fc55", returnError: errors.New("failed"), expectedCode: http.StatusInternalServerError},
		{name: "valid response", inputUUID: "6dafac59-c6a3-473c-8056-19e03620fc55", returnError: nil, expectedCode: http.StatusNoContent},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {
			// Given
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{
				{
					Key:   "taskID",
					Value: test.inputUUID,
				},
			}

			// When
			handler := NewTaskHandler(&mockTaskService{returnValue: models.Task{}, returnError: test.returnError})
			handler.DeleteTask(c)

			// Then
			if test.expectedCode != w.Code {
				t.Errorf("expected %v, got %v", test.expectedCode, w.Code)
			}
		})
	}
}

// TestPatchTask Validates the correct behavior of Patch task
func TestPatchTask(t *testing.T) {

	timeval := time.Now().Unix() + 500
	fmt.Printf("%v", timeval)

	testingMap := []struct {
		name         string
		inputUUID    string
		returnValue  models.Task
		returnError  error
		inputBody    string
		expectedCode int
	}{
		{
			name:         "invalid uuid",
			inputUUID:    "invalid uuid",
			returnValue:  models.Task{},
			returnError:  nil,
			inputBody:    `{"firstname":"test","email":"test@test.com"}`,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid body",
			inputUUID:    "6dafac59-c6a3-473c-8056-19e03620fc55",
			returnValue:  models.Task{},
			returnError:  nil,
			inputBody:    `{"firstname":"test","email":"test@test.com"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid summary len",
			inputUUID:    "6dafac59-c6a3-473c-8056-19e03620fc55",
			returnValue:  models.Task{},
			returnError:  nil,
			inputBody:    fmt.Sprintf(`{"summary":"%v"}`, randStringRunes(3000)),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "service error",
			inputUUID:    "6dafac59-c6a3-473c-8056-19e03620fc55",
			returnValue:  models.Task{},
			returnError:  errors.New("failed"),
			inputBody:    fmt.Sprintf(`{"summary":"%v"}`, randStringRunes(10)),
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "valid summary request",
			inputUUID:    "6dafac59-c6a3-473c-8056-19e03620fc55",
			returnValue:  models.Task{ID: uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466")},
			returnError:  nil,
			inputBody:    fmt.Sprintf(`{"summary":"%v"}`, randStringRunes(10)),
			expectedCode: http.StatusNoContent,
		},
		{
			name:         "invalid completed date request string value",
			inputUUID:    "6dafac59-c6a3-473c-8056-19e03620fc55",
			returnValue:  models.Task{},
			returnError:  nil,
			inputBody:    fmt.Sprintf(`{"completed_date":%v}`, "string test"),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid completed date request negative value",
			inputUUID:    "6dafac59-c6a3-473c-8056-19e03620fc55",
			returnValue:  models.Task{},
			returnError:  nil,
			inputBody:    fmt.Sprintf(`{"completed_date":%v}`, -10),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid completed date request 0",
			inputUUID:    "6dafac59-c6a3-473c-8056-19e03620fc55",
			returnValue:  models.Task{},
			returnError:  nil,
			inputBody:    fmt.Sprintf(`{"completed_date":%v}`, 0),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid completed date request future",
			inputUUID:    "6dafac59-c6a3-473c-8056-19e03620fc55",
			returnValue:  models.Task{},
			returnError:  nil,
			inputBody:    fmt.Sprintf(`{"completed_date":%v}`, time.Now().Unix()+500),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "valid completed date request",
			inputUUID:    "6dafac59-c6a3-473c-8056-19e03620fc55",
			returnValue:  models.Task{ID: uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466")},
			returnError:  nil,
			inputBody:    fmt.Sprintf(`{"completed_date":%v}`, time.Now().Unix()-500),
			expectedCode: http.StatusNoContent,
		},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {
			// Given
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = []gin.Param{
				{
					Key:   "taskID",
					Value: test.inputUUID,
				},
			}

			req := &http.Request{
				URL:    &url.URL{},
				Header: make(http.Header),
				Body:   io.NopCloser(strings.NewReader(string(test.inputBody))),
			}

			c.Request = req

			// When
			handler := NewTaskHandler(&mockTaskService{returnValue: test.returnValue, returnError: test.returnError})
			handler.PatchTask(c)

			// Then
			if w.Code != test.expectedCode {
				t.Errorf("expected %v got: %v", test.expectedCode, w.Code)
			}
		})

	}
}

// TestGetTask Validates the correct behavior of GetTask
func TestGetTask(t *testing.T) {

	timeLayout := "2006-01-02T15:04:05.000Z"
	timeValue, err := time.Parse(timeLayout, "2020-05-20T15:00:00.000Z")
	if err != nil {
		t.Error("unexpected time error")
	}

	testingMap := []struct {
		name         string
		inputUUID    string
		returnError  error
		returnTask   models.Task
		expectedCode int
		expectedBody string
	}{
		{
			name:         "invalid uuid",
			inputUUID:    "invalid-uuid",
			returnError:  nil,
			returnTask:   models.Task{},
			expectedCode: http.StatusNotFound,
			expectedBody: ``,
		},
		{
			name:         "valid uuid - service error",
			inputUUID:    "6dafac59-c6a3-473c-8056-19e03620fc55",
			returnError:  errors.New("failed"),
			returnTask:   models.Task{},
			expectedCode: http.StatusInternalServerError,
			expectedBody: ``,
		},
		{
			name:         "valid uuid - no user_id and completed_date",
			inputUUID:    "6dafac59-c6a3-473c-8056-19e03620fc55",
			returnError:  errors.New("failed"),
			returnTask:   models.Task{CreatedAt: timeValue, UpdatedAt: timeValue, Summary: "my summary"},
			expectedCode: http.StatusOK,
			expectedBody: `{"created_at": 1589986800, "modified_at": 1589986800, "summary": "my summary", "task_id":"00000000-0000-0000-0000-000000000000"}`,
		},
		{
			name:         "valid uuid - no completed_date",
			inputUUID:    "6dafac59-c6a3-473c-8056-19e03620fc55",
			returnError:  errors.New("failed"),
			returnTask:   models.Task{CreatedAt: timeValue, UpdatedAt: timeValue, Summary: "my summary", UserID: uuid.MustParse("6dafac59-c6a3-473c-8056-19e03620fc55")},
			expectedCode: http.StatusOK,
			expectedBody: `{"user_id": "6dafac59-c6a3-473c-8056-19e03620fc55", "created_at": 1589986800, "modified_at": 1589986800, "summary": "my summary", "task_id":"00000000-0000-0000-0000-000000000000"}`,
		},
		{
			name:         "complete uuid",
			inputUUID:    "6dafac59-c6a3-473c-8056-19e03620fc55",
			returnError:  errors.New("failed"),
			returnTask:   models.Task{CreatedAt: timeValue, UpdatedAt: timeValue, Summary: "my summary", UserID: uuid.MustParse("6dafac59-c6a3-473c-8056-19e03620fc55"), CompletedDate: sql.NullTime{Time: timeValue, Valid: true}},
			expectedCode: http.StatusOK,
			expectedBody: `{"completed_date": 1589986800, "user_id": "6dafac59-c6a3-473c-8056-19e03620fc55", "created_at": 1589986800, "modified_at": 1589986800, "summary": "my summary", "task_id":"00000000-0000-0000-0000-000000000000"}`,
		},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {
			// Given
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{
				{
					Key:   "taskID",
					Value: test.inputUUID,
				},
			}

			// When
			handler := NewTaskHandler(&mockTaskService{returnValue: test.returnTask, returnError: test.returnError})
			handler.GetTask(c)

			// Then
			if test.expectedCode != w.Code {
				t.Errorf("expected %v, got %v", test.expectedCode, w.Code)
			}
			var expectedBody interface{}
			var gotBody interface{}
			json.Unmarshal([]byte(test.expectedBody), &expectedBody)
			json.Unmarshal([]byte(w.Body.String()), &gotBody)

			if !reflect.DeepEqual(expectedBody, gotBody) {
				t.Errorf("expected %v got: %v", expectedBody, gotBody)
			}
		})
	}
}

func TestListTasks(t *testing.T) {

	timeLayout := "2006-01-02T15:04:05.000Z"
	timeValue, err := time.Parse(timeLayout, "2020-05-20T15:00:00.000Z")
	if err != nil {
		t.Error("unexpected time error")
	}

	testingMap := []struct {
		name         string
		inputFilter  url.Values
		returnError  error
		returnTask   []*models.Task
		expectedCode int
		expectedBody string
	}{
		{
			name:         "invalid limit",
			inputFilter:  url.Values{"limit": []string{"test"}},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"status":"invalid limit value"}`,
		},
		{
			name:         "invalid page",
			inputFilter:  url.Values{"page": []string{"test"}},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"status":"invalid page value"}`,
		},
		{
			name:         "invalid sort",
			inputFilter:  url.Values{"sort": []string{"test"}},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"status":"invalid sort suffix must be asc or desc"}`,
		},
		{
			name:         "server error",
			returnError:  errors.New("failed"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: ``,
		},
		{
			name:         "empty tasks",
			returnError:  nil,
			returnTask:   nil,
			expectedCode: http.StatusNotFound,
			expectedBody: ``,
		},
		{
			name:        "single task sort asc",
			inputFilter: url.Values{"sort": []string{"created_at asc"}, "my_value": []string{"my_value"}},
			returnError: nil,
			returnTask: []*models.Task{
				{
					ID:            uuid.MustParse("db5ea21a-a7f7-4123-92b4-292c86da8a49"),
					UserID:        uuid.MustParse("db5ea21a-a7f7-4123-92b4-292c86da8a49"),
					CreatedAt:     timeValue,
					UpdatedAt:     timeValue,
					CompletedDate: sql.NullTime{Valid: true, Time: timeValue},
				},
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"status": "OK", "tasks": [{"user_id": "db5ea21a-a7f7-4123-92b4-292c86da8a49", "task_id": "db5ea21a-a7f7-4123-92b4-292c86da8a49", "created_at": 1589986800, "modified_at":1589986800, "completed_date":1589986800}]}`,
		},
		{
			name:        "single task sort desc",
			inputFilter: url.Values{"sort": []string{"created_at desc"}},
			returnError: nil,
			returnTask: []*models.Task{
				{
					ID:            uuid.MustParse("db5ea21a-a7f7-4123-92b4-292c86da8a49"),
					UserID:        uuid.MustParse("db5ea21a-a7f7-4123-92b4-292c86da8a49"),
					CreatedAt:     timeValue,
					UpdatedAt:     timeValue,
					CompletedDate: sql.NullTime{Valid: true, Time: timeValue},
				},
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"status": "OK", "tasks": [{"user_id": "db5ea21a-a7f7-4123-92b4-292c86da8a49", "task_id": "db5ea21a-a7f7-4123-92b4-292c86da8a49", "created_at": 1589986800, "modified_at":1589986800, "completed_date":1589986800}]}`,
		},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {
			// Given
			w := httptest.NewRecorder()
			req := &http.Request{
				URL:    &url.URL{RawQuery: test.inputFilter.Encode()},
				Header: make(http.Header),
			}

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// When
			handler := NewTaskHandler(&mockTaskService{returnArray: test.returnTask, returnError: test.returnError})
			handler.ListTasks(c)

			// Then
			if test.expectedCode != w.Code {
				t.Errorf("expected %v, got %v", test.expectedCode, w.Code)
			}
			var expectedBody interface{}
			var gotBody interface{}
			json.Unmarshal([]byte(test.expectedBody), &expectedBody)
			json.Unmarshal([]byte(w.Body.String()), &gotBody)

			if !reflect.DeepEqual(expectedBody, gotBody) {
				t.Errorf("expected %v got: %v", expectedBody, gotBody)
			}
		})
	}

}
