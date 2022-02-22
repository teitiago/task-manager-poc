package repository

import (
	"github.com/google/uuid"
	"github.com/teitiago/task-manager-poc/pkg/models"
)

// Storage is the common interface to interact with the database
type Storage interface {
	Migrate(interface{}) error
	Close()
	Get(uuid.UUID, interface{}) error
	Filter([]Query, interface{}, models.Pagination, ...string) error
	Create(interface{}) error
	Save(interface{}) error
	Delete(interface{}) error
}

// query represents a common way to perform queries
// on a storage
type Query struct {
	Field    string
	Operator string
	Value    interface{}
}

type QueryBuilder interface {
	AddQuery(string, interface{}) (Query, error)
}
