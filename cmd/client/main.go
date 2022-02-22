package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/teitiago/task-manager-poc/internal/config"
	"github.com/teitiago/task-manager-poc/pkg/dto"
	"github.com/teitiago/task-manager-poc/pkg/logwrapper"
	"github.com/teitiago/task-manager-poc/pkg/rabbitmq"
	"go.uber.org/zap"
)

func main() {

	// init logger
	logwrapper.InitLogger()

	// Init RMQ Broker
	broker := rabbitmq.NewRabbitMQBroker("tasks", "tasks")
	defer broker.Close()

	broker.Consume(config.GetEnv("TASKS_COMPLETE_ROUTING", "tasks.completed"), func(payload []byte) bool {
		var task dto.CompletedTaskMessage
		if err := json.Unmarshal(payload, &task); err != nil {
			zap.L().Error("can't unmarshal payload", zap.Any("payload", payload), zap.Error(err))
			return false
		}
		completedDate := time.Unix(task.CompletedDate, 0).UTC()
		zap.L().Info("completed task", zap.String("task_id", task.ID), zap.String("user_id", task.ID), zap.String("completed_date", completedDate.Format("2006-01-02 15:04:05")))
		return true
	})

	// wait for signal termination
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	for {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			return
		case syscall.SIGTERM:
			return
		}
	}

}
