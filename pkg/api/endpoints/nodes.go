package endpoints

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/abteilung6/tilmancloud/pkg/api/generated"
	"github.com/abteilung6/tilmancloud/pkg/ec2"
	"github.com/abteilung6/tilmancloud/pkg/image"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-chi/chi/v5"
)

type NodesHandler struct {
	EC2Client ec2.EC2Client
	AMIFinder image.AMIFinder
}

func NewNodesHandler(ec2Client ec2.EC2Client, amiFinder image.AMIFinder) *NodesHandler {
	return &NodesHandler{
		EC2Client: ec2Client,
		AMIFinder: amiFinder,
	}
}

func (h *NodesHandler) CreateNode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	amiID, err := h.AMIFinder.FindLatestAMI(ctx)
	if err != nil {
		http.Error(w, "No AMI available. Please build an AMI first.", http.StatusServiceUnavailable)
		return
	}

	config := ec2.CreateInstanceConfig{
		ImageID:      amiID,
		InstanceType: types.InstanceTypeT4gMicro,
	}

	instanceInfo, err := ec2.CreateInstance(ctx, h.EC2Client, config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state := generated.NodeState(instanceInfo.State)
	response := generated.Node{
		Name:         instanceInfo.InstanceID,
		State:        &state,
		InstanceType: stringPtrOrNil(instanceInfo.InstanceType),
		PublicIp:     stringPtrOrNil(instanceInfo.PublicIP),
		PrivateIp:    stringPtrOrNil(instanceInfo.PrivateIP),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *NodesHandler) ListNodes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	instances, err := ec2.ListInstances(ctx, h.EC2Client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nodes := make([]generated.Node, 0, len(instances))
	for _, instanceInfo := range instances {
		state := generated.NodeState(instanceInfo.State)
		node := generated.Node{
			Name:         instanceInfo.InstanceID,
			State:        &state,
			InstanceType: stringPtrOrNil(instanceInfo.InstanceType),
			PublicIp:     stringPtrOrNil(instanceInfo.PublicIP),
			PrivateIp:    stringPtrOrNil(instanceInfo.PrivateIP),
		}
		nodes = append(nodes, node)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nodes)
}

func (h *NodesHandler) DeleteNode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	nodeId := chi.URLParam(r, "nodeId")

	if nodeId == "" {
		http.Error(w, "nodeId is required", http.StatusBadRequest)
		return
	}

	err := ec2.DeleteInstance(ctx, h.EC2Client, nodeId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "InvalidInstanceID") {
			http.Error(w, "Node not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
