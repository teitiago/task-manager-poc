package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/teitiago/task-manager-poc/api/docs"
	"github.com/teitiago/task-manager-poc/internal/config"
	"github.com/teitiago/task-manager-poc/internal/server/middleware"
	"go.uber.org/zap"
)

type Server struct {
	router      *gin.Engine
	taskService TaskService
}

func NewServer(router *gin.Engine, taskService TaskService) *Server {
	return &Server{
		router:      router,
		taskService: taskService,
	}
}

// @title Tasks API
// @version 1.0
// @host 0.0.0.0:8000
// @description Swagger API for Golang Project tasks.
// @BasePath /api/v1/
func (s *Server) routes() *gin.Engine {
	router := s.router

	taskHandler := NewTaskHandler(s.taskService)

	v1 := s.router.Group("/api/v1")
	{
		v1.GET(
			"/tasks",
			middleware.RequestIDMiddleware(),
			middleware.JWTMiddlewareExtract(),
			taskHandler.ListTasks,
			middleware.RequestIDMiddleware(),
		)
		v1.POST(
			"/tasks",
			middleware.RequestIDMiddleware(),
			middleware.JWTMiddlewareExtract(),
			middleware.RBACMiddleware("Technician"),
			taskHandler.CreateTask,
			middleware.RequestIDMiddleware(),
		)

		v1.GET(
			"/tasks/:taskID",
			middleware.RequestIDMiddleware(),
			middleware.JWTMiddlewareExtract(),
			taskHandler.GetTask,
			middleware.RequestIDMiddleware(),
		)
		v1.DELETE(
			"/tasks/:taskID",
			middleware.RequestIDMiddleware(),
			middleware.JWTMiddlewareExtract(),
			middleware.RBACMiddleware("Manager"),
			taskHandler.DeleteTask,
			middleware.RequestIDMiddleware(),
		)
		v1.PATCH("/tasks/:taskID",
			middleware.RequestIDMiddleware(),
			middleware.JWTMiddlewareExtract(),
			middleware.RBACMiddleware("Technician"),
			taskHandler.PatchTask,
			middleware.RequestIDMiddleware(),
		)

	}

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

func (s *Server) Run() error {
	// run function that initializes the routes
	r := s.routes()

	// run the server through the router
	err := r.Run(fmt.Sprintf(":%v", config.GetEnv("SERVER_PORT", "8000")))

	if err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}
