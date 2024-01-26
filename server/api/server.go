package api

import (
	"context"
	"net/http"

	"github.com/SwarnimWalavalkar/aether/api/handlers"
	"github.com/SwarnimWalavalkar/aether/database"

	"github.com/gin-gonic/gin"
)

type Server struct {
	port   string
	gin    *gin.Engine
	server *http.Server
	db     *database.Database
}

func NewServer(port string, db *database.Database) *Server {
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
	}
}

func (s *Server) Start() error {
	s.gin.GET("/ping", handlers.Ping)
	s.gin.GET("/users/:uuid", handlers.GetUser(s.db))
	s.gin.POST("/users", handlers.CreateUser(s.db))
	s.gin.POST("/auth", handlers.GetAuthToken(s.db))

	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
