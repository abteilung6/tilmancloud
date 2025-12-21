package endpoints

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abteilung6/tilmancloud/pkg/api/generated"
)

func TestHealthHandler_Health(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response generated.Health
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status == nil || *response.Status != "ok" {
		t.Errorf("expected status 'ok', got %v", response.Status)
	}
}

