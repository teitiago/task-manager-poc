package dto

// TaskCreateRequest Is the object that needs to be provided in order to create
// a new task.
type TaskCreateRequest struct {
	Summary string `json:"summary" binding:"required,max=2500"`
}

// TaskCreateResponse Is the object that needs to be provided as a response
// to a create task attempt.
type TaskCreateResponse struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status" binding:"required"`
}

// TaskPatchRequest is the request needed to update a given task.

type TaskPatchRequest struct {
	Summary       string `json:"summary,omitempty" binding:"max=2500"`
	CompletedDate int64  `json:"completed_date,omitempty"`
}

// TaskSingleResponse is the dto to be used on the filter method
type TaskSingleResponse struct {
	ID            string `json:"task_id"`
	UserID        string `json:"user_id,omitempty"`
	CreatedAt     int64  `json:"created_at"`
	ModifiedAt    int64  `json:"modified_at"`
	CompletedDate int64  `json:"completed_date,omitempty"`
}

// TaskResponse is the whole task information
type TaskResponse struct {
	TaskSingleResponse
	Summary string `json:"summary"`
}

// TaskFilterResponse is the response sent to the client when it filtered the tasks
type TaskFilterResponse struct {
	Status string               `json:"status"`
	Tasks  []TaskSingleResponse `json:"tasks,omitempty"`
}

// CompletedTaskMessage Is the message sent to the consumer saying that a message
// was completed.
type CompletedTaskMessage struct {
	ID            string `json:"task_id"`
	UserID        string `json:"user_id"`
	CompletedDate int64  `json:"completed_date"`
}
