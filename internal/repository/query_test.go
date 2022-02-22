//go:build unit

package repository

import (
	"reflect"
	"sync"
	"testing"
)

// TestValidAddQuery Validates the correct behavior of AddQuery
// It evaluates the query builder and that it can be used on multiple goroutines.
func TestValidAddQuery(t *testing.T) {

	// Given
	queriesMapInput := map[string]interface{}{
		"user_id":            "test_id",
		"completed_date_gte": "1589979600",
		"completed_date_gt":  int64(1589979601),
		"completed_date_lte": int32(1589979602),
		"completed_date_lt":  1589979603,
	}
	expectedQuery := []Query{
		{Field: "user_id", Operator: "=", Value: "test_id"},
		{Field: "completed_date", Operator: ">=", Value: "2020-05-20 13:00:00"},
		{Field: "completed_date", Operator: ">", Value: "2020-05-20 13:00:01"},
		{Field: "completed_date", Operator: "<=", Value: "2020-05-20 13:00:02"},
		{Field: "completed_date", Operator: "<", Value: "2020-05-20 13:00:03"},
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(queriesMapInput))
	tq, _ := NewTaskQuery(len(queriesMapInput))

	for k, v := range queriesMapInput {
		// When
		go tq.AddQuery(k, v, wg)
	}
	wg.Wait()

	// Then
	if len(expectedQuery) != len(tq.queries) {
		t.Errorf("expected %v, got %v", len(expectedQuery), len(tq.queries))
	}

	var exist bool
	for i := 0; i < len(expectedQuery); i++ {
		exist = false
		for j := 0; j < len(tq.queries); j++ {
			if reflect.DeepEqual(expectedQuery[i], tq.queries[j]) {
				exist = true
				break
			}
		}
		if !exist {
			t.Errorf("expected %v but not on the array %v", expectedQuery[i], tq.queries)
		}
	}

}

// TestAddInvalidDate Validates that there is an error when providing an invalid date to the query.
func TestAddInvalidDate(t *testing.T) {

	tq, _ := NewTaskQuery(2)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	tq.AddQuery("completed_date_gte", "2020-05-20", wg)
	if tq.err == nil {
		t.Error("expecting an error")
	}
	tq.AddQuery("completed_date_te", "2020-05-20", wg)
	if tq.err == nil {
		t.Error("expecting an error")
	}

}

// TestAddQueryInvalidParam Validates the correct behavior of AddQuery when an invalid param is provided.
func TestAddQueryInvalidParam(t *testing.T) {

	tq, _ := NewTaskQuery(1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	tq.AddQuery("invalid", 1, wg)
	if tq.err == nil {
		t.Error("expecting an error")
	}

}

// TestAddQueryInvalidSize Validates the correct behavior of AddQuery when
// trying to add more queries than the expected.
func TestAddQueryInvalidSize(t *testing.T) {

	tq, err := NewTaskQuery(-1)
	if err == nil {
		t.Errorf("expecting an error")
	}

	tq, _ = NewTaskQuery(1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	tq.AddQuery("user_id", 1, wg)
	if tq.err != nil {
		t.Errorf("unexpected error, %v", tq.err.Error())
	}

	wg.Add(1)
	tq.AddQuery("user_id", 12, wg)
	if tq.err == nil {
		t.Error("expecting an error")
	}

}
