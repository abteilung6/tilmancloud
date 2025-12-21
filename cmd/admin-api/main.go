package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Router *chi.Mux
}

type HealthResponse struct {
	Status string `json:"status"`
}

func CreateNewServer() *Server {
	server := &Server{}
	server.Router = chi.NewRouter()
	return server
}

func MountHandlers(server *Server) {
	// Middleware
	server.Router.Use(middleware.Logger)
	server.Router.Use(middleware.Recoverer)

	// Routes
	server.Router.Get("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	server := CreateNewServer()
	MountHandlers(server)

	port := ":8080"
	log.Printf("Admin API server starting on port %s", port)
	log.Printf("Health check available at http://localhost%s/health", port)
	if err := http.ListenAndServe(port, server.Router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
