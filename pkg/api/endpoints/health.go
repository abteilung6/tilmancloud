package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/abteilung6/tilmancloud/pkg/api/generated"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	response := generated.Health{
		Status: &status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

