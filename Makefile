SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

CLUSTER_NAME ?= tilmancloud
KIND_CONFIG ?= deploy/kind/cluster.yaml
MANIFESTS ?= deploy/manifests
CLUSTER_CREATE_ATTEMPTS ?= 3
CLICKHOUSE_NS ?= clickhouse-lab
CLICKHOUSE_STS ?= clickhouse
CERT_MANAGER_VERSION ?= v1.19.2
CERT_MANAGER_URL ?= https://github.com/cert-manager/cert-manager/releases/download/$(CERT_MANAGER_VERSION)/cert-manager.yaml
CLICKHOUSE_OPERATOR_VERSION ?= v0.0.7
CLICKHOUSE_OPERATOR_URL ?= https://github.com/ClickHouse/clickhouse-operator/releases/download/$(CLICKHOUSE_OPERATOR_VERSION)/clickhouse-operator.yaml
CLICKHOUSE_OPERATOR_NS ?= clickhouse-operator-system

# Prefer project-local binaries if present; Make does not download them.
export PATH := $(CURDIR)/bin:$(PATH)

.PHONY: help check cluster-up cluster-down apply env clickhouse-wait verify-clickhouse verify-managed apply-reference apply-headlamp operator-generate operator-run operator-test apply-operator

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

verify-clickhouse: check ## Insert rows, delete the pod, confirm data and headless DNS
	CLICKHOUSE_NS=$(CLICKHOUSE_NS) CLICKHOUSE_STS=$(CLICKHOUSE_STS) ./scripts/verify-clickhouse.sh

verify-managed: check ## Same check on the operator-owned ClickHouse; fail if the lab STS was adopted
	./scripts/verify-managed.sh

apply-reference: check ## Install cert-manager and the official ClickHouse operator (study only)
	CERT_MANAGER_URL=$(CERT_MANAGER_URL) \
		CLICKHOUSE_OPERATOR_URL=$(CLICKHOUSE_OPERATOR_URL) \
		CLICKHOUSE_OPERATOR_NS=$(CLICKHOUSE_OPERATOR_NS) \
		./scripts/apply-reference.sh

apply-headlamp: check ## Install Headlamp UI (study only; not cluster-up)
	kubectl apply -k deploy/reference/headlamp
	kubectl wait --for=condition=Available deploy/headlamp -n headlamp --timeout=10m
	@echo "Headlamp token (paste in the UI):"
	kubectl -n headlamp create token headlamp --duration=24h
	@echo "Access: kubectl -n headlamp port-forward svc/headlamp 8080:80"
	@echo "then open http://127.0.0.1:8080"

# controller-gen via go tool (see go.mod). Does not download into ./bin.
operator-generate: ## Generate CRD and DeepCopy from api/v1alpha1
	go tool controller-gen object paths=./api/...
	go tool controller-gen crd paths=./api/... output:crd:artifacts:config=deploy/operator/crd

operator-run: ## Run the operator against the current kubeconfig
	go run ./cmd/operator --leader-elect=true

operator-test: ## Test Reconcile with the same Request the manager passes in
	go test ./internal/controller/ -count=1

OPERATOR_IMAGE ?= tilmancloud-operator:dev

apply-operator: check ## Build the operator image, load it into Kind, apply CRD and Deployment
	docker build -t $(OPERATOR_IMAGE) .
	kind load docker-image $(OPERATOR_IMAGE) --name $(CLUSTER_NAME)
	kubectl apply -k deploy/operator
	kubectl wait --for=condition=Available deploy/clickhouseservice -n tilmancloud-system --timeout=5m

env: ## Print export PATH so the shell can use ./bin
	@echo 'export PATH="$(CURDIR)/bin:$$PATH"'

cluster-down: check ## Delete the local Kind cluster
	kind delete cluster --name $(CLUSTER_NAME)
