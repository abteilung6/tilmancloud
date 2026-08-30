SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

CLUSTER_NAME ?= tilmancloud
KIND_CONFIG ?= deploy/kind/cluster.yaml
MANIFESTS ?= deploy/manifests
CLUSTER_CREATE_ATTEMPTS ?= 3
CLICKHOUSE_NS ?= clickhouse-lab
CLICKHOUSE_STS ?= clickhouse

# Prefer project-local binaries if present; Make does not download them.
export PATH := $(CURDIR)/bin:$(PATH)

.PHONY: help check cluster-up cluster-down apply env clickhouse-wait

help: ## Show targets
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

check:
	@command -v kind >/dev/null && command -v kubectl >/dev/null || { \
		echo "kind and kubectl required on PATH or in ./bin"; \
		exit 1; \
	}
	@docker info >/dev/null 2>&1 || { echo "Docker daemon is not running."; exit 1; }

cluster-up: check ## Create the local Kind cluster, apply lab manifests, wait for ClickHouse
	@if kind get clusters 2>/dev/null | grep -qx '$(CLUSTER_NAME)'; then \
		echo "cluster $(CLUSTER_NAME) already exists"; \
	else \
		ok=0; \
		for i in $$(seq 1 $(CLUSTER_CREATE_ATTEMPTS)); do \
			echo "creating cluster $(CLUSTER_NAME) (attempt $$i/$(CLUSTER_CREATE_ATTEMPTS))"; \
			if kind create cluster --name $(CLUSTER_NAME) --config $(KIND_CONFIG) --wait 5m; then \
				ok=1; \
				break; \
			fi; \
			kind delete cluster --name $(CLUSTER_NAME) >/dev/null 2>&1 || true; \
		done; \
		if [ "$$ok" -ne 1 ]; then \
			echo "failed to create cluster $(CLUSTER_NAME) after $(CLUSTER_CREATE_ATTEMPTS) attempts"; \
			exit 1; \
		fi; \
	fi
	kubectl wait --for=condition=Ready node --all --timeout=180s
	@count="$$(kubectl get nodes --no-headers | wc -l | tr -d ' ')"; \
	if [ "$$count" -ne 4 ]; then \
		echo "expected 4 nodes, got $$count"; \
		exit 1; \
	fi
	kubectl get nodes -L topology.kubernetes.io/zone,node-role.kubernetes.io/control-plane
	@$(MAKE) apply
	@$(MAKE) clickhouse-wait

apply: check ## Apply lab manifests with kustomize
	kubectl apply -k $(MANIFESTS)

clickhouse-wait: check ## Wait until the lab ClickHouse StatefulSet is Ready
	kubectl rollout status sts/$(CLICKHOUSE_STS) -n $(CLICKHOUSE_NS) --timeout=10m

env: ## Print export PATH so the shell can use ./bin
	@echo 'export PATH="$(CURDIR)/bin:$$PATH"'

cluster-down: check ## Delete the local Kind cluster
	kind delete cluster --name $(CLUSTER_NAME)
