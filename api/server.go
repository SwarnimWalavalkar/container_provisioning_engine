package api

import (
	"context"
	"net/http"

	"github.com/SwarnimWalavalkar/container_provisioning_engine/api/handlers"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/database"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/middlewares"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/queue"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/services"

	"github.com/gin-gonic/gin"
)

type Server struct {
	port           string
	gin            *gin.Engine
	server         *http.Server
	db             *database.Database
	docker         *services.DockerService
	taskDispatcher *queue.TaskDispatcher
}

func NewServer(port string, db *database.Database, docker *services.DockerService, taskDispatcher *queue.TaskDispatcher) *Server {
	ginRouter := gin.New()

	ginRouter.Use(gin.Logger())
	ginRouter.Use(gin.Recovery())

	server := &http.Server{
		Addr:    port,
		Handler: ginRouter,
	}

	return &Server{
		port:           port,
		gin:            ginRouter,
		server:         server,
		db:             db,
		docker:         docker,
		taskDispatcher: taskDispatcher,
	}
}

func (s *Server) Start() error {

	s.gin.GET("/ping", handlers.Ping)

	v1 := s.gin.Group("/api/v1")

	{
		v1.GET("/users/:uuid", handlers.GetUser(s.db))
		v1.POST("/users", handlers.CreateUser(s.db))
		v1.POST("/auth", handlers.GetAuthToken(s.db))

		deployments := v1.Group("/deployments")

		deployments.Use(middlewares.AuthRequired)
		{
			deployments.GET("/", middlewares.AuthRequired, handlers.GetAllDeploymentsForUser(s.db))
			deployments.POST("/", middlewares.AuthRequired, handlers.CreateDeployment(s.db, s.docker, s.taskDispatcher))
			deployments.POST("/:uuid", middlewares.AuthRequired, handlers.UpdateDeployment(s.db, s.docker, s.taskDispatcher))

			deployments.GET("/:uuid", middlewares.AuthRequired, handlers.GetDeployment(s.db))
			deployments.DELETE("/:uuid", middlewares.AuthRequired, handlers.DeleteDeployment(s.db, s.docker, s.taskDispatcher))
		}
	}

	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
