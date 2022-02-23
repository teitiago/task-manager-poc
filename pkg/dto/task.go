package dto

// ==================
// REST
// ==================

// TaskCreateRequest object to create a new task.
// @Description Request to create a new task
type TaskCreateRequest struct {
	Summary string `json:"summary" binding:"required,max=2500" example:"My task description"` // Description of a task containing at most 2500 chars
}

// TaskCreateResponse object returned as a response of creating a task.
// @Description Response to create a new task
type TaskCreateResponse struct {
	ID     string `json:"id,omitempty" example:"74531653-252b-48c7-b562-63e82f5e3466"` // New task identifier following an uuid pattern
	Status string `json:"status" binding:"required" example:"OK"`                      // The response status, that can be OK or error
}

// TaskPatchRequest object to update a task.
// @Description Request to update a task. The summary can be updated as well as the compledDate (that will notify the managers)
type TaskPatchRequest struct {
	Summary       string `json:"summary,omitempty" binding:"max=2500" example:"A brand new task summary"` // New task summary
	CompletedDate int64  `json:"completed_date,omitempty" example:"1645606033"`                           // The timestamp in seconds when the task was completed
}

// TaskSingleResponse is the dto to be used to represent a single response.
// @Description Response object sent when collecting tasks.
type TaskSingleResponse struct {
	ID            string `json:"task_id" example:"74531653-252b-48c7-b562-63e82f5e3466"`           // The task identifier (To be removed)
	UserID        string `json:"user_id,omitempty" example:"74531653-252b-48c7-b562-63e82f5e3466"` // The task owner (that might not be provided)
	CreatedAt     int64  `json:"created_at" example:"1645606033"`                                  // The timestamp in seconds when the task was created
	ModifiedAt    int64  `json:"modified_at" example:"1645606033"`                                 // The timestamp in seconds when the task was last updated
	CompletedDate int64  `json:"completed_date,omitempty" example:"1645606035"`                    // The timestamp when the task was completed
}

// TaskResponse extends the task single response with the summary information. Used when getting a specific task.
// @Description Response object sent when getting a specific task.
type TaskResponse struct {
	TaskSingleResponse
	Summary string `json:"summary" example:"My task summary"` // The task summary
}

// TaskFilterResponse is the response sent when filtering tasks
// @Description Response object sent when filtering specific tasks
type TaskFilterResponse struct {
	Status string               `json:"status" example:"OK"` // Filter status response
	Tasks  []TaskSingleResponse `json:"tasks,omitempty"`     // List of collected tasks
}

// TODO: A query object that allow to filter tasks, removing the need of internal query process on the repository

// ==================
// Messages
// ==================

// CompletedTaskMessage Is the message sent to the consumer saying that a message
// was completed.
type CompletedTaskMessage struct {
	ID            string `json:"task_id"`
	UserID        string `json:"user_id"`
	CompletedDate int64  `json:"completed_date"`
}
