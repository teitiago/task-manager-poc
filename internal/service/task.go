package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/teitiago/task-manager-poc/internal/config"
	"github.com/teitiago/task-manager-poc/internal/server/middleware"
	"github.com/teitiago/task-manager-poc/pkg/dto"
	"github.com/teitiago/task-manager-poc/pkg/models"
	"go.uber.org/zap"
)

type taskRepo interface {
	Get(uuid.UUID) (models.Task, error)
	Filter(map[string]interface{}, models.Pagination, ...string) ([]*models.Task, error)
	Create(*models.Task) error
	Update(*models.Task) error
	Delete(id uuid.UUID) error
}

type publisher interface {
	Publish(interface{}, string, *sync.WaitGroup)
}

type encrypter interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}

type taskService struct {
	publisher  publisher
	repo       taskRepo
	encrypter  encrypter
	completeRK string
}

// getRequestID Collects the requestID from the provided context
func getRequestID(ctx context.Context) string {
	requestID, ok := ctx.Value(middleware.RequestID).(string)
	if !ok {
		return ""
	}
	return requestID
}

// NewTaskService Creates a new instance of Task service that will allow to apply the task
// business logic
func NewTaskService(repo taskRepo, publisher publisher, encrypter encrypter) *taskService {

	return &taskService{
		repo:       repo,
		publisher:  publisher,
		encrypter:  encrypter,
		completeRK: config.GetEnv("TASKS_COMPLETE_ROUTING", "tasks.completed"),
	}

}

// Get Validates that the task exists and the logged user can actually see the resource.
func (s *taskService) Get(ctx context.Context, id uuid.UUID) (models.Task, error) {

	requestID := getRequestID(ctx)
	user, err := NewUserInfo(ctx)
	if err != nil {
		zap.L().Error("no user info provided")
		return models.Task{}, errors.New("no user info")
	}

	// get task
	task, err := s.repo.Get(id)
	if err != nil {
		zap.L().Error("error getting task", zap.Error(err), zap.String("request_id", requestID))
		return models.Task{}, err
	}

	if task == (models.Task{}) {
		return task, nil
	}

	// validate rbac
	ok := user.validateTask(task)
	if !ok {
		zap.L().Error(
			"invalid access",
			zap.String("user_id", user.ID),
			zap.String("user_roles", user.roles),
			zap.String("request_id", requestID),
			zap.Any("task", task.ID),
		)
		return models.Task{}, errors.New("user cant access the resource")
	}
	task.Summary, err = s.encrypter.Decrypt(task.Summary)

	if err != nil {
		zap.L().Error(
			"decrypt error",
			zap.String("user_id", user.ID),
			zap.String("user_roles", user.roles),
			zap.String("request_id", requestID),
			zap.Any("task", task.ID),
		)
		return models.Task{}, err
	}
	return task, nil
}

// Filter Collects a list of tasks. This method forces to query only specific user_id tasks when it's not a manager to perform
// the query.
func (s *taskService) Filter(ctx context.Context, filter map[string]interface{}, page models.Pagination) ([]*models.Task, error) {

	requestID := getRequestID(ctx)
	user, err := NewUserInfo(ctx)
	if err != nil {
		zap.L().Error("no user info provided")
		return nil, errors.New("no user info")
	}

	// TODO: better page validation
	if page == (models.Pagination{}) {
		zap.L().Error("page must not be empty", zap.String("request_id", requestID))
		return nil, errors.New("invalid page")
	}
	if filter == nil {
		filter = make(map[string]interface{})
	}

	// When not a manager force to query only the user id tasks
	if !strings.Contains(user.roles, "Manager") {
		zap.L().Debug("forcing user_id filter", zap.String("request_id", requestID), zap.Any("user_id", user.ID))
		filter["user_id"] = user.ID
	}

	// get tasks
	tasks, err := s.repo.Filter(filter, page, "id", "user_id", "created_at", "updated_at", "completed_date")
	if err != nil {
		zap.L().Error("error getting tasks", zap.Error(err), zap.String("request_id", requestID))
		return nil, err
	}

	return tasks, nil

}

// Create creates a new task on the database. It validates the rbac and encrypt the summary
// to hide possible PII. This will ignore any provided completed_date because we cant create
// completed tasks
func (s *taskService) Create(ctx context.Context, task models.Task) (models.Task, error) {

	requestID := getRequestID(ctx)
	user, err := NewUserInfo(ctx)
	if err != nil {
		zap.L().Error("no user info provided")
		return models.Task{}, errors.New("no user info")
	}

	// validate role access
	if !strings.Contains(user.roles, "Technician") {
		zap.L().Error("only Technician can create tasks", zap.String("request_id", requestID))
		return models.Task{}, errors.New("invalid role")
	}

	// 0 value any provided completed date
	task.CompletedDate.Valid = false
	task.CompletedDate.Time = time.Time{}

	// enforce user_id from jwt
	// TODO: Review this if manager can actually create tasks on behalf of technicians
	task.UserID, err = uuid.Parse(user.ID)
	if err != nil {
		zap.L().Error("invalid user id convert", zap.Error(err), zap.String("request_id", requestID))
		return models.Task{}, err
	}

	// encrypt summary possible pii data
	task.Summary, err = s.encrypter.Encrypt(task.Summary)
	if err != nil {
		// don't log the summary to hide PII
		zap.L().Error("error encrypting summary", zap.String("request_id", requestID), zap.Error(err))
		return models.Task{}, err
	}

	// create the task on the repo
	err = s.repo.Create(&task)
	if err != nil {
		zap.L().Error("error persisting task", zap.String("request_id", requestID), zap.Error(err))
		return models.Task{}, err
	}

	return task, nil

}

// Patch updates a given task. It is possible to update the task summary and the completed date.
// The completed date is only possible to update if it's null.
// Only technicians can update their tasks and can only update their own tasks
func (s *taskService) Patch(ctx context.Context, id uuid.UUID, request models.Task) (models.Task, error) {

	requestID := getRequestID(ctx)
	user, err := NewUserInfo(ctx)
	if err != nil {
		zap.L().Error("no user info provided")
		return models.Task{}, errors.New("no user info")
	}

	// validate role access
	if !strings.Contains(user.roles, "Technician") {
		zap.L().Error("only Technician can update tasks", zap.String("request_id", requestID))
		return models.Task{}, errors.New("invalid role")
	}

	// get the task to update
	task, err := s.repo.Get(id)
	if err != nil {
		zap.L().Error("error getting task", zap.Error(err), zap.Any("task_id", id), zap.String("request_id", requestID))
		return models.Task{}, err
	}
	if task == (models.Task{}) {
		zap.L().Error("cant update task that don't exist", zap.Error(err), zap.Any("task_id", id), zap.String("request_id", requestID))
		return models.Task{}, errors.New("task don't exist")
	}

	// check if user can update the task
	if task.UserID.String() != user.ID {
		zap.L().Error(
			"user is trying to update a task that don't own",
			zap.String("request_id", requestID),
			zap.Any("task_id", id),
			zap.String("user_id", user.ID),
		)
		return models.Task{}, errors.New("invalid task owner")
	}

	// validate the update request
	if request.Summary != "" {
		task.Summary, err = s.encrypter.Encrypt(task.Summary)
		if err != nil {
			// don't log the summary to hide PII
			zap.L().Error("error encrypting summary", zap.String("request_id", requestID), zap.Error(err))
			return models.Task{}, err
		}
	}
	if task.CompletedDate.Valid {
		zap.L().Error(
			"can't change a completed date",
			zap.String("request_id", requestID),
			zap.Any("task_id", id),
		)
		return models.Task{}, errors.New("task is completed")
	}

	wg := &sync.WaitGroup{}
	notify := false
	if request.CompletedDate.Valid && request.CompletedDate.Time.UTC().Unix() > time.Now().UTC().Unix() {
		zap.L().Error("cant complete tasks in the future", zap.Any("completed_date", task.CompletedDate.Time.String()), zap.String("request_id", requestID))
		return models.Task{}, errors.New("completed date in the future")
	} else if request.CompletedDate.Valid {
		notify = true
		task.CompletedDate = request.CompletedDate
		wg.Add(1)
		go s.publisher.Publish(
			dto.CompletedTaskMessage{
				ID:            task.ID.String(),
				UserID:        task.ID.String(),
				CompletedDate: task.CompletedDate.Time.UTC().Unix(),
			}, s.completeRK, wg)
	}

	// perform the update
	err = s.repo.Update(&task)
	if err != nil {
		zap.L().Error("error updating task", zap.Any("task_id", id), zap.String("request_id", requestID))
		return models.Task{}, err
	}

	// wait for the publisher
	if notify {
		wg.Wait()
	}

	return task, nil

}

// Delete Deletes a given task from the database.
// This method evaluates if the user can delete the task.
func (s *taskService) Delete(ctx context.Context, id uuid.UUID) error {

	requestID := getRequestID(ctx)
	user, err := NewUserInfo(ctx)
	if err != nil {
		zap.L().Error("no user info provided")
		return errors.New("no user info")
	}

	// validate role access
	if !strings.Contains(user.roles, "Manager") {
		zap.L().Error("only Manager can delete tasks", zap.String("request_id", requestID))
		return errors.New("invalid role")
	}

	err = s.repo.Delete(id)
	if err != nil {
		zap.L().Error("can't delete task", zap.Any("task_id", id), zap.Error(err))
	}
	return err

}
