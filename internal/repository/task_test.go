//go:build unit

package repository

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/teitiago/task-manager-poc/pkg/models"
)

// ==================
// Store
// ==================

// mockStore Mocks the Store logic
type mockStore struct {
	returnValue models.Task
	returnError error
}

func (m *mockStore) defaultMockBehavior(instance *models.Task) error {
	if m.returnValue != (models.Task{}) {
		*instance = m.returnValue
		return nil
	}
	return m.returnError
}

func (m *mockStore) Migrate(instance interface{}) error {
	return nil
}
func (m *mockStore) Close() {
}

func (m *mockStore) Get(id uuid.UUID, instance interface{}) error {
	return m.defaultMockBehavior(instance.(*models.Task))
}

func (m *mockStore) Filter(filter []Query, instance interface{}, pagination models.Pagination, fields ...string) error {
	return m.returnError
}

func (m *mockStore) Create(instance interface{}) error {
	return m.defaultMockBehavior(instance.(*models.Task))
}

func (m *mockStore) Save(instance interface{}) error {
	return m.defaultMockBehavior(instance.(*models.Task))
}

func (m *mockStore) Delete(instance interface{}) error {
	return m.returnError
}

// ==================
// Tests
// ==================

var defaultTask = models.Task{UserID: uuid.MustParse("ace7fed1-f213-4b20-816e-0101a2db45fe"), Summary: "Test summary"}

// TestGetTask Validates the correct behavior of Get.
func TestGetTask(t *testing.T) {
	testingMap := []struct {
		name          string
		returnValue   models.Task
		returnError   error
		expectedValue models.Task
		expectedError bool
	}{
		{
			name:          "Task - Get - Valid",
			returnValue:   defaultTask,
			returnError:   nil,
			expectedValue: defaultTask,
			expectedError: false,
		},
		{
			name:          "Task - Get - No Result",
			returnValue:   models.Task{},
			returnError:   nil,
			expectedValue: models.Task{},
			expectedError: false,
		},
		{
			name:          "Task - Get - Error",
			returnValue:   models.Task{},
			returnError:   errors.New("failed"),
			expectedValue: models.Task{},
			expectedError: true,
		},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {
			// Given
			taskRepo := NewTaskRepo(&mockStore{returnValue: test.returnValue, returnError: test.returnError})

			// When
			task, err := taskRepo.Get(uuid.MustParse("ace7fed1-f213-4b20-816e-0101a2db45fe"))

			// Then
			if task != test.expectedValue {
				t.Errorf("expected %v, got %v", test.expectedValue, task)
			}
			if !test.expectedError && err != nil {
				t.Error("unexpected error")
			}
			if test.expectedError && err == nil {
				t.Error("expecting an error")
			}

		})
	}
}

// TestFilterTask Validates the correct behavior of Filter.
func TestFilterTask(t *testing.T) {
	testingMap := []struct {
		name          string
		returnArray   []*models.Task
		inputFields   []string
		returnError   error
		expectedValue []*models.Task
		expectedError bool
	}{
		{
			name:          "Task - Filter - Field Invalid",
			returnArray:   nil,
			returnError:   errors.New("failed"),
			inputFields:   []string{"my amazing field"},
			expectedValue: nil,
			expectedError: true,
		},
		{
			name:          "Task - Filter - Error",
			returnArray:   nil,
			returnError:   errors.New("failed"),
			inputFields:   []string{"user_id"},
			expectedValue: nil,
			expectedError: true,
		},
		{
			name:          "Task - Filter - Invalid field",
			returnArray:   nil,
			returnError:   errors.New("failed"),
			inputFields:   []string{"created_at"},
			expectedValue: nil,
			expectedError: true,
		},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {
			// Given
			taskRepo := NewTaskRepo(&mockStore{returnError: test.returnError})

			// When
			tasks, err := taskRepo.Filter(map[string]interface{}{"created_at_gt": "2020-05-20"}, models.Pagination{Limit: 1, Page: 1}, test.inputFields...)

			// Then
			if !reflect.DeepEqual(&tasks, &test.expectedValue) {
				t.Errorf("expected %v, got %v", test.expectedValue, tasks)
			}
			if !test.expectedError && err != nil {
				t.Error("unexpected error")
			}
			if test.expectedError && err == nil {
				t.Error("expecting an error")
			}
		})
	}
}

// TestCreateTask Validates the correct behavior of Create.
func TestCreateTask(t *testing.T) {
	testingMap := []struct {
		name          string
		returnValue   models.Task
		returnError   error
		expectedValue models.Task
		expectedError bool
	}{
		{
			name:          "Task - Create - Valid",
			returnValue:   defaultTask,
			returnError:   nil,
			expectedValue: defaultTask,
			expectedError: false,
		},
		{
			name:          "Task - Create - Error",
			returnValue:   models.Task{},
			returnError:   errors.New("failed"),
			expectedValue: models.Task{},
			expectedError: true,
		},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {
			// Given
			taskRepo := NewTaskRepo(&mockStore{returnValue: test.returnValue, returnError: test.returnError})
			emptyTask := models.Task{}

			// When
			err := taskRepo.Create(&emptyTask)

			// Then
			if emptyTask != test.expectedValue {
				t.Errorf("expected %v, got %v", test.expectedValue, emptyTask)
			}
			if !test.expectedError && err != nil {
				t.Error("unexpected error")
			}
			if test.expectedError && err == nil {
				t.Error("expecting an error")
			}

		})
	}
}

func TestUpdateTask(t *testing.T) {
	testingMap := []struct {
		name          string
		returnValue   models.Task
		returnError   error
		expectedValue models.Task
		expectedError bool
	}{
		{
			name:          "Task - Update - Valid",
			returnValue:   defaultTask,
			returnError:   nil,
			expectedValue: defaultTask,
			expectedError: false,
		},
		{
			name:          "Task - Update - Error",
			returnValue:   models.Task{},
			returnError:   errors.New("failed"),
			expectedValue: models.Task{},
			expectedError: true,
		},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {
			// Given
			taskRepo := NewTaskRepo(&mockStore{returnValue: test.returnValue, returnError: test.returnError})
			emptyTask := models.Task{}

			// When
			err := taskRepo.Update(&emptyTask)

			// Then
			if emptyTask != test.expectedValue {
				t.Errorf("expected %v, got %v", test.expectedValue, emptyTask)
			}
			if !test.expectedError && err != nil {
				t.Error("unexpected error")
			}
			if test.expectedError && err == nil {
				t.Error("expecting an error")
			}

		})
	}
}

func TestDeleteTask(t *testing.T) {
	testingMap := []struct {
		name          string
		returnValue   models.Task
		returnError   error
		expectedValue models.Task
		expectedError bool
	}{
		{
			name:          "Task - Delete - Valid",
			returnValue:   defaultTask,
			returnError:   nil,
			expectedValue: defaultTask,
			expectedError: false,
		},
		{
			name:          "Task - Update - Error",
			returnValue:   models.Task{},
			returnError:   errors.New("failed"),
			expectedValue: models.Task{},
			expectedError: true,
		},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {
			// Given
			taskRepo := NewTaskRepo(&mockStore{returnValue: test.returnValue, returnError: test.returnError})

			// When
			err := taskRepo.Delete(uuid.MustParse("ace7fed1-f213-4b20-816e-0101a2db45fe"))

			// Then
			if !test.expectedError && err != nil {
				t.Error("unexpected error")
			}
			if test.expectedError && err == nil {
				t.Error("expecting an error")
			}

		})
	}
}
