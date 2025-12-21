package main

import (
	"context"
	"log"
	"net/http"

	"github.com/abteilung6/tilmancloud/pkg/api/endpoints"
	"github.com/abteilung6/tilmancloud/pkg/ec2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Router *chi.Mux
}

func CreateNewServer() (*Server, error) {
	server := &Server{}
	server.Router = chi.NewRouter()
	return server, nil
}

func MountHandlers(server *Server, nodesHandler *endpoints.NodesHandler, healthHandler *endpoints.HealthHandler) {
	// Middleware
	server.Router.Use(middleware.Logger)
	server.Router.Use(middleware.Recoverer)

	// Routes
	server.Router.Get("/health", healthHandler.Health)
	server.Router.Post("/nodes", nodesHandler.CreateNode)
}

func main() {
	ctx := context.Background()

	ec2Client, err := ec2.NewClient(ctx, "eu-central-1")
	if err != nil {
		log.Fatalf("Failed to create EC2 client: %v", err)
	}

	nodesHandler := endpoints.NewNodesHandler(ec2Client)
	healthHandler := endpoints.NewHealthHandler()

	server, err := CreateNewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	MountHandlers(server, nodesHandler, healthHandler)

	port := ":8080"
	log.Printf("Admin API server starting on port %s", port)
	log.Printf("Health check available at http://localhost%s/health", port)
	if err := http.ListenAndServe(port, server.Router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
