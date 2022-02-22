package repository

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/teitiago/task-manager-poc/internal/config"
	"github.com/teitiago/task-manager-poc/pkg/models"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	"go.uber.org/zap"
)

var once sync.Once
var gormInstance gormStorage

// getMysqlDSN Builds the DSN string to connect to MySQL
func getMysqlDSN() string {
	dbAddr := config.GetEnv("DB_ADDR", "127.0.0.1")
	dbUser := config.GetEnv("DB_USER", "mysql")
	dbPass := config.GetEnv("DB_PWD", "mysql")
	dbPort := config.GetEnv("DB_PORT", "3306")
	dbName := config.GetEnv("DB_NAME", "tasks")

	return fmt.Sprintf(
		"%v:%v@tcp(%v:%v)/%v?charset=Latin1&parseTime=True&loc=Local",
		dbUser,
		dbPass,
		dbAddr,
		dbPort,
		dbName,
	)

}

type gormStorage struct {
	db *gorm.DB
}

// NewGormStorage creates a new instance of gorm storage integration.
// Since gorm is thread safe (https://v1.gorm.io/docs/method_chaining.html)
// only one connection will be created.
// If more than one DB is used use a map for dsn.
func NewGormStorage() gormStorage {

	once.Do(func() {
		conn, err := gorm.Open("mysql", getMysqlDSN())
		if err != nil {
			panic(err)
		}
		zap.L().Info("successfully connected to database using gorm")
		gormInstance = gormStorage{db: conn}
	})
	return gormInstance

}

// Migrate migrates the given database table.
func (s *gormStorage) Migrate(instance interface{}) error {
	err := s.db.AutoMigrate(instance).Error
	if err != nil {
		panic(err)
	}
	return nil
}

// Close closes the connection to the db.
func (s *gormStorage) Close() {
	err := s.db.Close()
	if err != nil {
		panic(err)
	}
}

// Get Collects the record with the given id and stores it on the provided
// instance.
func (s *gormStorage) Get(id uuid.UUID, instance interface{}) error {
	zap.L().Debug("get", zap.Any("id", id))
	err := s.db.Where("id = ?", id.String()).
		First(instance).
		Error

	// Ignore not found errors
	// let the downstram handle this.
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		zap.L().Info(
			"accessing a non existing db",
			zap.String("id", id.String()),
			zap.Any("object", instance),
		)
		return nil
	}

	return err
}

// Filter allows to filter for a given set of tasks.
// It requires a query array to filter with complex opperators
// and strings to filter the needed fields
func (s *gormStorage) Filter(filter []Query, instance interface{}, pagination models.Pagination, fields ...string) error {
	values := make([]interface{}, len(filter))
	whereClause := make([]string, len(filter))
	for i, query := range filter {
		values[i] = query.Value
		whereClause[i] = fmt.Sprintf("%v %v ?", query.Field, query.Operator)
	}

	whereString := strings.Join(whereClause[:], " AND ")
	offset := (pagination.Page - 1) * pagination.Limit

	var err error
	if len(whereString) > 0 {
		err = s.db.Select(fields).Where(whereString, values...).Limit(pagination.Limit).Offset(offset).Order(pagination.Sort).Find(instance).Error
	} else {
		err = s.db.Select(fields).Limit(pagination.Limit).Offset(offset).Order(pagination.Sort).Find(instance).Error
	}

	return err

}

// Create creates a new entry on the database given a specific instance.
func (s *gormStorage) Create(instance interface{}) error {
	err := s.db.Create(instance).Error
	return err
}

// Save updates a new entry on the database given a specific instance.
func (s *gormStorage) Save(instance interface{}) error {
	err := s.db.Save(instance).Error
	return err
}

// Delete deletes an entry from the DB.
func (s *gormStorage) Delete(instance interface{}) error {
	return s.db.Delete(instance).Error
}
