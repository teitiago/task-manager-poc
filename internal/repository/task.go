package repository

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/teitiago/task-manager-poc/pkg/models"
)

var validTaskFields = [...]string{"id", "user_id", "created_at", "updated_at", "completed_date"}

// taskRepo is the repo that allows the interaction between the application layer
// and the storage layer.
// storage is the object that actually performs the interaction.
// TODO: Create an interface for store and allow other ORM
type taskRepo struct {
	storage Storage
}

// NewTaskRepo creates a new taskRepo instance.
// Storage is the db storage implementation that allows the interaction with the DB
func NewTaskRepo(storage Storage) *taskRepo {
	err := storage.Migrate(&models.Task{})
	if err != nil {
		panic(err)
	}
	return &taskRepo{storage: storage}
}

// Get Collects a given task by ID
func (repo *taskRepo) Get(id uuid.UUID) (models.Task, error) {
	var task models.Task

	err := repo.storage.Get(id, &task)
	return task, err
}

// Filter allows to filter for a given set of tasks.
func (repo *taskRepo) Filter(filter map[string]interface{}, pagination models.Pagination, fields ...string) ([]*models.Task, error) {
	var tasks []*models.Task

	// validate the fields
	for _, field := range fields {
		if !isFieldValid(field) {
			return nil, errors.New("invalid query field provided")
		}
	}

	// create the taskquery
	filterLen := len(filter)
	taskQuery, err := NewTaskQuery(filterLen)
	if err != nil {
		return nil, err
	}

	// build the query
	wg := &sync.WaitGroup{}
	wg.Add(filterLen)
	for k, v := range filter {
		go taskQuery.AddQuery(k, v, wg)
	}
	wg.Wait()

	if taskQuery.err != nil {
		return nil, taskQuery.err
	}

	err = repo.storage.Filter(taskQuery.queries, &tasks, pagination, fields...)
	return tasks, err
}

// Create Stores a given task instance on the database.
func (repo *taskRepo) Create(task *models.Task) error {
	return repo.storage.Create(task)
}

// Update Updates a given task instance on the database.
func (repo *taskRepo) Update(task *models.Task) error {
	return repo.storage.Save(task)
}

// Delete Deletes a given record from the database.
func (repo *taskRepo) Delete(id uuid.UUID) error {
	var task models.Task
	task.ID = id
	return repo.storage.Delete(task)
}

// isFieldValid Validates if the field can be used as part of a filter query.
func isFieldValid(field string) bool {
	for i := 0; i < len(validTaskFields); i++ {
		if validTaskFields[i] == field {
			return true
		}
	}
	return false
}
