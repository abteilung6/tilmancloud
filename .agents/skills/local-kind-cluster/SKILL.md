---
name: local-kind-cluster
description: >-
  Local Kind lab for tilmancloud. Use when running cluster-up/cluster-down,
  editing deploy/kind/cluster.yaml, adding Make targets, or installing kind/kubectl.
---

# Local Kind cluster

No README for Make. Extend this file when targets change.

## Make

From repo root. Prefer Make over raw `kind`/`kubectl`.

```bash
make cluster-up      # create tilmancloud if missing; wait for 4 Ready nodes
make cluster-down    # delete cluster (idempotent)
```

Make does not install CLIs. Recreate = down then up. `cluster-down` is Kind-delete only.

## CLIs

Need Docker running, plus `kind` and `kubectl` on PATH or in `./bin` (gitignored). If missing, install into `./bin` — do not add a Make download target:

- `os`: `linux` | `darwin`
- `arch`: `amd64` | `arm64` (`uname -m` `aarch64` → `arm64`)
- kind `v0.32.0`: `https://kind.sigs.k8s.io/dl/v0.32.0/kind-$(os)-$(arch)`
- kubectl `v1.36.1`: `https://dl.k8s.io/release/v1.36.1/bin/$(os)/$(arch)/kubectl`
- `chmod +x bin/kind bin/kubectl`

Do not pin `image:` in `cluster.yaml` until a Kubernetes SKU is required.

## Cluster

`deploy/kind/cluster.yaml`: 1 control-plane + 3 workers. Kind `role` = Node (Docker container with kubelet), not a Pod. Control-plane is `NoSchedule`. Zone labels are for later topology spread.
