# Slice 0 — Implementation Document

Local Kind cluster plus a known-good single-node ClickHouse.
No Go operator. No public API. No Postgres.

This document is the implementation spec for the first working slice. Follow it commit by commit. Do not invent extra packages, Helm charts, or an empty monorepo tree.

---

## 1. Goal

After Slice 0:

1. `make cluster-up` creates a reproducible local Kubernetes cluster.
2. `make cluster-down` deletes it.
3. A pinned official ClickHouse image runs as a StatefulSet, keeps data after pod deletion, and is reachable over in-cluster DNS.
4. The official ClickHouse operator is installed in a **study** namespace so CRDs can be inspected. It is not the product control plane.

Exit criterion: destroy the cluster, recreate it from the Makefile, query ClickHouse, delete the pod, confirm the data is still there, and explain every Kubernetes object in the ClickHouse manifests.

---

## 2. Non-goals

Do not do these in Slice 0:

- Go module, Kubebuilder, CRDs we own
- Keeper, replicas, shards
- Public API, Postgres, workers
- Helm as the product interface
- Ingress, cert-manager usage except as a dependency of the **reference** operator
- Prometheus stack, service mesh, custom CSI
- Bitnami / third-party ClickHouse charts
- Using the official operator as the product data plane

---

## 3. Prerequisites (host)

| Tool | Who installs it | Why |
|---|---|---|
| Docker | You. Daemon must already be running. | Kind node runtime |
| curl | You (almost always present). | Used by Make to fetch binaries |
| kind | **Make.** `cluster-up` downloads the CLI into `./bin`. | Cluster lifecycle |
| kubectl | **Make.** Same, into `./bin`. | Wait for nodes / later apply |
| ~8 GiB RAM free for Docker | — | 1 control-plane + 3 workers + ClickHouse |

Do not install kind or kubectl system-wide for this repo. `./bin` is gitignored.

Optional later (commit 5): still no Helm if the reference operator is installed from a pinned YAML release.

---

## 4. Makefile contract

These two targets are the human interface for the cluster. They stay Kind-lifecycle commands. Workloads get their own targets from commit 2 onward.

```text
make cluster-up      # create Kind cluster if missing, wait until all nodes are Ready
make cluster-down    # delete Kind cluster (idempotent if it does not exist)
make help            # list targets
```

### `cluster-up`

1. Install kind + kubectl into `./bin` if missing (pinned *tool* versions in the Makefile).
2. Verify Docker is reachable. Do not install Docker.
3. If cluster `tilmancloud` already exists, do **not** recreate it. Print that it exists and continue to the wait step.
4. If missing: `kind create cluster --name tilmancloud --config deploy/kind/cluster.yaml --wait 5m`
5. `kubectl wait --for=condition=Ready node --all --timeout=180s`
6. Print `kubectl get nodes -L topology.kubernetes.io/zone`

Node images in `cluster.yaml` are **not** pinned in commit 1. Kind uses the default image for the CLI it just installed. A `kindest/node:...@sha256:...` pin is only required when we need a specific Kubernetes version; it must match the kind CLI release notes.

Commit 1 stops here. From commit 2, after nodes are Ready, also run the apply target so a fresh `cluster-up` converges the lab to whatever the repo currently defines.

### `cluster-down`

1. `kind delete cluster --name tilmancloud`
2. Must succeed if the cluster is already gone (check `kind get clusters` first).

Do not add a “reset in place” flag in Slice 0. Recreate means `cluster-down` then `cluster-up`.

### Variables

```make
CLUSTER_NAME ?= tilmancloud
KIND_CONFIG  := deploy/kind/cluster.yaml
KIND_VERSION ?= v0.32.0          # CLI we download into ./bin
KUBECTL_VERSION ?= v1.36.1       # CLI we download into ./bin
```

Do **not** set `image:` on nodes in `cluster.yaml` for commit 1.

---

## 5. Pins (resolve digest at implement time if these move)

Record the exact digest in git. Do not use `latest`.

| Thing | Pin in commit 1? |
|---|---|
| kind CLI (Makefile → `./bin`) | Yes, `KIND_VERSION` |
| kubectl CLI (Makefile → `./bin`) | Yes, `KUBECTL_VERSION` |
| Kubernetes node image in `cluster.yaml` | **No.** Kind default for that CLI. |
| ClickHouse (later commit) | Yes, official image + digest at implement time |
| cert-manager (commit 6 only) | Pin a release YAML (e.g. v1.19.2), not `/latest/` |
| Official ClickHouse operator (commit 6 only) | Pin a **release tag** asset URL, not `/releases/latest/` |

If a pin is yanked, bump it in the same commit that changes the manifest, and note it in the commit message.

---

## 6. Commit stack

Skip the previously proposed “repo bootstrap” commit (README/ADR-template-only). `.gitignore` already ignores `tmp/`. The Makefile is born with the cluster.

Commits are stacked and non-isolated: each one leaves the previous targets working and adds the next. Conventional commits. Do not mix ClickHouse YAML into the Kind commit.

| # | Commit | Adds | Done when |
|---|---|---|---|
| 1 | `feat(kind): add cluster-up and cluster-down` | Makefile, Kind config | `make cluster-up` → 4 Ready nodes with zone labels; `make cluster-down` removes them |
| 2 | `feat(lab): add clickhouse-lab namespace` | Namespace + apply target | `make cluster-up` (or `make apply`) creates `clickhouse-lab` |
| 3 | `feat(clickhouse): deploy single-node official image` | STS, Services, ConfigMap, Secret | `clickhouse-client` inside the pod returns `1` |
| 4 | `test(clickhouse): verify persistence and DNS` | `make verify-clickhouse` | insert → delete pod → data remains; headless DNS resolves |
| 5 | `feat(reference): install official ClickHouse operator for study` | cert-manager + pinned operator | `kubectl get crd \| grep clickhouse.com` works. No product dependency on it |

Optional tiny README in commit 1 is allowed if it only documents `cluster-up` / `cluster-down`. Do not write the full architecture README.

---

## 7. Commit 1 — Kind cluster

### Files

```text
Makefile
deploy/kind/cluster.yaml
```

Optional: `docs/adr/001-local-kind-cluster.md` (one page: Kind, 3 workers, zone labels, control-plane NoSchedule). Write it only if it stays under ~40 lines. Skip an ADR template repo.

### `deploy/kind/cluster.yaml`

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: tilmancloud
nodes:
  - role: control-plane
    labels:
      topology.kubernetes.io/zone: eu-central-1a
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          taints:
            - key: node-role.kubernetes.io/control-plane
              effect: NoSchedule
  - role: worker
    labels:
      topology.kubernetes.io/zone: eu-central-1a
  - role: worker
    labels:
      topology.kubernetes.io/zone: eu-central-1b
  - role: worker
    labels:
      topology.kubernetes.io/zone: eu-central-1c
```

No `image:` field. Kind supplies the node image that matches the CLI in `./bin`.

Notes:

- Kind’s control-plane is **schedulable by default**. The taint is required so ClickHouse does not land on the control-plane. Without it, three workers are theater.
- Zone labels exist so later topology spread is real. Slice 0 does not yet spread replicas.
- No `extraPortMappings` yet. Access ClickHouse with `kubectl exec` / `port-forward`. Mapping 80/443 now implies Ingress we do not have.
- Do not add extraMounts, a local registry, or ingress-ready kubelet labels.

### Makefile shape

Use `.PHONY`, a `help` target that greps `##` comments, and fail fast:

```make
.PHONY: help cluster-up cluster-down

help: ## Show targets
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

cluster-up: ## Create the local Kind cluster and wait until nodes are Ready
	@# tool checks, then kind create if missing, then kubectl wait

cluster-down: ## Delete the local Kind cluster
	@# kind delete if present
```

Implementation details for `cluster-up`:

- Use `/bin/bash` (`SHELL := /bin/bash`) if the recipe needs `pipefail` / `set -euo pipefail`.
- Missing kind/kubectl: download into `./bin`. Missing Docker: print a one-line hint and `exit 1`.
- After create, assert node count is 4. If someone used a different Kind config, fail loudly.
- After Ready, print:

```text
kubectl get nodes -L topology.kubernetes.io/zone,node-role.kubernetes.io/control-plane
```

Expected: three workers `eu-central-1a|b|c`, control-plane tainted.

### Verify commit 1

```bash
make cluster-up
kubectl get nodes
# 4 Ready
make cluster-up          # second call is a no-op create, still waits/prints
make cluster-down
kind get clusters        # tilmancloud gone
make cluster-down        # still succeeds
```

Do not commit kubeconfig files.

---

## 8. Commit 2 — Lab namespace

### Files

```text
deploy/manifests/kustomization.yaml
deploy/manifests/namespace.yaml
Makefile   # add apply, and call it at the end of cluster-up
```

### Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: clickhouse-lab
  labels:
    app.kubernetes.io/part-of: tilmancloud
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/enforce-version: latest
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

Restricted PSS from the first namespace, so the ClickHouse pod in commit 3 must already be non-root. That is the point.

### Kustomize

Root overlay `deploy/manifests/kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - namespace.yaml
```

Later commits append resources here (or add a `clickhouse/` directory and list it).

### Makefile additions

```text
make apply          # kubectl apply -k deploy/manifests
```

`cluster-up` after nodes are Ready runs `$(MAKE) apply`.

`cluster-down` still only deletes Kind. The namespace dies with the cluster.

### Verify commit 2

```bash
make cluster-down
make cluster-up
kubectl get ns clickhouse-lab
kubectl get ns clickhouse-lab -o jsonpath='{.metadata.labels.pod-security\.kubernetes\.io/enforce}'
# restricted
```

### StorageClass note

Kind ships `standard` via local-path-provisioner. Do not install another provisioner. Document in a 5-line comment in `namespace.yaml` or in `deploy/kind/cluster.yaml`: PVCs use `standard`; this is not cloud disk durability.

No extra StorageClass manifest.

---

## 9. Commit 3 — Single-node ClickHouse

Reuse **only** the official image. Hand-written YAML. No Helm.

### Files

```text
deploy/manifests/clickhouse/kustomization.yaml
deploy/manifests/clickhouse/configmap.yaml
deploy/manifests/clickhouse/secret.yaml
deploy/manifests/clickhouse/service.yaml
deploy/manifests/clickhouse/headless-service.yaml
deploy/manifests/clickhouse/statefulset.yaml
deploy/manifests/kustomization.yaml   # add - clickhouse/
```

### Resource rules

| Object | Why it exists |
|---|---|
| ConfigMap | `config.d` listen + logger. Keep it tiny. No kitchen-sink `config.xml`. |
| Secret | Lab password for the default user. Obvious dummy value. Not production. |
| Service (ClusterIP) | Stable in-cluster client endpoint `clickhouse.clickhouse-lab.svc` on 8123/9000 |
| Headless Service | Per-pod DNS `clickhouse-0.clickhouse-headless.clickhouse-lab.svc.cluster.local`. Required for later replicas; include it now so the object is not a surprise. |
| StatefulSet | Stable identity `clickhouse-0` + `volumeClaimTemplates`. **Do not** also create a standalone PVC. |

### StatefulSet constraints

- `replicas: 1`
- `serviceName: clickhouse-headless`
- Selector labels: `app.kubernetes.io/name: clickhouse`, `app.kubernetes.io/instance: lab`
- Image: pinned digest of `clickhouse/clickhouse-server:26.3.20.7`
- `imagePullPolicy: IfNotPresent`
- Ports: `http` 8123, `native` 9000
- **Readiness and liveness:** HTTP `GET /ping` on 8123. Not TCP 9000. `/ping` returns `Ok.` when the server can serve.
- `volumeClaimTemplates`: `name: data`, mount `/var/lib/clickhouse`, `10Gi`, `ReadWriteOnce`, `storageClassName: standard`
- EmptyDir for `/var/log/clickhouse-server` is fine
- Mount ConfigMap at `/etc/clickhouse-server/config.d/`
- Password: official image env `CLICKHOUSE_PASSWORD` from Secret is acceptable for Slice 0. Do not bake XML users yet.

Security context (must pass namespace `restricted`):

```yaml
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 101
        runAsGroup: 101
        fsGroup: 101
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: clickhouse
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
```

Do not set `readOnlyRootFilesystem: true` in this slice. ClickHouse writes under `/var/lib/clickhouse` and the image layout is awkward with a fully read-only root. Revisit later with extra emptyDirs.

Resources (lab, not a production plan):

```yaml
resources:
  requests:
    cpu: 500m
    memory: 2Gi
  limits:
    cpu: "2"
    memory: 4Gi
```

If a laptop cannot tolerate 4Gi, lower the **limit** to 2Gi in the same commit and note it. Do not omit requests. Memory is the real constraint.

### ConfigMap (minimal)

`listen.xml`: listen on `0.0.0.0`.
Keep default logger. Do not disable default user networks in a way that blocks probes (`::/0` or `0.0.0.0/0` for default user in this lab).

### Secret

```yaml
stringData:
  password: clickhouse-lab
```

Label it clearly as lab-only. The verify script uses this password.

### Makefile

```text
make clickhouse-wait    # kubectl rollout status sts/clickhouse -n clickhouse-lab
```

`apply` already includes ClickHouse via kustomize. After `cluster-up` / `apply`, wait until the StatefulSet is ready (timeout ~3m). Either fold wait into `cluster-up` after apply, or document `make apply && make clickhouse-wait`. Prefer folding a wait into `cluster-up` once ClickHouse exists, so one command still means “lab is usable.”

### Verify commit 3

```bash
make cluster-down && make cluster-up
kubectl -n clickhouse-lab get sts,svc,pvc,cm,secret
kubectl -n clickhouse-lab exec sts/clickhouse -- \
  clickhouse-client --password clickhouse-lab --query "SELECT 1"
```

Must print `1`. Pod must not be on the control-plane node.

---

## 10. Commit 4 — Persistence and DNS verification

### Files

```text
hack/verify-clickhouse.sh
Makefile   # verify-clickhouse target
```

`hack/verify-clickhouse.sh` must be non-interactive, `set -euo pipefail`, namespace `clickhouse-lab`.

Steps:

1. Wait for `sts/clickhouse` ready.
2. `CREATE DATABASE IF NOT EXISTS lab`.
3. `CREATE TABLE IF NOT EXISTS lab.persist (id UInt32) ENGINE = MergeTree ORDER BY id`.
4. `INSERT INTO lab.persist VALUES (1), (2), (3)`.
5. Record `SELECT count() FROM lab.persist` (expect 3, or accumulate if re-run — better: `TRUNCATE` then insert so the script is idempotent).
6. `kubectl -n clickhouse-lab delete pod clickhouse-0`.
7. Wait for the StatefulSet to become ready again.
8. `SELECT count() FROM lab.persist` equals 3.
9. DNS: run a short-lived pod (e.g. `busybox:1.36`) in `clickhouse-lab` and `nslookup clickhouse-0.clickhouse-headless.clickhouse-lab.svc.cluster.local`. Restricted PSS may block random pods — give the probe pod the same non-root securityContext, or `kubectl exec` into `clickhouse-0` and `getent hosts clickhouse-0.clickhouse-headless`. Prefer **exec into the ClickHouse pod** to avoid PSS fights.
10. Exit 0 with a one-line summary. Exit non-zero on any mismatch.

```text
make verify-clickhouse   # runs hack/verify-clickhouse.sh
```

This commit is the Slice 0 exit criterion for **our** ClickHouse. Commit 5 is study-only.

### Verify commit 4

```bash
make cluster-down && make cluster-up
make verify-clickhouse
make verify-clickhouse   # second run still passes (idempotent truncate)
```

---

## 11. Commit 5 — Official operator as reference

Study workload. Product ClickHouse in `clickhouse-lab` must keep working without it.

### Why this commit exists

Before writing our own operator, inspect a real one: CRDs `clickhouseclusters.clickhouse.com` and `keeperclusters.clickhouse.com`, status, webhooks, MultiSTS if a sample is applied.

### Files

```text
deploy/reference/README.md          # how to inspect, not a product runbook
Makefile                            # apply-reference / optional
```

Do **not** vendor the full operator YAML into git if it is huge; pin a release URL in the Makefile. If the release asset is moderate, vendoring a pinned file under `deploy/reference/` is better for air-gapped reproducibility. Prefer vendoring a pinned `clickhouse-operator.yaml` **only if** size is acceptable; otherwise Makefile URL + version variable.

### Install order

1. cert-manager (pinned YAML). Wait for webhook to be Ready or later operator apply will race.
2. Official operator via:

```bash
kubectl apply --server-side --force-conflicts \
  -f https://github.com/ClickHouse/clickhouse-operator/releases/download/<TAG>/clickhouse-operator.yaml
```

Server-side apply is required; combined CRDs exceed client-side apply size.

3. Do **not** apply `examples/minimal.yaml` by default. A second ClickHouse + Keeper on Kind will starve RAM next to the lab STS. Inspection target is CRDs + operator pod:

```bash
kubectl get pods -n clickhouse-operator-system
kubectl get crd | grep clickhouse.com
kubectl explain clickhousecluster.spec
```

Optional documented extra: apply official `examples/minimal.yaml` into namespace `clickhouse-reference` when the machine has ≥16 GiB for Docker.

### Makefile

```text
make apply-reference     # cert-manager + operator, waits for operator pod
```

Do **not** fold this into `cluster-up`. Slice 0 daily loop is Kind + our single-node ClickHouse. The reference operator is opt-in so `cluster-up` stays fast and RAM-light.

### Verify commit 5

```bash
make cluster-up
make verify-clickhouse          # still passes
make apply-reference
kubectl get crd clickhouseclusters.clickhouse.com keeperclusters.clickhouse.com
make verify-clickhouse          # still passes; operator must not own our STS
```

Confirm our StatefulSet has no ownerReference to a ClickHouseCluster.

---

## 12. Target tree after Slice 0

```text
Makefile
README.md                          # optional, only cluster-up/down + apply + verify
deploy/
  kind/
    cluster.yaml
  manifests/
    kustomization.yaml
    namespace.yaml
    clickhouse/
      kustomization.yaml
      configmap.yaml
      secret.yaml
      service.yaml
      headless-service.yaml
      statefulset.yaml
  reference/
    README.md
    # optional vendored operator YAML
docs/
  adr/
    001-local-kind-cluster.md      # optional
hack/
  verify-clickhouse.sh
```

Still no `cmd/`, `api/`, `go.mod`, Helm chart, or `internal/`.

---

## 13. Daily loop after Slice 0

```bash
make cluster-up
make verify-clickhouse
# work
make cluster-down                 # when the VM needs RAM back
```

Reference operator only when studying CRDs:

```bash
make apply-reference
kubectl explain clickhousecluster
```

---

## 14. Implementation order inside each commit

1. Add files.
2. Run the verify steps for **that** commit (and previous `cluster-up` / `verify-clickhouse` if they already exist).
3. `git add` only those files. No `tmp/`.
4. Commit with the message in the table above.

If verification fails, fix in the same commit. Do not leave a red Makefile target on the branch.

---

## 15. Common failure modes

| Symptom | Likely cause |
|---|---|
| ClickHouse Pending / FailedCreate | control-plane taint missing and… actually that would still schedule. Pending: PVC / resources. Check `kubectl describe pod` |
| ClickHouse CreateContainerConfigError | Secret key name mismatch vs env |
| CrashLoop, permission denied on `/var/lib/clickhouse` | `fsGroup` / `runAsUser` not 101 |
| Pod rejected by PSS | missing `runAsNonRoot`, capabilities drop, or `seccompProfile` |
| Probe fail, server logs OK | probe using TCP 9000 or default user network restriction blocking 8123 |
| Data gone after delete pod | standalone PVC instead of `volumeClaimTemplates`, or emptyDir used for data |
| `cluster-up` recreates cluster every time | name mismatch between Makefile and `cluster.yaml` `name:` |
| Second `cluster-up` hangs | `--wait` on create of an existing cluster — skip create when it exists |
| Operator apply fails on CRD | client-side apply; use server-side apply |
| Operator never Ready | cert-manager not ready; webhook cert missing |
| Laptop OOM | 4 workers + 4Gi CH + operator. Drop CH limit to 2Gi; do not apply reference operator |

---

## 16. Learning checks (do not skip)

After commit 1, be able to explain:

- Kind node vs Kubernetes Node
- Why 3 workers if this slice has 1 replica
- Why the control-plane taint

After commit 3, be able to explain:

- Pod vs StatefulSet vs Deployment
- Why `volumeClaimTemplates` instead of a PVC object
- Service vs headless Service
- Why `/ping` not TCP
- Why user 101

After commit 4:

- What Kubernetes restart vs what ClickHouse recovery is
- What Kind local-path does **not** prove about cloud disks

---

## 17. After Slice 0 (not this document)

Slice 1: Kubebuilder, `ClickHouseService` `v1alpha1`, reconcile the **same** YAML shape (Service, ConfigMap, StatefulSet with volumeClaimTemplates). No extra PVC reconciler. Server-Side Apply. Leader election on.

Do not start Slice 1 until `make verify-clickhouse` is green on a cluster created only from this repo.
