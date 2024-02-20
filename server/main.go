package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SwarnimWalavalkar/aether/api"
	"github.com/SwarnimWalavalkar/aether/database"
	"github.com/SwarnimWalavalkar/aether/services"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

func main() {

	fmt.Println("Connecting to database...")

	db, err := database.NewDatabase()
	if err != nil {
		fmt.Println("Error connecting to the database", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		log.Fatal("Error pinging database", err)
	}

	fmt.Println("successfully connected to database")

	docker, err := services.NewDockerService()
	if err != nil {
		fmt.Println("Error connecting to Docker", err)
	}

	if err := docker.Ping(context.Background()); err != nil {
		log.Fatal("Error pinging Docker", err)
	}

	server := api.NewServer(fmt.Sprintf(":%s", os.Getenv("PORT")), db, docker)

	fmt.Println("Starting server...")

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP Server Error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Stop(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Goodbye...")
}
