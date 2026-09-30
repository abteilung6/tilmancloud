# Glossary

Words we use in this repo. Short meanings only. Kubernetes short names in parentheses.

## Cluster

**Kind** — Tool that runs a Kubernetes cluster as Docker containers on your machine. Cluster name here: `tilmancloud`.

**Node** — A machine in the cluster (here: one Docker container with a kubelet). Not a Pod.

**kube-apiserver** — The Kubernetes HTTP API. `kubectl` talks to this. Custom types live under `/apis/<group>/<version>/…`.

**etcd** — Where the API server stores objects after you apply them.

**Namespace** — A name prefix so two things can both be called `clickhouse`. Ours: `clickhouse-lab` (hand-written YAML), `clickhouse-managed` (our Custom Resource), `clickhouse-operator-system` (official operator process).

**kubectl apply** — Send a YAML file to the API server (create or update). Does not start ClickHouse by itself.

## Workloads and networking

**Pod** — One or more containers scheduled together. ClickHouse runs in a Pod named `clickhouse-0`.

**StatefulSet** (`sts`) — Controller that gives Pods stable names (`clickhouse-0`) and a volume per Pod. The lab ClickHouse is a StatefulSet.

**Deployment** — Controller for stateless replicas (operator process, Headlamp). Not used for ClickHouse data.

**Service** (`svc`) — Stable DNS name and ports in front of Pods. **ClusterIP** has a virtual IP. **Headless** (`clusterIP: None`) gives DNS for each Pod (`clickhouse-0.clickhouse-headless.…`).

**ConfigMap** — Non-secret files mounted into the Pod (here: `listen.xml`).

**Secret** — Same idea for the password.

**PersistentVolumeClaim** — A request for disk. The StatefulSet creates `data-clickhouse-0` via **volumeClaimTemplates**. We do not create that claim by hand.

**StorageClass `standard`** — Kind’s local-path disk. Survives deleting the Pod, not `cluster-down`.

**ownerReferences** — “This object is owned by that one.” If a Custom Resource owns a StatefulSet, deleting the Custom Resource can delete the StatefulSet. The lab StatefulSet has none.

**Probe** — kubelet HTTP check. ClickHouse readiness is `GET /ping` on port 8123.

## Our API

**CustomResourceDefinition** — Teaches the API server a new type. File under `deploy/operator/crd/`. Generated from Go types.

**Custom Resource** — One instance of that type. Example: `ClickHouseService` named `clickhouse` in `clickhouse-managed`.

**ClickHouseService** (`chs`) — Our type (`clickhouse.tilmancloud.io/v1alpha1`). Identity is name + namespace. `spec` is empty for now.

**apiVersion** — `group/version`, e.g. `clickhouse.tilmancloud.io/v1alpha1`. Not a ClickHouse server version.

**v1alpha1** — First API version. We may change the schema later.

**spec** — What you asked for. **status** — What the controller observed. `Ready` follows the owned StatefulSet.

**sample-request.yaml** — Example `kubectl apply` body for a ClickHouseService. Stores the Custom Resource only.

**Source of truth (schema)** — Go structs in `api/v1alpha1/`. The CustomResourceDefinition YAML is generated from them.

## Controllers

**Controller / reconciler** — A loop: see a Custom Resource, make the cluster match. Ours applies ConfigMap, Secret, Services, and StatefulSet in `clickhouse-managed`, not in `clickhouse-lab`.

**Reconcile** — One turn of that loop. Input is a **Request**: namespace + name only. The function then `Get`s the current ClickHouseService.

**Server-Side Apply** — Tell the API server the whole desired object and let it merge fields. We use `Patch(..., client.Apply)` instead of Create-then-Update.

**FieldOwner** — Name stamped on fields we last wrote (`clickhouse-service`). Other owners can write other fields without a fight unless we set **ForceOwnership**.

**Owns** — Watch child kinds (ConfigMap, Secret, Service, StatefulSet) so a change to a child runs Reconcile again.

**operator-run** — `go run ./cmd/operator` using your kubeconfig. Laptop inner loop.

**apply-operator** — Docker build + `kind load` + Deployment in `tilmancloud-system`. Does not create a ClickHouseService. Not part of `cluster-up`.

**tilmancloud-system** — Namespace for our operator process. Not `clickhouse-operator-system`.

**Operator** — A controller plus the types it owns. “Our operator” ≠ ClickHouse Inc’s operator.

**ClickHouseCluster** (`chc`) — Official ClickHouse Inc type (`clickhouse.com`). Installed for study. We have not created one.

**ClickHouseInstallation (CHI)** — Altinity’s type. Not installed.

**Leader election** — Only one controller replica writes at a time.

## Go tooling

**Kubebuilder** — Layout and comment conventions (`PROJECT`, `// +kubebuilder:`). We wrote those by hand; we do not need the Kubebuilder CLI.

**controller-gen** — Reads those comments and writes the CustomResourceDefinition, DeepCopy code, and ClusterRole. `make operator-generate`.

**controller-runtime** — Go library that will run the watch loop (`operator-run`). Already a module dependency.

## Other installs (opt-in)

**cert-manager** — Issues TLS certs for the official operator’s webhooks. Not used by our API.

**Headlamp** — Cluster web UI. Needs `kubectl port-forward` and a token.

**`make apply-reference`** — Installs cert-manager + the official operator. Not part of `cluster-up`.

**`make verify-managed`** — Applies the sample ClickHouseService if missing, then persist + headless DNS on `clickhouse-managed`. Fails if the lab StatefulSet has ownerReferences.
