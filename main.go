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

	"github.com/SwarnimWalavalkar/container_provisioning_engine/api"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/database"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/queue"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/services"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

func main() {

	log.Println("Connecting to database...")

	db, err := database.NewDatabase()
	if err != nil {
		log.Println("Error connecting to the database", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		log.Fatal("Error pinging database", err)
	}

	log.Println("successfully connected to database")

	docker, err := services.NewDockerService()
	if err != nil {
		log.Println("Error connecting to Docker", err)
	}

	if err := docker.Ping(context.Background()); err != nil {
		log.Fatal("Error pinging Docker", err)
	}

	taskDispatcher := queue.NewTaskDispatcher(queue.Options{MaxWorkers: 10, MaxQueueSize: 100})
	taskDispatcherCTX, taskDispatcherCTXCancel := context.WithCancel(context.Background())

	go taskDispatcher.Start(taskDispatcherCTX)

	server := api.NewServer(fmt.Sprintf(":%s", os.Getenv("PORT")), db, docker, taskDispatcher)

	log.Println("Starting server...")

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP Server Error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println()
	log.Println("==============================")

	log.Println("Shutting down server...")

	log.Println("==============================")

	taskDispatcherCTXCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Stop(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println()
	log.Println("Goodbye...")
}
