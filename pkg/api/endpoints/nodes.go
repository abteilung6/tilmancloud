package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/abteilung6/tilmancloud/pkg/api/generated"
	"github.com/abteilung6/tilmancloud/pkg/ec2"
)

type NodesHandler struct {
	EC2Client ec2.EC2Client
}

func NewNodesHandler(ec2Client ec2.EC2Client) *NodesHandler {
	return &NodesHandler{
		EC2Client: ec2Client,
	}
}

func (h *NodesHandler) CreateNode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	instanceID, err := ec2.CreateInstance(ctx, h.EC2Client)
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
