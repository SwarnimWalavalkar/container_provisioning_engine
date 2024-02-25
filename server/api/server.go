package api

import (
	"context"
	"net/http"

	"github.com/SwarnimWalavalkar/aether/api/handlers"
	"github.com/SwarnimWalavalkar/aether/database"
	"github.com/SwarnimWalavalkar/aether/middlewares"
	"github.com/SwarnimWalavalkar/aether/services"

	"github.com/gin-gonic/gin"
)

type Server struct {
	port   string
	gin    *gin.Engine
	server *http.Server
	db     *database.Database
	docker *services.DockerService
}

func NewServer(port string, db *database.Database, docker *services.DockerService) *Server {
	ginRouter := gin.New()

	ginRouter.Use(gin.Logger())
	ginRouter.Use(gin.Recovery())

	server := &http.Server{
		Addr:    port,
		Handler: ginRouter,
	}

	return &Server{
		port:   port,
		gin:    ginRouter,
		server: server,
		db:     db,
		docker: docker,
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
			deployments.POST("/", middlewares.AuthRequired, handlers.CreateDeployment(s.db, s.docker))
			deployments.POST("/:uuid", middlewares.AuthRequired, handlers.UpdateDeployment(s.db, s.docker))

			deployments.GET("/:uuid", middlewares.AuthRequired, handlers.GetDeployment(s.db))
			deployments.DELETE("/:uuid", middlewares.AuthRequired, handlers.DeleteDeployment(s.db, s.docker))
		}
	}

	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
