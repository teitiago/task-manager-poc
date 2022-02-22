//go:build unit

package service

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/teitiago/task-manager-poc/internal/server/middleware"
	"github.com/teitiago/task-manager-poc/pkg/models"
)

// ==================
// Repo
// ==================

// mockStore Mocks the Store logic
type mockRepo struct {
	returnValue       models.Task
	returnArray       []*models.Task
	returnError       error
	returnUpdateError error
}

func (m *mockRepo) Get(id uuid.UUID) (models.Task, error) {
	if m.returnValue != (models.Task{}) {
		return m.returnValue, nil
	}
	return models.Task{}, m.returnError
}

func (m *mockRepo) Filter(filter map[string]interface{}, page models.Pagination, fields ...string) ([]*models.Task, error) {
	if m.returnArray != nil {
		return m.returnArray, nil
	}
	return nil, m.returnError
}

func (m *mockRepo) Create(task *models.Task) error {
	return m.returnError
}

func (m *mockRepo) Update(task *models.Task) error {
	return m.returnUpdateError
}

func (m *mockRepo) Delete(id uuid.UUID) error {
	return m.returnError
}

// ==================
// Publisher
// ==================

type mockPublisher struct {
}

func (m *mockPublisher) Publish(payload interface{}, routing string, wg *sync.WaitGroup) {
	defer wg.Done()
}

// ==================
// Encrypter
// ==================

type mockEncrypter struct {
	returnValue string
	returnError error
}

func (m *mockEncrypter) Encrypt(text string) (string, error) {
	if m.returnValue != "" {
		return m.returnValue, nil
	}
	return "", m.returnError
}

func (m *mockEncrypter) Decrypt(text string) (string, error) {
	if m.returnValue != "" {
		return m.returnValue, nil
	}
	return "", m.returnError
}

// ==================
// Tests
// ==================

// TestGet Validates the correct behavior of get function
func TestGet(t *testing.T) {

	techUserContext := context.WithValue(context.Background(), middleware.RoleClaim, "Technician")
	techUserContext = context.WithValue(techUserContext, middleware.SubClaim, "99d8320e-aa07-476a-9d4a-54008bdf6d25")

	managerUserContext := context.WithValue(context.Background(), middleware.RoleClaim, "Technician,Manager")
	managerUserContext = context.WithValue(managerUserContext, middleware.SubClaim, "99d8320e-aa07-476a-9d4a-54008bdf6d25")

	testingMap := []struct {
		name           string
		ctx            context.Context
		returnValue    models.Task
		returnError    error
		decryptReturn  string
		decryptError   error
		expectedTask   models.Task
		expectingError bool
	}{

		{
			name:           "invalid context",
			ctx:            context.Background(),
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "repo error",
			ctx:            techUserContext,
			returnValue:    models.Task{},
			returnError:    errors.New("failed"),
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "empty task no error",
			ctx:            techUserContext,
			returnValue:    models.Task{},
			returnError:    nil,
			expectedTask:   models.Task{},
			expectingError: false,
		},
		{
			name:           "task no access",
			ctx:            techUserContext,
			returnValue:    models.Task{UserID: uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8")},
			returnError:    nil,
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "task access decrypt error",
			ctx:            managerUserContext,
			returnValue:    models.Task{UserID: uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8"), Summary: "This is a summary"},
			returnError:    nil,
			decryptReturn:  "",
			decryptError:   errors.New("failed"),
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "task access decrypt error",
			ctx:            managerUserContext,
			returnValue:    models.Task{UserID: uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8"), Summary: "This is a summary"},
			returnError:    nil,
			decryptReturn:  "decrypted",
			decryptError:   nil,
			expectedTask:   models.Task{UserID: uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8"), Summary: "decrypted"},
			expectingError: false,
		},
		{
			name:           "task access decrypt error",
			ctx:            techUserContext,
			returnValue:    models.Task{UserID: uuid.MustParse("99d8320e-aa07-476a-9d4a-54008bdf6d25"), Summary: "This is a summary"},
			returnError:    nil,
			decryptReturn:  "decrypted",
			decryptError:   nil,
			expectedTask:   models.Task{UserID: uuid.MustParse("99d8320e-aa07-476a-9d4a-54008bdf6d25"), Summary: "decrypted"},
			expectingError: false,
		},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {

			// Given
			service := NewTaskService(
				&mockRepo{returnValue: test.returnValue, returnError: test.returnError},
				nil,
				&mockEncrypter{returnValue: test.decryptReturn, returnError: test.decryptError},
			)

			// When
			task, err := service.Get(test.ctx, uuid.MustParse("59fb4663-ccfb-45f0-96c8-c715f2f9d2b7"))

			// Then
			if !test.expectingError && err != nil {
				t.Errorf("unexpected error %v", err.Error())
			}
			if test.expectingError && err == nil {
				t.Errorf("expecting error")
			}
			if task != test.expectedTask {
				t.Errorf("expected %v, got %v", test.expectedTask, task)
			}

		})
	}

}

// TestFilter Validates the correct behavior of filter.
func TestFilter(t *testing.T) {

	techUserContext := context.WithValue(context.Background(), middleware.RoleClaim, "Technician")
	techUserContext = context.WithValue(techUserContext, middleware.SubClaim, "99d8320e-aa07-476a-9d4a-54008bdf6d25")

	managerUserContext := context.WithValue(context.Background(), middleware.RoleClaim, "Technician,Manager")
	managerUserContext = context.WithValue(managerUserContext, middleware.SubClaim, "99d8320e-aa07-476a-9d4a-54008bdf6d25")

	testingMap := []struct {
		name           string
		ctx            context.Context
		inputFilter    map[string]interface{}
		inputPage      models.Pagination
		returnArray    []*models.Task
		returnError    error
		expectedTask   []*models.Task
		expectingError bool
	}{
		{
			name:           "invalid context",
			ctx:            context.Background(),
			expectedTask:   nil,
			expectingError: true,
		},
		{
			name:           "invalid page",
			ctx:            techUserContext,
			expectedTask:   nil,
			expectingError: true,
		},
		{
			name:           "service error",
			ctx:            techUserContext,
			inputPage:      models.Pagination{Page: 1, Limit: 1},
			returnError:    errors.New("failed"),
			expectedTask:   nil,
			expectingError: true,
		},
		{
			name:           "empty response",
			ctx:            techUserContext,
			inputPage:      models.Pagination{Page: 1, Limit: 1},
			returnError:    nil,
			returnArray:    []*models.Task{},
			expectedTask:   []*models.Task{},
			expectingError: false,
		},
		{
			name:           "nil response",
			ctx:            techUserContext,
			inputPage:      models.Pagination{Page: 1, Limit: 1},
			returnError:    nil,
			returnArray:    nil,
			expectedTask:   nil,
			expectingError: false,
		},
		{
			name:           "valid response",
			ctx:            techUserContext,
			inputPage:      models.Pagination{Page: 1, Limit: 1},
			returnError:    nil,
			returnArray:    []*models.Task{{Summary: "test"}, {Summary: "test 2"}},
			expectedTask:   []*models.Task{{Summary: "test"}, {Summary: "test 2"}},
			expectingError: false,
		},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {

			// Given
			service := NewTaskService(
				&mockRepo{returnArray: test.returnArray, returnError: test.returnError},
				nil,
				nil,
			)

			// When
			tasks, err := service.Filter(test.ctx, test.inputFilter, test.inputPage)

			// Then
			if !test.expectingError && err != nil {
				t.Fatalf("unexpected error %v", err.Error())
			}
			if test.expectingError && err == nil {
				t.Fatal("expecting error")
			}
			if len(tasks) != len(test.expectedTask) {
				t.Errorf("expected %v, got %v", test.expectedTask, tasks)
			}

		})
	}

}

// TestCreateTask Validates the correct behavior of create task
func TestCreateTask(t *testing.T) {

	managerUserContext := context.WithValue(context.Background(), middleware.RoleClaim, "Manager")
	managerUserContext = context.WithValue(managerUserContext, middleware.SubClaim, "99d8320e-aa07-476a-9d4a-54008bdf6d25")

	allUserContext := context.WithValue(context.Background(), middleware.RoleClaim, "Technician,Manager")
	allUserContext = context.WithValue(allUserContext, middleware.SubClaim, "6975dcc2-0ad3-4e05-843a-f39ba22c68f8")

	invalidUserIDContext := context.WithValue(context.Background(), middleware.RoleClaim, "Technician,Manager")
	invalidUserIDContext = context.WithValue(invalidUserIDContext, middleware.SubClaim, "my amazing id")

	testingMap := []struct {
		name           string
		ctx            context.Context
		inputTask      models.Task
		returnError    error
		decryptReturn  string
		decryptError   error
		expectedTask   models.Task
		expectingError bool
	}{
		{
			name:           "invalid context",
			ctx:            context.Background(),
			inputTask:      models.Task{},
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "invalid role",
			ctx:            managerUserContext,
			inputTask:      models.Task{},
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "invalid user id",
			ctx:            invalidUserIDContext,
			inputTask:      models.Task{},
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "decrypt error",
			ctx:            allUserContext,
			inputTask:      models.Task{},
			decryptError:   errors.New("failed"),
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "server error",
			ctx:            allUserContext,
			inputTask:      models.Task{},
			returnError:    errors.New("failed"),
			decryptError:   nil,
			decryptReturn:  "encrypted",
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "valid task",
			ctx:            allUserContext,
			inputTask:      models.Task{UserID: uuid.MustParse("99d8320e-aa07-476a-9d4a-54008bdf6d25"), CompletedDate: sql.NullTime{Valid: true, Time: time.Now()}, Summary: "decrypted"},
			returnError:    nil,
			decryptError:   nil,
			decryptReturn:  "encrypted",
			expectedTask:   models.Task{UserID: uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8"), Summary: "encrypted"},
			expectingError: false,
		},
	}
	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {

			// Given
			service := NewTaskService(
				&mockRepo{returnError: test.returnError},
				nil,
				&mockEncrypter{returnValue: test.decryptReturn, returnError: test.decryptError},
			)

			// When
			task, err := service.Create(test.ctx, test.inputTask)

			// Then
			if !test.expectingError && err != nil {
				t.Errorf("unexpected error %v", err.Error())
			}
			if test.expectingError && err == nil {
				t.Errorf("expecting error")
			}
			if task != test.expectedTask {
				t.Errorf("expected %v, got %v", test.expectedTask, task)
			}

		})
	}
}

// TestUpdate Validates the correct behavior of update task
func TestUpdate(t *testing.T) {

	managerUserContext := context.WithValue(context.Background(), middleware.RoleClaim, "Manager")
	managerUserContext = context.WithValue(managerUserContext, middleware.SubClaim, "99d8320e-aa07-476a-9d4a-54008bdf6d25")

	allUserContext := context.WithValue(context.Background(), middleware.RoleClaim, "Technician,Manager")
	allUserContext = context.WithValue(allUserContext, middleware.SubClaim, "6975dcc2-0ad3-4e05-843a-f39ba22c68f8")

	validCompletedDate := sql.NullTime{Valid: true, Time: time.Now().Add(-12 * time.Hour)}
	invalidCompletedDate := sql.NullTime{Valid: true, Time: time.Now().Add(12 * time.Hour)}

	testingMap := []struct {
		name              string
		ctx               context.Context
		request           models.Task
		returnValue       models.Task
		returnError       error
		returnUpdateError error
		decryptReturn     string
		decryptError      error
		expectedTask      models.Task
		expectingError    bool
	}{

		{
			name:           "invalid context",
			ctx:            context.Background(),
			request:        models.Task{},
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "invalid role",
			ctx:            managerUserContext,
			request:        models.Task{},
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "get task error",
			ctx:            allUserContext,
			request:        models.Task{},
			returnError:    errors.New("failed"),
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "invalid tech user",
			ctx:            allUserContext,
			request:        models.Task{},
			returnError:    nil,
			returnValue:    models.Task{UserID: uuid.MustParse("99d8320e-aa07-476a-9d4a-54008bdf6d25")},
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "empty task get",
			ctx:            allUserContext,
			request:        models.Task{},
			returnError:    nil,
			returnValue:    models.Task{},
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "task summary encrypt error",
			ctx:            allUserContext,
			request:        models.Task{Summary: "decrypted"},
			returnError:    nil,
			returnValue:    models.Task{UserID: uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8")},
			decryptError:   errors.New("failed"),
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "task summary encrypt",
			ctx:            allUserContext,
			request:        models.Task{Summary: "decrypted"},
			returnError:    nil,
			returnValue:    models.Task{UserID: uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8")},
			decryptError:   nil,
			decryptReturn:  "encrypted",
			expectedTask:   models.Task{UserID: uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8"), Summary: "encrypted"},
			expectingError: false,
		},
		{
			name:              "task summary update error",
			ctx:               allUserContext,
			request:           models.Task{Summary: "decrypted"},
			returnError:       nil,
			returnValue:       models.Task{UserID: uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8")},
			returnUpdateError: errors.New("failed"),
			decryptError:      nil,
			decryptReturn:     "encrypted",
			expectedTask:      models.Task{},
			expectingError:    true,
		},

		{
			name:           "task already completed",
			ctx:            allUserContext,
			request:        models.Task{Summary: "decrypted"},
			returnError:    nil,
			returnValue:    models.Task{UserID: uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8"), CompletedDate: invalidCompletedDate},
			decryptError:   nil,
			decryptReturn:  "encrypted",
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "task completed in the future",
			ctx:            allUserContext,
			request:        models.Task{Summary: "decrypted", CompletedDate: invalidCompletedDate},
			returnError:    nil,
			returnValue:    models.Task{UserID: uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8")},
			decryptError:   nil,
			decryptReturn:  "encrypted",
			expectedTask:   models.Task{},
			expectingError: true,
		},
		{
			name:           "task invalid completed date",
			ctx:            allUserContext,
			request:        models.Task{Summary: "decrypted", CompletedDate: validCompletedDate},
			returnError:    nil,
			returnValue:    models.Task{UserID: uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8")},
			decryptError:   nil,
			decryptReturn:  "encrypted",
			expectedTask:   models.Task{UserID: uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8"), Summary: "encrypted", CompletedDate: validCompletedDate},
			expectingError: false,
		},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {

			// Given
			service := NewTaskService(
				&mockRepo{returnError: test.returnError, returnValue: test.returnValue, returnUpdateError: test.returnUpdateError},
				&mockPublisher{},
				&mockEncrypter{returnValue: test.decryptReturn, returnError: test.decryptError},
			)

			// When
			task, err := service.Patch(test.ctx, uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8"), test.request)

			// Then
			if !test.expectingError && err != nil {
				t.Errorf("unexpected error %v", err.Error())
			}
			if test.expectingError && err == nil {
				t.Errorf("expecting error")
			}
			if task != test.expectedTask {
				t.Errorf("expected %v, got %v", test.expectedTask, task)
			}

		})
	}
}

// TestDelete Validates the correct behavior of delete task
func TestDelete(t *testing.T) {

	managerUserContext := context.WithValue(context.Background(), middleware.RoleClaim, "Manager")
	managerUserContext = context.WithValue(managerUserContext, middleware.SubClaim, "99d8320e-aa07-476a-9d4a-54008bdf6d25")

	techUserContext := context.WithValue(context.Background(), middleware.RoleClaim, "Technician")
	techUserContext = context.WithValue(techUserContext, middleware.SubClaim, "6975dcc2-0ad3-4e05-843a-f39ba22c68f8")

	testingMap := []struct {
		name           string
		ctx            context.Context
		returnError    error
		expectingError bool
	}{
		{
			name:           "invalid context",
			ctx:            context.Background(),
			expectingError: true,
		},
		{
			name:           "invalid role",
			ctx:            techUserContext,
			expectingError: true,
		},
		{
			name:           "repo error",
			ctx:            managerUserContext,
			returnError:    errors.New("failed"),
			expectingError: true,
		},
		{
			name:           "repo error",
			ctx:            managerUserContext,
			returnError:    nil,
			expectingError: false,
		},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {

			// Given
			service := NewTaskService(
				&mockRepo{returnError: test.returnError},
				nil,
				nil,
			)

			// When
			err := service.Delete(test.ctx, uuid.MustParse("6975dcc2-0ad3-4e05-843a-f39ba22c68f8"))

			// Then
			if !test.expectingError && err != nil {
				t.Errorf("unexpected error %v", err.Error())
			}
			if test.expectingError && err == nil {
				t.Errorf("expecting error")
			}
		})
	}

}
