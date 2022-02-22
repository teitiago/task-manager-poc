//go:build integration

package rabbitmq

import (
	"strings"
	"sync"
	"testing"
)

// TestPublishConsume Validates the correct behavior of publishing
// a message and receive it on the correct queue.
func TestPublishConsume(t *testing.T) {

	msg := `{"task_id": "894233df-7756-43db-a246-4e2171ccef4d", completed_date: 1495285200}`
	routingKey := "tasks.completed"

	broker := NewRabbitMQBroker("task_exch", "task_queue")
	defer broker.Close()

	// Submit message
	wg := &sync.WaitGroup{}
	wg.Add(1)
	broker.Publish(msg, routingKey, wg)
	wg.Wait()

	// Receive message
	go broker.Consume(routingKey, func(payload []byte) bool {
		escapedPayload := strings.ReplaceAll(string(payload), `\"`, `"`)
		escapedPayload = strings.TrimSuffix(escapedPayload, `"`)
		escapedPayload = strings.TrimPrefix(escapedPayload, `"`)
		if escapedPayload != msg {
			t.Errorf("expected %v got %v", msg, escapedPayload)
		}
		return true
	})

}
