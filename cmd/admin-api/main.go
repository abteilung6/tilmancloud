package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/abteilung6/tilmancloud/pkg/api/generated"
	"github.com/abteilung6/tilmancloud/pkg/ec2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Router    *chi.Mux
	EC2Client ec2.EC2Client
}

type HealthResponse struct {
	Status string `json:"status"`
}

func CreateNewServer() (*Server, error) {
	server := &Server{}
	server.Router = chi.NewRouter()

	ctx := context.Background()
	ec2Client, err := ec2.NewClient(ctx, "eu-central-1")
	if err != nil {
		return nil, err
	}
	server.EC2Client = ec2Client

	return server, nil
}

func MountHandlers(server *Server) {
	// Middleware
	server.Router.Use(middleware.Logger)
	server.Router.Use(middleware.Recoverer)

	// Routes
	server.Router.Get("/health", healthHandler)
	server.Router.Post("/nodes", server.createNodeHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) createNodeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	instanceID, err := ec2.CreateInstance(ctx, s.EC2Client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := generated.Node{
		Name: &instanceID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func main() {
	server, err := CreateNewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	MountHandlers(server)

	port := ":8080"
	log.Printf("Admin API server starting on port %s", port)
	log.Printf("Health check available at http://localhost%s/health", port)
	if err := http.ListenAndServe(port, server.Router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
