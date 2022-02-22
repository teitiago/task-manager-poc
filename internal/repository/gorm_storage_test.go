//go:build integration

package repository

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/teitiago/task-manager-poc/pkg/models"
)

func clearTasks(storage *gormStorage) {
	storage.db.Exec("TRUNCATE TABLE tasks")
}

func TestMain(m *testing.M) {
	storage := NewGormStorage()
	clearTasks(&storage)
	m.Run()
	clearTasks(&storage)
	storage.Close()
}

// TestValidGet Validates that when querying a specific id the corresponding
// task object is retrieved.
func TestValidGet(t *testing.T) {

	// Given
	storage := NewGormStorage()

	taskRepo := NewTaskRepo(&storage)

	expectedTask := &models.Task{
		UserID:  uuid.MustParse("ace7fed1-f213-4b20-816e-0101a2db45fe"),
		Summary: "This is a simple test",
	}

	err := taskRepo.Create(expectedTask)
	if err != nil {
		t.Fatalf("unexpected error %v", err.Error())
	}

	// When
	gotTask, err := taskRepo.Get(expectedTask.ID)

	// Then
	if gotTask.ID != expectedTask.ID ||
		gotTask.UserID != expectedTask.UserID ||
		gotTask.Summary != expectedTask.Summary ||
		gotTask.CompletedDate != expectedTask.CompletedDate {
		t.Errorf("expected %v got %v", expectedTask, gotTask)
	}
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

}

// TestInvalidGet Validates that when trying to get a non existing record
// no error is raised and an empty struct is returned.
func TestInvalidGet(t *testing.T) {
	// Given
	storage := NewGormStorage()

	taskRepo := NewTaskRepo(&storage)

	// When
	gotTask, err := taskRepo.Get(uuid.MustParse("cb466277-960d-44ec-a588-94cbc1c85c3e"))

	// Then
	if gotTask != (models.Task{}) {
		t.Errorf("expecting empty struct")
	}
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

// TestFilter Validates that when trying to use the filter the correct tasks are collected.
func TestFilter(t *testing.T) {

	// Given
	storage := NewGormStorage()
	taskRepo := NewTaskRepo(&storage)

	expectedTasks := []models.Task{
		{
			UserID:        uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466"),
			CompletedDate: sql.NullTime{Time: time.Now().AddDate(0, 0, -30), Valid: true},
		},
		{
			UserID:        uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466"),
			CompletedDate: sql.NullTime{Time: time.Now().AddDate(0, 0, -30), Valid: true},
		},
		{
			UserID:  uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466"),
			Summary: "same user task, different description",
		},
	}
	for _, task := range expectedTasks {
		taskRepo.Create(&task)
	}

	nonExpectedTask := models.Task{
		UserID:  uuid.MustParse("0df92b9c-103d-4c97-9b11-986c3f3e23a2"),
		Summary: "different user task",
	}
	taskRepo.Create(&nonExpectedTask)

	testingMap := []struct {
		name            string
		inputFilter     map[string]interface{}
		inputPagination models.Pagination
		inputFields     []string
		expectedLen     int
		expectedError   bool
	}{
		{
			name:            "filter by one field",
			inputFilter:     map[string]interface{}{"user_id": uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466")},
			inputPagination: models.Pagination{Limit: 10, Page: 1, Sort: "created_at asc"},
			inputFields:     []string{"user_id"},
			expectedLen:     3,
			expectedError:   false,
		},
		{
			name:            "filter by two fields and lte",
			inputFilter:     map[string]interface{}{"user_id": uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466"), "completed_date_lte": fmt.Sprint(time.Now().UTC().Unix())},
			inputPagination: models.Pagination{Limit: 10, Page: 1, Sort: "created_at asc"},
			inputFields:     []string{"user_id"},
			expectedLen:     2,
			expectedError:   false,
		},
		{
			name:            "filter by field lt",
			inputFilter:     map[string]interface{}{"user_id": uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466"), "completed_date_lt": fmt.Sprint(time.Now().UTC().Unix())},
			inputPagination: models.Pagination{Limit: 10, Page: 1, Sort: "created_at asc"},
			inputFields:     []string{"user_id"},
			expectedLen:     2,
			expectedError:   false,
		},
		{
			name:            "filter by field gte",
			inputFilter:     map[string]interface{}{"user_id": uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466"), "completed_date_gte": fmt.Sprint(time.Now().AddDate(0, 0, -40).UTC().Unix())},
			inputPagination: models.Pagination{Limit: 10, Page: 1, Sort: "created_at asc"},
			inputFields:     []string{"user_id"},
			expectedLen:     2,
			expectedError:   false,
		},
		{
			name:            "filter by field gt",
			inputFilter:     map[string]interface{}{"user_id": uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466"), "completed_date_gt": fmt.Sprint(time.Now().AddDate(0, 0, -40).UTC().Unix())},
			inputPagination: models.Pagination{Limit: 10, Page: 1, Sort: "created_at asc"},
			inputFields:     []string{"user_id"},
			expectedLen:     2,
			expectedError:   false,
		},
		{
			name:            "filter with no results",
			inputFilter:     map[string]interface{}{"user_id": "invalid owner"},
			inputPagination: models.Pagination{Limit: 10, Page: 1, Sort: "created_at asc"},
			inputFields:     []string{"user_id"},
			expectedLen:     0,
			expectedError:   false,
		},
		{
			name:            "filter pagination single result",
			inputFilter:     map[string]interface{}{"user_id": uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466")},
			inputPagination: models.Pagination{Limit: 1, Page: 1, Sort: "created_at asc"},
			inputFields:     []string{"user_id"},
			expectedLen:     1,
			expectedError:   false,
		},
		{
			name:            "filter pagination single result",
			inputFilter:     map[string]interface{}{"user_id": uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466")},
			inputPagination: models.Pagination{Limit: 2, Page: 1, Sort: "created_at asc"},
			inputFields:     []string{"user_id"},
			expectedLen:     2,
			expectedError:   false,
		},
		{
			name:            "filter pagination single result",
			inputFilter:     map[string]interface{}{"user_id": uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466")},
			inputPagination: models.Pagination{Limit: 10, Page: 2, Sort: "created_at asc"},
			inputFields:     []string{"user_id"},
			expectedLen:     0,
			expectedError:   false,
		},
		{
			name:            "filter pagination invalid sort",
			inputFilter:     map[string]interface{}{"user_id": uuid.MustParse("74531653-252b-48c7-b562-63e82f5e3466")},
			inputPagination: models.Pagination{Limit: 10, Page: 2, Sort: "invalid sort asc"},
			inputFields:     []string{"user_id"},
			expectedLen:     0,
			expectedError:   true,
		},
		{
			name:            "filter no filter",
			inputFilter:     map[string]interface{}{},
			inputPagination: models.Pagination{Limit: 10, Page: 2},
			inputFields:     []string{"user_id"},
			expectedLen:     0,
			expectedError:   false,
		},
	}

	for _, test := range testingMap {
		// When
		gotTasks, err := taskRepo.Filter(test.inputFilter, test.inputPagination, test.inputFields...)

		// Then
		if gotTasks != nil && len(gotTasks) != test.expectedLen {
			t.Errorf("expected %v, got %v", test.expectedLen, gotTasks)
		}
		if !test.expectedError && err != nil {
			t.Error("unexpected error", err.Error())
		}
		if test.expectedError && err == nil {
			t.Error("expecting an error")
		}
	}

}

// TestUpdate Validates that when a task is updated the actual value is changed on the database.
func TestUpdate(t *testing.T) {
	// Given
	storage := NewGormStorage()
	taskRepo := NewTaskRepo(&storage)

	inputTask := models.Task{
		UserID:  uuid.MustParse("5f1173fd-bf02-4252-8c9f-1bb4325980ad"),
		Summary: "Non updated task",
	}

	err := taskRepo.Create(&inputTask)
	if err != nil {
		t.Fatalf("unexpected test error %v", err.Error())
	}

	// When
	inputTask.Summary = "updated task"
	taskRepo.Update(&inputTask)

	// Then
	gotTask, err := taskRepo.Get(inputTask.ID)
	if gotTask.Summary != inputTask.Summary {
		t.Errorf("expected same summary, got different values, %v, %v", gotTask.Summary, inputTask.Summary)
	}
	if err != nil {
		t.Errorf("unexpected test error, %v", err.Error())
	}
}

// TestUpdateNonExisting Validates that is not possible to update a non existing value.
func TestUpdateNonExisting(t *testing.T) {

	// Given
	storage := NewGormStorage()
	taskRepo := NewTaskRepo(&storage)

	inputTask := models.Task{
		UserID:  uuid.MustParse("c8a2204b-5d87-488a-9bf3-cccef596981f"),
		Summary: "Non updated task",
	}

	// When
	err := taskRepo.Update(&inputTask)

	// Then
	if err == nil {
		t.Errorf("expecting an error because it's not possible to update a non existing task")
	}

}

// TestDelete Validates the correct behavior of Delete task when the record exist on the database.
func TestDelete(t *testing.T) {

	// Given
	storage := NewGormStorage()
	taskRepo := NewTaskRepo(&storage)

	inputTask := models.Task{
		UserID:  uuid.MustParse("d3ec726d-35f9-4468-9d12-82e5b71c0824"),
		Summary: "Non updated task",
	}

	err := taskRepo.Create(&inputTask)
	if err != nil {
		t.Fatalf("unexpected test error %v", err.Error())
	}

	// When
	err = taskRepo.Delete(inputTask.ID)

	// Then
	if err != nil {
		t.Errorf("unexpected test error, %v, %v", err.Error(), inputTask)
	}

	task, err := taskRepo.Get(inputTask.ID)
	if task != (models.Task{}) {
		t.Errorf("unexpected task value %v", task)
	}
	if err != nil {
		t.Errorf("unexpected error value %v", err.Error())
	}

}

// TestDeleteNonExisting Validates the correct behavior of Delete task when the record don't exist
func TestDeleteNonExisting(t *testing.T) {
	// Given
	storage := NewGormStorage()
	taskRepo := NewTaskRepo(&storage)

	// When
	err := taskRepo.Delete(uuid.MustParse("7bbe6cc7-70ac-4b3a-8f52-c5a0e0001071"))

	// Then
	if err != nil {
		t.Errorf("not expecting an error %v", err.Error())
	}
}

// TestClose Validates that the connection can be closed without panic
func TestClose(t *testing.T) {
	defer func() { recover() }()

	storage := NewGormStorage()
	storage.Close()

}
