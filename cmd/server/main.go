package main

import (
	"io"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/teitiago/task-manager-poc/internal/repository"
	"github.com/teitiago/task-manager-poc/internal/server"
	"github.com/teitiago/task-manager-poc/internal/service"
	"github.com/teitiago/task-manager-poc/pkg/encryption"
	"github.com/teitiago/task-manager-poc/pkg/logwrapper"
	"github.com/teitiago/task-manager-poc/pkg/rabbitmq"
)

func main() {

	// init logger
	logwrapper.InitLogger()

	// Init Storage
	storage := repository.NewGormStorage()
	defer storage.Close()
	repo := repository.NewTaskRepo(&storage)

	// Init RMQ Broker
	broker := rabbitmq.NewRabbitMQBroker("tasks", "tasks")
	defer broker.Close()

	// Encrypt
	encrypt := encryption.NewAESEncryption()

	// Service
	taskService := service.NewTaskService(repo, broker, encrypt)

	// Server
	router := gin.Default()
	router.Use(cors.Default())
	gin.DefaultWriter = io.MultiWriter(os.Stdout)

	serv := server.NewServer(router, taskService)
	err := serv.Run()
	if err != nil {
		panic(err)
	}
}
