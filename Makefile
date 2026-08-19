SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

CLUSTER_NAME ?= tilmancloud
KIND_CONFIG ?= deploy/kind/cluster.yaml

# Prefer project-local binaries if present; Make does not download them.
export PATH := $(CURDIR)/bin:$(PATH)

.PHONY: help check cluster-up cluster-down

help: ## Show targets
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

check:
	@command -v kind >/dev/null && command -v kubectl >/dev/null || { \
		echo "kind and kubectl required on PATH or in ./bin"; \
		exit 1; \
	}
	@docker info >/dev/null 2>&1 || { echo "Docker daemon is not running."; exit 1; }

cluster-up: check ## Create the local Kind cluster and wait until nodes are Ready
	@if kind get clusters 2>/dev/null | grep -qx '$(CLUSTER_NAME)'; then \
		echo "cluster $(CLUSTER_NAME) already exists"; \
	else \
		kind create cluster --name $(CLUSTER_NAME) --config $(KIND_CONFIG) --wait 5m; \
	fi
	kubectl wait --for=condition=Ready node --all --timeout=180s
	@count="$$(kubectl get nodes --no-headers | wc -l | tr -d ' ')"; \
	if [ "$$count" -ne 4 ]; then \
		echo "expected 4 nodes, got $$count"; \
		exit 1; \
	fi
	kubectl get nodes -L topology.kubernetes.io/zone,node-role.kubernetes.io/control-plane

cluster-down: check ## Delete the local Kind cluster
	kind delete cluster --name $(CLUSTER_NAME)
