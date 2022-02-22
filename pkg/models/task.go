package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// Task Represents a task that needs to be completed.
// Base contains model base data such the internal id and created and updated datetimes (DB audit purposes)
// UserID is the user_id that owns this specific task
// Summary is the task description that might contain PII information
// CompletedDate is the date when the task was completed.
type Task struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`

	UserID        uuid.UUID    `gorm:"type:char(36);column:user_id"`
	Summary       string       `gorm:"type:varchar(2500)"`
	CompletedDate sql.NullTime `gorm:"column:completed_date"`
}

// BeforeCreate ensures the ID is built when creating the record.
func (t *Task) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}
