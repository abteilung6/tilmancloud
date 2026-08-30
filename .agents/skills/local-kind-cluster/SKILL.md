---
name: local-kind-cluster
description: >-
  Local Kind lab for tilmancloud. Use when running cluster-up/cluster-down/apply/env/clickhouse-wait,
  editing deploy/kind/cluster.yaml or deploy/manifests, adding Make targets,
  or installing kind/kubectl.
---

# Local Kind cluster

No README for Make. Extend this file when targets change.

## Make

From repo root. Prefer Make over raw `kind`/`kubectl`.

```bash
make cluster-up      # create tilmancloud if missing; retry create; wait for 4 Ready nodes; apply lab; wait for ClickHouse
make cluster-down    # delete cluster (idempotent)
make apply           # kubectl apply -k deploy/manifests
make clickhouse-wait # kubectl rollout status sts/clickhouse -n clickhouse-lab
make env             # print export PATH for ./bin; eval "$(make env)"
```

Make does not install CLIs. Recreate = down then up. `cluster-down` is Kind-delete only; the namespace and PVCs die with the cluster. `cluster-up` retries `kind create` up to 3 times because kubeadm can time out before the API server accepts the bootstrap ClusterRoleBinding. Control-plane `InitConfiguration.timeouts.kubernetesAPICall` is 4m (kubeadm default is 1m). After nodes are Ready, `cluster-up` runs `apply` then `clickhouse-wait` (10m — first image pull of the pinned ClickHouse digest is slow).

## CLIs

Need Docker running, plus `kind` and `kubectl` on PATH or in `./bin` (gitignored). Make prepends `./bin` for its recipes. For raw `kubectl` in the shell, `eval "$(make env)"` — do not document an absolute home path. If missing, install into `./bin` — do not add a Make download target:

- `os`: `linux` | `darwin`
- `arch`: `amd64` | `arm64` (`uname -m` `aarch64` → `arm64`)
- kind `v0.32.0`: `https://kind.sigs.k8s.io/dl/v0.32.0/kind-$(os)-$(arch)`
- kubectl `v1.36.1`: `https://dl.k8s.io/release/v1.36.1/bin/$(os)/$(arch)/kubectl`
- `chmod +x bin/kind bin/kubectl`

Do not pin `image:` in `cluster.yaml` until a Kubernetes SKU is required.

## Cluster

`deploy/kind/cluster.yaml`: 1 control-plane + 3 workers. Kind `role` = Node (Docker container with kubelet), not a Pod. Control-plane is `NoSchedule`. Zone labels are for later topology spread.

## Lab

`deploy/manifests`: kustomize overlay. Namespace `clickhouse-lab` has Pod Security `restricted` (`enforce`/`audit`/`warn`). `clickhouse/` is a single-node official `clickhouse-server` StatefulSet (digest-pinned `26.3.20.7`), ClusterIP + headless Services, ConfigMap listen on `0.0.0.0`, lab Secret password `clickhouse-lab`, uid 101. Kind already provides StorageClass `standard` (local-path); do not add another provisioner. The STS `volumeClaimTemplates` PVC uses `standard` — node-local disk, not cloud durability. Verify with `clickhouse-client --password clickhouse-lab --query "SELECT 1"` inside the pod. Pod must land on a worker.
