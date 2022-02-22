package server

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/teitiago/task-manager-poc/internal/server/middleware"
	"github.com/teitiago/task-manager-poc/pkg/dto"
	"github.com/teitiago/task-manager-poc/pkg/models"
	"go.uber.org/zap"
)

// TaskService is the interface that the business layer needs to follow in order
// for this handler to properly work.
type TaskService interface {
	Get(context.Context, uuid.UUID) (models.Task, error)
	Filter(context.Context, map[string]interface{}, models.Pagination) ([]*models.Task, error)
	Create(context.Context, models.Task) (models.Task, error)
	Patch(context.Context, uuid.UUID, models.Task) (models.Task, error)
	Delete(context.Context, uuid.UUID) error
}

// getRequestID Collects the request trace id from the middleware.
func getRequestID(c *gin.Context) string {

	value, exists := c.Get(middleware.RequestID)
	if !exists {
		return ""
	}
	return value.(string)

}

// getTaskID Collect the task id from the context
// if the task id is not a valid uuid false is returned, otherwise is returned the uuid and true
func getTaskID(requestID string, c *gin.Context) (uuid.UUID, bool) {
	taskIDParam := c.Param("taskID")
	taskID, err := uuid.Parse(taskIDParam)
	if err != nil {
		zap.L().Error("invalid task id format", zap.String("task_id", taskIDParam), zap.String("request_id", requestID))
		c.AbortWithStatus(http.StatusNotFound)
		return uuid.UUID{}, false
	}
	return taskID, true
}

// convertTaskToDTO Converts a task to a DTO object
func convertTaskToDTO(task models.Task) dto.TaskSingleResponse {
	var response dto.TaskSingleResponse

	response.CreatedAt = task.CreatedAt.Unix()
	response.ModifiedAt = task.UpdatedAt.Unix()
	if task.UserID != (uuid.UUID{}) {
		response.UserID = task.UserID.String()
	}
	if task.CompletedDate.Valid {
		response.CompletedDate = task.CompletedDate.Time.UTC().Unix()
	}
	response.ID = task.ID.String()
	return response
}

// taskHandler Is the handler that will be the interface between rest users and the business layer.
type taskHandler struct {
	service TaskService
}

// NewTaskHandler Creates a new instance of the task handler
func NewTaskHandler(service TaskService) *taskHandler {
	return &taskHandler{service: service}
}

// TODO: Add sort parameter

// @Summary Get a specific task
// @Description Collects the whole task information
// @Tags task
// @Param Authorization header string true "Bearer"
// @Param user_id query string false  "name search by user_id"
// @Param completed_date_gt query int false  "date to search for completed tasks"
// @Param completed_date_lt query int false  "date to search for completed tasks"
// @Param created_at_gt query int false  "date to search for created tasks"
// @Param created_at_lt query int false  "date to search for created tasks"
// @Param modified_at_gt query int false  "date to search for modified tasks"
// @Param modified_at_lt query int false  "date to search for modified tasks"
// @Param page query int false  "results page"
// @Param limit query int false  "results limit"
// @Produce json
// @Success 200 {object} dto.TaskFilterResponse
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tasks [get]
func (handler *taskHandler) ListTasks(c *gin.Context) {
	filter := make(map[string]interface{})

	// TODO: REFACTOR USE GIN QUERY BINDING (https://github.com/gin-gonic/gin/issues/742#issuecomment-264681292)
	// Maybe remove task query
	var err error
	limit := 10
	page := 1
	sort := "created_at asc"
	query := c.Request.URL.Query()
	for key, value := range query {
		queryValue := value[len(value)-1]
		switch key {
		case "limit":
			limit, err = strconv.Atoi(queryValue)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, dto.TaskFilterResponse{Status: "invalid limit value"})
				return
			}
		case "page":
			page, err = strconv.Atoi(queryValue)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, dto.TaskFilterResponse{Status: "invalid page value"})
				return
			}
		case "sort":
			if !strings.HasSuffix(queryValue, " asc") && !strings.HasSuffix(queryValue, " desc") {
				c.AbortWithStatusJSON(http.StatusBadRequest, dto.TaskFilterResponse{Status: "invalid sort suffix must be asc or desc"})
				return
			}
		default:
			filter[key] = queryValue
		}
	}

	pagination := models.Pagination{
		Limit: limit,
		Page:  page,
		Sort:  sort,
	}

	// evaluate service response
	tasks, err := handler.service.Filter(c, filter, pagination)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if tasks == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// convert tasks to dto
	taskLen := len(tasks)
	responseTasks := make([]dto.TaskSingleResponse, taskLen)
	for i, task := range tasks {
		responseTasks[i] = convertTaskToDTO(*task)
	}

	c.IndentedJSON(http.StatusOK, dto.TaskFilterResponse{Status: "OK", Tasks: responseTasks})

}

// @Summary Get a specific task
// @Description Collects the whole task information
// @Tags task
// @Param Authorization header string true "Bearer"
// @Param task_id path string  true  "the task identifier to collect"
// @Produce json
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tasks/{task_id} [get]
func (handler *taskHandler) GetTask(c *gin.Context) {
	requestID := getRequestID(c)
	taskID, ok := getTaskID(requestID, c)
	if !ok {
		zap.L().Error("no taskID provided", zap.String("request_id", requestID))
		return
	}

	task, err := handler.service.Get(c, taskID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	singleResponse := convertTaskToDTO(task)
	c.IndentedJSON(http.StatusOK, dto.TaskResponse{
		TaskSingleResponse: singleResponse,
		Summary:            task.Summary,
	})

}

// @Summary Creates a new task
// @Description Creates a new task
// @Tags task
// @Param Authorization header string true "Bearer"
// @Accept json
// @Produce json
// @Param message body dto.TaskCreateRequest true "The task body to be created"
// @Success 201 {object} dto.TaskCreateResponse
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tasks [post]
func (handler *taskHandler) CreateTask(c *gin.Context) {
	requestID := getRequestID(c)

	var taskRequest dto.TaskCreateRequest

	// bind request
	if err := c.BindJSON(&taskRequest); err != nil {
		zap.L().Error(err.Error(), zap.String("request_id", requestID))
		c.IndentedJSON(http.StatusBadRequest, dto.TaskCreateResponse{Status: "error while binding request"})
		return
	}

	// convert request to task
	task := models.Task{Summary: taskRequest.Summary}

	// handle service response
	if newTask, err := handler.service.Create(c, task); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, dto.TaskCreateResponse{Status: "task creation failed"})
	} else {
		c.IndentedJSON(http.StatusCreated, dto.TaskCreateResponse{ID: newTask.ID.String(), Status: "task created"})
	}

}

// @Summary Deletes a task
// @Description Deletes a given task
// @Tags task
// @Param Authorization header string true "Bearer"
// @Param task_id path string  true  "the task identifier to be deleted"
// @Success 204
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tasks/{task_id} [delete]
func (handler *taskHandler) DeleteTask(c *gin.Context) {
	requestID := getRequestID(c)

	taskID, ok := getTaskID(requestID, c)
	if !ok {
		return
	}

	err := handler.service.Delete(c, taskID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.AbortWithStatus(http.StatusNoContent)

}

// @Summary Updates a task
// @Description Updates a given task
// @Tags task
// @Param Authorization header string true "Bearer"
// @Param task_id path string  true  "the task identifier to updated"
// @Accept json
// @Produce json
// @Param message body dto.TaskPatchRequest true "The task body to be created"
// @Success 204
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tasks/{task_id} [patch]
func (handler *taskHandler) PatchTask(c *gin.Context) {
	requestID := getRequestID(c)
	taskID, ok := getTaskID(requestID, c)
	if !ok {
		return
	}

	// bind request
	var request dto.TaskPatchRequest
	if err := c.BindJSON(&request); err != nil {
		zap.L().Error(err.Error(), zap.String("request_id", requestID))
		c.IndentedJSON(http.StatusBadRequest, dto.TaskCreateResponse{Status: "error while binding request"})
		return
	}

	// is there something to patch
	if request.Summary == "" && request.CompletedDate == int64(0) {
		zap.L().Error("nothing to update", zap.String("request_id", requestID))
		c.IndentedJSON(http.StatusBadRequest, dto.TaskCreateResponse{Status: "nothing to patch"})
		return
	}

	var task models.Task

	if request.Summary != "" {
		task.Summary = request.Summary
	}

	// validate and convert the timestamp to sqlTime
	if request.CompletedDate != int64(0) {
		requestCompletedDate := int64(request.CompletedDate)
		if requestCompletedDate < 0 {
			zap.L().Error("timestamp invalid", zap.Any("timestamp", request.CompletedDate), zap.String("request_id", requestID))
			c.IndentedJSON(http.StatusBadRequest, dto.TaskCreateResponse{Status: "invalid timestamp"})
			return
		}
		if requestCompletedDate > time.Now().Unix() {
			zap.L().Error("timestamp in the future", zap.Any("timestamp", request.CompletedDate), zap.String("request_id", requestID))
			c.IndentedJSON(http.StatusBadRequest, dto.TaskCreateResponse{Status: "invalid timestamp in the future"})
			return
		}
		completedDate := sql.NullTime{
			Time:  time.Unix(int64(requestCompletedDate), 0).UTC(),
			Valid: true,
		}
		task.CompletedDate = completedDate
	}

	// perform the patch on the service
	reponseTask, err := handler.service.Patch(c, taskID, task)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// empty response no model was updated
	if reponseTask == (models.Task{}) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// updated and nothing to report
	c.AbortWithStatus(http.StatusNoContent)

}
