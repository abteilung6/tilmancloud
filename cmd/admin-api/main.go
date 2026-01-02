package main

import (
	"context"
	"log"
	"net/http"

	"github.com/abteilung6/tilmancloud/pkg/api/endpoints"
	"github.com/abteilung6/tilmancloud/pkg/ec2"
	"github.com/abteilung6/tilmancloud/pkg/image"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	Router *chi.Mux
}

func CreateNewServer() (*Server, error) {
	server := &Server{}
	server.Router = chi.NewRouter()
	return server, nil
}

func MountHandlers(server *Server, nodesHandler *endpoints.NodesHandler, imagesHandler *endpoints.ImagesHandler, healthHandler *endpoints.HealthHandler) {
	// Middleware
	server.Router.Use(middleware.Logger)
	server.Router.Use(middleware.Recoverer)

	server.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	server.Router.Get("/health", healthHandler.Health)
	server.Router.Get("/nodes", nodesHandler.ListNodes)
	server.Router.Post("/nodes", nodesHandler.CreateNode)
	server.Router.Delete("/nodes/{nodeId}", nodesHandler.DeleteNode)
	server.Router.Get("/images", imagesHandler.ListImages)
}

func main() {
	ctx := context.Background()
	region := "eu-central-1"

	ec2Client, err := ec2.NewClient(ctx, region)
	if err != nil {
		log.Fatalf("Failed to create EC2 client: %v", err)
	}

	amiRegistrar, err := image.NewAMIRegistrar(ctx, region)
	if err != nil {
		log.Fatalf("Failed to create AMI registrar: %v", err)
	}

	nodesHandler := endpoints.NewNodesHandler(ec2Client, amiRegistrar)
	imagesHandler := endpoints.NewImagesHandler(amiRegistrar)
	healthHandler := endpoints.NewHealthHandler()

	server, err := CreateNewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	MountHandlers(server, nodesHandler, imagesHandler, healthHandler)

	port := ":8080"
	log.Printf("Admin API server starting on port %s", port)
	log.Printf("Health check available at http://localhost%s/health", port)
	if err := http.ListenAndServe(port, server.Router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
