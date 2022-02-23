package repository

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

var validTaskQueryFields = [...]string{
	"user_id",
	"completed_date_gte", "completed_date_gt",
	"completed_date_lte", "completed_date_lt",
	"created_at_gte", "created_at_gt",
	"created_at_lte", "created_at_lt",
}

type taskQuery struct {
	queries []Query
	sync.Mutex
	intCounter int
	err        error
}

// NewTaskQuery Builds a new task query to be used on the
// storage interface.
// NOTE: THIS MUST BE REFACTORED TO USE THE OPTIONS PATTERN (https://www.sohamkamani.com/golang/options-pattern/)
func NewTaskQuery(querySize int) (*taskQuery, error) {

	if querySize < 0 {
		return nil, errors.New("invalid querySize provided")
	}

	queries := make([]Query, querySize)
	return &taskQuery{queries: queries, intCounter: 0, err: nil}, nil

}

// AddQuery Creates a Task query based on the provided field
// and its suffix.
func (q *taskQuery) AddQuery(param string, value interface{}, wg *sync.WaitGroup) {

	// set function to done
	defer wg.Done()

	// error already reported
	if q.err != nil {
		return
	}

	if q.intCounter >= len(q.queries) {
		q.err = errors.New("reached query limit")
		return
	}

	if !q.isParamValid(param) {
		q.err = errors.New("invalid query parameter provided")
		return
	}

	var operator string
	var newParam string
	switch suffix := param[len(param)-3:]; suffix {
	case "gte":
		operator = ">="
		newParam = strings.Replace(param, "_gte", "", -1)
	case "_gt":
		operator = ">"
		newParam = strings.Replace(param, "_gt", "", -1)
	case "lte":
		operator = "<="
		newParam = strings.Replace(param, "_lte", "", -1)
	case "_lt":
		operator = "<"
		newParam = strings.Replace(param, "_lt", "", -1)
	default:
		operator = "="
		newParam = param
	}

	// TODO: refactor this hammer! Or the query itself!
	if newParam != "user_id" {
		var i int64
		var err error
		switch v := value.(type) {
		case string:
			i, err = strconv.ParseInt(value.(string), 10, 64)
		case int:
			i = int64(v)
		case int32:
			i = int64(v)
		case int64:
			i = v
		}

		if err != nil {
			zap.L().Error("error converting date value", zap.Error(err), zap.Any("value", value))
			q.err = err
			return
		} else {
			value = time.Unix(i, 0).UTC().Format(("2006-01-02 15:04:05"))
		}

	}

	// Avoid race conditions
	q.Lock()
	defer q.Unlock()

	q.queries[q.intCounter] = Query{Field: newParam, Operator: operator, Value: value}

	q.intCounter++
}

// isParamValid Validates if a given param exists on the fields.
func (q *taskQuery) isParamValid(param string) bool {
	for _, field := range validTaskQueryFields {
		if param == field {
			return true
		}
	}
	return false
}
