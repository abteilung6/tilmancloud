# ClickHouse Managed Cloud on Kubernetes
## Engineering Guide, Architecture Reference, and Learning Roadmap

**Purpose:** Build an end-to-end managed ClickHouse platform in Go, backed by Kubernetes, with a custom Kubernetes operator and a public cloud-style API.

**Primary goals:**
1. Learn Kubernetes by building against its real control-plane primitives.
2. Learn Go through production-style APIs, controllers, persistence, concurrency, and testing.
3. Demonstrate deep ClickHouse operational knowledge.
4. Build a portfolio project that can support interviews for database, infrastructure, platform, and distributed-systems roles.
5. Evolve the project from a local operator into a small managed ClickHouse provider.

---

# 1. Executive Architecture Decision

Build this as a **monorepo** containing several independently runnable binaries:

- `api`: public REST API for customers.
- `operator`: Kubernetes controller manager.
- `worker`: asynchronous control-plane operations.
- `ctl`: optional CLI for the public API.

The public API owns the **customer-facing product model**.

The operator owns the **Kubernetes/database execution model**.

Do not expose Kubernetes concepts directly through the public API.

The first implementation may let the API/worker write Custom Resources directly to one Kubernetes cluster. The long-term architecture should preserve a boundary so that the same public API can later target multiple workload clusters, cloud providers, or BYOC environments.

```text
                         CUSTOMER / CLI / TERRAFORM
                                  |
                                  v
                         +-------------------+
                         |   Public Go API   |
                         |      /v1/...      |
                         +---------+---------+
                                   |
                         desired product state
                                   |
                          +--------v--------+
                          | Control Plane DB |
                          |   PostgreSQL     |
                          +--------+--------+
                                   |
                            async operations
                                   |
                          +--------v--------+
                          | Provisioner /   |
                          | Workflow Worker |
                          +--------+--------+
                                   |
                              placement
                                   |
                    +--------------+--------------+
                    |                             |
                    v                             v
             Kubernetes Cluster A          Kubernetes Cluster B
                    |                             |
                    v                             v
             ClickHouseService CR          ClickHouseService CR
                    |                             |
                    v                             v
               Go Operator                    Go Operator
                    |                             |
          +---------+---------+         +---------+---------+
          |         |         |         |         |         |
       Keeper   ClickHouse  Network   Keeper   ClickHouse  Network
                   |                             |
               PVC / S3                       PVC / S3
```

The architectural story is:

> **The API declares a managed database service. The control plane places and orchestrates it. The Kubernetes operator reconciles database-specific desired state into Kubernetes and ClickHouse state. Kubernetes executes infrastructure primitives. ClickHouse itself remains responsible for database behavior wherever possible.**

That separation is much stronger than building a single HTTP handler that creates StatefulSets.

---

# 2. What You Are Actually Building

The project has four conceptual layers.

## Layer A — Product API

Customer-visible concepts:

- Organization
- Project
- Service
- Region
- Cloud provider
- Compute profile
- Replica count
- Version / release channel
- Backup
- Restore
- Operation
- Credential
- Network access rule

A customer should think:

```text
Create me a production ClickHouse service
in eu-central
with three replicas
and 16 GB memory per replica.
```

The customer should **not** need to think:

```text
Create a StatefulSet,
three PVCs,
a headless Service,
a Keeper StatefulSet,
a PodDisruptionBudget,
some ConfigMaps,
and these topologySpreadConstraints.
```

That translation is the purpose of the platform.

---

## Layer B — Control Plane

Responsibilities:

- authentication
- authorization
- organizations/projects
- API validation
- persistent desired state
- asynchronous operations
- placement into workload clusters
- operation history
- quotas
- version policy
- plan definitions
- region catalog
- audit log
- eventual billing hooks
- multi-cluster orchestration

The control plane should know that a service is:

```text
service_id = svc_123
organization = org_42
region = eu-central
plan = production
replicas = 3
version_channel = stable
state = running
```

It should not contain low-level reconciliation code for individual Pods.

---

## Layer C — Kubernetes Operator

Responsibilities:

- reconcile `ClickHouseService`
- reconcile `KeeperCluster`
- render Kubernetes resources
- control replica lifecycle
- generate ClickHouse configuration
- observe database health
- update resource status
- perform safe database-aware changes
- backups/restores
- upgrade coordination
- failure handling

This is where you learn Kubernetes deeply.

---

## Layer D — Data Plane

Actual customer workload:

- ClickHouse server replicas
- ClickHouse Keeper
- persistent volumes or object storage
- Services / LoadBalancers
- TLS certificates
- monitoring agents
- backup storage

The data plane should remain functional if the public API is temporarily unavailable.

---

# 3. Repository Strategy

## Recommendation: start with one monorepo

Use one repository until components have truly independent release lifecycles.

Why:

- API and operator schemas will evolve together.
- E2E tests need all components.
- one CI pipeline is easier while learning.
- refactoring across boundaries remains easy.
- shared Go types can remain internal.
- Docker images can still be built independently.
- each binary can have its own deployment even though code lives together.

A monorepo does **not** mean a monolith.

You can have:

```text
one git repository
four binaries
three deployments
multiple containers
clear package boundaries
```

## Split into separate repositories only when one of these becomes true

A component:

1. has an independent release cadence,
2. has a separate maintainer/community,
3. requires different permissions/security ownership,
4. should be consumed by unrelated projects,
5. needs compatibility with multiple control-plane versions,
6. is distributed independently.

Likely future split candidates:

- Terraform provider
- generated SDKs
- documentation website
- benchmarking/load-generation project
- BYOC agent
- reusable backup agent

Do **not** start with separate repositories for `api`, `operator`, and `worker`.

That creates versioning and integration overhead before it creates value.

---

# 4. Recommended Repository Structure

Use a structure that resembles mature Go/Kubernetes projects without copying scaffolding blindly.

```text
clickhouse-platform/
├── cmd/
│   ├── api/
│   │   └── main.go
│   ├── operator/
│   │   └── main.go
│   ├── worker/
│   │   └── main.go
│   └── ctl/
│       └── main.go
│
├── api/
│   └── v1/
│       ├── groupversion_info.go
│       ├── clickhouseservice_types.go
│       ├── keepercluster_types.go
│       ├── backup_types.go
│       ├── restore_types.go
│       ├── conditions.go
│       ├── defaults.go
│       └── validation.go
│
├── openapi/
│   └── cloud.yaml
│
├── internal/
│   ├── apiserver/
│   │   ├── http/
│   │   │   ├── handler.go
│   │   │   ├── services.go
│   │   │   ├── backups.go
│   │   │   ├── operations.go
│   │   │   └── errors.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── requestid.go
│   │   │   └── logging.go
│   │   └── server.go
│   │
│   ├── controlplane/
│   │   ├── service/
│   │   │   ├── model.go
│   │   │   ├── service.go
│   │   │   └── validation.go
│   │   ├── operation/
│   │   │   ├── model.go
│   │   │   └── service.go
│   │   ├── placement/
│   │   │   ├── planner.go
│   │   │   └── policy.go
│   │   ├── repository/
│   │   ├── credentials/
│   │   ├── regions/
│   │   └── plans/
│   │
│   ├── operator/
│   │   ├── clickhouseservice/
│   │   │   ├── controller.go
│   │   │   ├── reconcile.go
│   │   │   ├── desired.go
│   │   │   └── status.go
│   │   ├── keeper/
│   │   ├── backup/
│   │   ├── restore/
│   │   └── common/
│   │
│   ├── kubernetes/
│   │   ├── resources/
│   │   │   ├── statefulset.go
│   │   │   ├── service.go
│   │   │   ├── configmap.go
│   │   │   ├── secret.go
│   │   │   ├── pvc.go
│   │   │   └── pdb.go
│   │   ├── ownership/
│   │   └── apply/
│   │
│   ├── clickhouse/
│   │   ├── client.go
│   │   ├── health.go
│   │   ├── replication.go
│   │   ├── topology.go
│   │   ├── configuration.go
│   │   ├── backup.go
│   │   └── version.go
│   │
│   ├── workflow/
│   │   ├── workflow.go
│   │   ├── local/
│   │   └── temporal/
│   │
│   └── platform/
│       ├── versions/
│       ├── plans/
│       └── regions/
│
├── db/
│   ├── migrations/
│   └── queries/
│
├── config/
│   ├── crd/
│   ├── rbac/
│   ├── manager/
│   ├── webhook/
│   └── default/
│
├── deploy/
│   ├── kind/
│   ├── kustomize/
│   └── helm/
│
├── examples/
│   ├── api/
│   └── kubernetes/
│
├── test/
│   ├── integration/
│   ├── e2e/
│   ├── chaos/
│   └── fixtures/
│
├── docs/
│   ├── architecture.md
│   ├── api.md
│   ├── operations.md
│   ├── threat-model.md
│   ├── runbooks/
│   └── adr/
│
├── hack/
├── scripts/
├── Dockerfile.api
├── Dockerfile.operator
├── Dockerfile.worker
├── Makefile
├── go.mod
└── README.md
```

## Why `internal/` matters

Go prevents outside projects from importing packages under `internal/`.

Use that deliberately.

Avoid a giant `pkg/` directory just because many Kubernetes projects have one.

Only create public reusable Go packages when there is an actual external consumer.

---

# 5. API Versioning Decision

The examples in this guide use:

- public REST API: `/v1`
- Kubernetes Go API package: `api/v1`
- Kubernetes API group: `database.<your-domain>/v1`

This keeps the project clean and avoids alpha-style names throughout the codebase.

Important engineering caveat:

A real Kubernetes project normally uses prerelease API versions while schemas are unstable. If you expose your CRDs to third-party users, calling them `v1` creates a compatibility promise.

For this project, treat the Kubernetes CRD as an **internal platform contract** during development. Keep it compatible once the project is publicly released.

The **public REST API** is the primary customer compatibility boundary.

That distinction is important.

---

# 6. Public API Design

Use an OpenAPI-first REST API.

Recommended Go stack:

- `net/http`
- `chi`
- `oapi-codegen`
- `slog`
- `pgx`
- `sqlc`
- `goose`

Avoid a large web framework.

The interesting engineering is the platform, not framework magic.

---

## Core resources

### Organizations

```text
/v1/organizations
/v1/organizations/{organizationId}
```

### Projects

```text
/v1/organizations/{organizationId}/projects
/v1/organizations/{organizationId}/projects/{projectId}
```

Projects are optional for the first version but useful once one organization can own many services.

### Services

```text
POST   /v1/organizations/{organizationId}/services
GET    /v1/organizations/{organizationId}/services
GET    /v1/organizations/{organizationId}/services/{serviceId}
PATCH  /v1/organizations/{organizationId}/services/{serviceId}
DELETE /v1/organizations/{organizationId}/services/{serviceId}
```

### Backups

```text
POST /v1/organizations/{organizationId}/services/{serviceId}/backups
GET  /v1/organizations/{organizationId}/services/{serviceId}/backups
GET  /v1/organizations/{organizationId}/services/{serviceId}/backups/{backupId}
```

### Restore

Prefer a resource/action with a durable operation rather than blocking HTTP.

```text
POST /v1/organizations/{organizationId}/services/{serviceId}/restores
```

### Operations

```text
GET /v1/organizations/{organizationId}/operations/{operationId}
```

Every long-running action should have an observable operation.

---

# 7. Example Customer Service Model

A customer request should be product-oriented.

```json
{
  "name": "analytics-production",
  "provider": "aws",
  "region": "eu-central-1",
  "replicas": 3,
  "compute": {
    "minMemoryGb": 16,
    "maxMemoryGb": 32
  },
  "storage": {
    "sizeGb": 500
  },
  "releaseChannel": "stable"
}
```

Response:

```json
{
  "id": "svc_01K...",
  "name": "analytics-production",
  "state": "provisioning",
  "provider": "aws",
  "region": "eu-central-1",
  "replicas": 3,
  "endpoint": null,
  "operationId": "op_01K..."
}
```

Later:

```json
{
  "id": "svc_01K...",
  "state": "running",
  "endpoint": {
    "hostname": "svc_01K.eu-central.example.com",
    "httpsPort": 8443,
    "nativePort": 9440
  }
}
```

Notice what is missing:

- StatefulSet
- Pod
- PVC
- ConfigMap
- Namespace
- Keeper
- Kubernetes labels

Those are implementation details.

---

# 8. API Best Practices Worth Implementing

## Idempotency

`POST /services` should support an `Idempotency-Key`.

Why:

A client may retry after a timeout without knowing whether creation succeeded.

The API should return the original resource/operation rather than create a second cluster.

This is a very good cloud-platform interview topic.

---

## Optimistic concurrency

For mutable resources, eventually support:

- resource revision
- `ETag`
- `If-Match`

This prevents two clients from accidentally overwriting each other's changes.

You do not need it in the first week.

---

## Long-running operations

Never keep a create request open until the database is running.

Return quickly:

```text
HTTP 202 Accepted
operationId = op_123
```

or create the Service immediately and return:

```text
HTTP 201 Created
state = provisioning
operationId = op_123
```

Then clients poll or stream operation state.

Possible operation states:

```text
queued
running
succeeded
failed
cancelled
```

Do not model every internal controller transition as an API state.

---

## Stable errors

Use structured errors:

```json
{
  "code": "REGION_NOT_AVAILABLE",
  "message": "The selected region does not support this service plan.",
  "requestId": "req_..."
}
```

A client should not need to parse human-readable strings.

---

## Pagination

Use cursor-based pagination for list endpoints.

Not required for the MVP, but design the API so it can be added without breaking clients.

---

# 9. Control Plane Database

Use PostgreSQL.

Suggested initial tables:

```text
organizations
projects
api_keys
services
service_revisions
operations
workload_clusters
placements
backups
audit_events
```

Potential later tables:

```text
quotas
usage_samples
billing_accounts
network_policies
maintenance_windows
version_channels
```

## Do not make PostgreSQL a copy of Kubernetes

Avoid tables like:

```text
pods
statefulsets
persistent_volume_claims
configmaps
```

Kubernetes already owns that state.

The control-plane database stores:

- customer intent
- product metadata
- operation history
- placement
- audit information

Kubernetes stores execution state.

---

# 10. Desired State Flow

The strongest architecture for the managed-provider direction is:

```text
HTTP request
   |
   v
validate
   |
   v
transaction:
  create/update service desired state
  create operation
   |
   v
commit
   |
   v
worker sees operation
   |
   v
placement selects workload cluster
   |
   v
worker creates/patches ClickHouseService CR
   |
   v
operator reconciles
   |
   v
operator updates CR status
   |
   v
worker/control-plane observer maps status
   |
   v
public Service state updated
```

This makes the HTTP request independent from Kubernetes latency.

---

# 11. MVP Shortcut vs Target Architecture

## MVP

```text
API
 |
 v
Kubernetes Client
 |
 v
ClickHouseService CR
```

This is acceptable for the first demonstration.

It lets you get to the operator quickly.

## Better architecture

```text
API
 |
 v
PostgreSQL
 |
 v
operation worker
 |
 v
Kubernetes
 |
 v
operator
```

## Multi-cluster architecture

```text
Public API
   |
Control Plane DB
   |
Placement Engine
   |
Workflow Engine
   |
   +--------------------+
   |                    |
Cluster A            Cluster B
Agent/Client         Agent/Client
   |                    |
Operator             Operator
```

Do not begin with multi-cluster.

Design the interfaces so it is possible later.

---

# 12. Workload Cluster Registry

Eventually represent execution clusters in the control plane.

Example:

```text
cluster_id
provider
region
environment
state
capacity_class
kubernetes_version
operator_version
supported_clickhouse_versions
available_cpu
available_memory
```

A placement interface might look like:

```go
type Planner interface {
    Place(ctx context.Context, service Service, clusters []WorkloadCluster) (Placement, error)
}
```

Start with:

```text
there is exactly one cluster
```

Then evolve to:

```text
provider + region match
```

Then:

```text
capacity + health + maintenance + quotas
```

---

# 13. Kubernetes Resource Model

Recommended high-level internal resources:

```text
ClickHouseService
KeeperCluster
Backup
Restore
```

Start with only `ClickHouseService`.

Add the others when their lifecycle deserves an independent controller.

---

## ClickHouseService

Represents one managed ClickHouse deployment.

Conceptual spec:

```text
image/version
replica count
shard count
resources
storage
keeper reference
networking
scheduling policy
settings overrides
backup policy
```

Conceptual status:

```text
observed generation
conditions
ready replicas
desired replicas
current version
target version
endpoint
topology summary
last successful backup
```

Use Kubernetes `metav1.Condition`.

Useful conditions:

```text
Ready
KeeperReady
ReplicasReady
ReplicationHealthy
BackupReady
UpgradeBlocked
```

Conditions are observations, not your complete workflow state machine.

---

# 14. Kubernetes API Design Rules

Follow Kubernetes API conventions.

## Spec = desired state

Users/controllers write intent to `spec`.

## Status = observed state

The operator writes observations to `status`.

## Track observed generation

If:

```text
metadata.generation = 12
status.observedGeneration = 10
```

the status is stale.

This matters to both humans and automation.

## Use owner references

Resources owned entirely by your CR should have controller owner references where appropriate.

## Use finalizers only for real external cleanup

Good finalizer use:

- remove external DNS record
- preserve or finalize backup
- clean a cloud load balancer
- revoke external credentials

Bad finalizer use:

- arbitrary ordering of normal Kubernetes child deletion

## Defaulting and validation

Use CRD validation for simple constraints.

Use validating/defaulting webhooks when logic cannot be expressed cleanly in the schema.

Examples:

- Keeper replica count should be odd.
- replicas must be compatible with selected plan.
- invalid upgrade paths should be rejected or reported safely.

---

# 15. Reconciler Design

A controller must be level-based and idempotent.

Bad mental model:

```text
on create:
  create cluster

on update:
  update cluster

on delete:
  delete cluster
```

Better mental model:

```text
observe desired state
observe actual state
calculate delta
perform one or more safe changes
publish status
repeat
```

A crash after any step should be recoverable by running `Reconcile` again.

---

# 16. Keep Planning Separate From Execution

A useful pattern:

```text
CurrentState + DesiredState
          |
          v
        Planner
          |
          v
      []Operation
          |
          v
       Executor
```

For example:

```go
type UpgradePlanner interface {
    Plan(current ClusterState, desired ClusterState) ([]Action, error)
}
```

This gives you pure Go logic that is easy to unit test.

Do not make every decision directly inside one 2,000-line reconciler.

---

# 17. Reconcile Pipeline

A production-style reconcile loop can be structured into explicit steps:

```text
Load resource
   |
Handle deletion/finalizer
   |
Apply defaults / validate assumptions
   |
Reconcile Keeper dependency
   |
Reconcile configuration
   |
Reconcile Services
   |
Reconcile storage
   |
Reconcile ClickHouse replicas
   |
Observe ClickHouse health
   |
Update status
```

Each step should answer:

```text
done
requeue
error
```

This style is also visible in the official ClickHouse operator, which uses reconcile-step/pipeline concepts internally.

---

# 18. Stateful Workload Strategy

This is one of the most important architectural topics in the project.

There are three serious options you should understand.

---

## Option A — One StatefulSet for the whole replica set

Example:

```text
StatefulSet clickhouse
  replicas = 3

clickhouse-0
clickhouse-1
clickhouse-2
```

### Advantages

- simplest
- Kubernetes-native
- stable identities
- easy PVC templates
- easy initial learning

### Disadvantages

- ordinal lifecycle is restrictive
- arbitrary replica replacement is difficult
- make-before-break is difficult
- per-replica configuration/version control is difficult

### Use

Build this once manually in the learning phase.

Do not make it your final architecture.

---

## Option B — One StatefulSet per ClickHouse replica

Example:

```text
ch-s0-r0 -> StatefulSet -> one Pod
ch-s0-r1 -> StatefulSet -> one Pod
ch-s0-r2 -> StatefulSet -> one Pod
```

### Advantages

- fine-grained replica lifecycle
- each replica gets stable storage/identity
- Kubernetes StatefulSet controller still keeps Pod alive
- easier arbitrary replacement
- better foundation for make-before-break
- aligns with the official ClickHouse operator design
- aligns with ClickHouse Cloud's MultiSTS lessons

### Disadvantages

- more Kubernetes objects
- operator has more orchestration responsibility
- topology management is more complex

### Recommendation

**Use this as the target architecture for your project.**

---

## Option C — Direct Pod and PVC management

The operator creates Pods/PVCs directly.

### Advantages

- maximum lifecycle control
- arbitrary instance replacement
- no StatefulSet ordinal restrictions
- explicit instance identity

### Disadvantages

- your controller owns more correctness
- more lifecycle logic
- more opportunity to cause downtime
- harder first implementation

### Why you should know it

CloudNativePG intentionally manages Pods/PVCs directly.

TiDB Operator v2 also moved away from StatefulSets for finer-grained control.

This is an important architectural alternative.

### Recommendation

Do not start here.

Consider it later if you introduce an explicit `ClickHouseInstance` resource.

---

# 19. Potential Future Resource Hierarchy

A sophisticated future design can borrow from TiDB Operator v2:

```text
ClickHouseService
      |
      +-- KeeperCluster
      |
      +-- ClickHouseShard
              |
              +-- ClickHouseInstance
              +-- ClickHouseInstance
```

Or:

```text
ClickHouseService
      |
      +-- ClickHouseReplicaSet
              |
              +-- ClickHouseInstance
              +-- ClickHouseInstance
```

This gives every server an independent identity and lifecycle.

Do **not** introduce this hierarchy at the beginning.

It becomes valuable only when:

- per-instance lifecycle is complex,
- MBB becomes important,
- migration workflows need stable identities,
- direct Pod management becomes desirable.

---

# 20. ClickHouse Keeper

Keeper should be treated as a first-class distributed subsystem.

Learn:

- quorum
- leader election
- failure tolerance
- why odd replica counts matter
- persistent state
- network identity
- readiness vs process liveness

Recommended production-style topology:

```text
keeper-0 -> zone-a
keeper-1 -> zone-b
keeper-2 -> zone-c
```

A three-member Keeper cluster tolerates one unavailable member.

The ClickHouse operator should not consider the database cluster healthy simply because all ClickHouse Pods are `Running`.

---

# 21. ClickHouse Topology

Support topology in stages.

## Stage 1

```text
1 shard
1 replica
```

## Stage 2

```text
1 shard
3 replicas
```

## Stage 3

```text
N shards
M replicas
```

Example:

```text
shard 0
  replica 0
  replica 1

shard 1
  replica 0
  replica 1
```

The operator should understand:

- shard identity
- replica identity
- ClickHouse macros
- cluster configuration
- Keeper paths
- replica membership
- service discovery
- replicated database/table behavior

---

# 22. Scheduling and Failure Domains

Do not equate "three replicas" with "high availability."

Three replicas on one node are not meaningful HA.

Use:

- `topologySpreadConstraints`
- pod anti-affinity where appropriate
- node selectors
- tolerations
- availability-zone labels

Desired placement:

```text
replica 0 -> zone a
replica 1 -> zone b
replica 2 -> zone c
```

Then test:

```text
lose one Pod
lose one worker node
drain one worker node
lose one zone in a simulated environment
```

---

# 23. Database-Aware Health

Kubernetes health:

```text
Pod Running
container Ready
```

Database health:

```text
Keeper connected
replica active
replication queue healthy
read-only mode disabled
replication lag acceptable
tables initialized
database reachable
```

The operator should query ClickHouse itself.

Useful system tables include, depending on your implementation:

- `system.replicas`
- `system.replication_queue`
- `system.clusters`
- `system.tables`
- `system.processes`
- backup-related system state

Do not run expensive health queries on every reconcile.

Design bounded checks and reasonable requeue periods.

---

# 24. Networking

Initial local environment:

```text
ClusterIP
port-forward
```

Then:

```text
LoadBalancer or Ingress/Gateway
TLS
DNS
```

Managed-cloud abstraction:

```text
service endpoint
```

not:

```text
Kubernetes Service
```

Eventually support:

- public endpoint
- IP allow lists
- private endpoint
- VPC peering / private networking
- TLS-only access

Do not implement private networking early.

---

# 25. Credentials

MVP:

- generate ClickHouse username/password
- store the actual secret in Kubernetes Secret
- public API returns the password once
- control-plane database stores only metadata/reference where possible

API keys for the cloud control plane:

- generate high-entropy secret
- show once
- store a strong hash, not plaintext
- store a visible key ID/prefix for lookup
- support revocation
- support scoped roles

Later:

- OIDC
- service accounts
- short-lived credentials
- Vault / cloud KMS integration

---

# 26. Backup and Restore Model

Backups deserve independent resources because they have a lifecycle.

```text
Backup
  Pending
  Running
  Succeeded
  Failed
```

Restore is also long-running.

Do not implement backup by copying files yourself first.

Use ClickHouse's supported backup mechanisms and object storage.

The operator's responsibility is orchestration:

```text
validate destination
choose safe execution target
start backup
observe progress
publish status
handle retry/failure
```

---

# 27. Restore Verification

A particularly strong portfolio feature:

```text
backup production-like service
        |
        v
create temporary validation service
        |
        v
restore backup
        |
        v
run validation queries/checksums
        |
        v
mark backup verified
        |
        v
destroy temporary service
```

This demonstrates that a backup is not considered useful merely because a file exists.

---

# 28. Upgrades

Upgrade logic is where your ClickHouse knowledge becomes highly visible.

A safe upgrade controller needs to reason about:

- compatible source/target versions
- Keeper health
- replica count
- database health
- active backups
- long-running queries
- replication lag
- topology
- rollback/abort behavior

Basic rolling strategy:

```text
select replica
remove/drain from traffic
wait for safe point
replace replica
wait Kubernetes ready
wait ClickHouse ready
wait replication healthy
return to traffic
continue
```

A failed upgrade should block further destruction.

---

# 29. Make-Before-Break

This is an advanced milestone, not MVP.

Concept:

```text
old replicas: A B C

create:
A B C D E F

wait until D E F are healthy
shift traffic

remove:
A B C

result:
D E F
```

Benefits:

- capacity does not disappear during scaling/upgrades
- safer under heavy load
- new replicas can prove themselves before old ones disappear

This is one of the strongest possible portfolio features because it directly relates Kubernetes lifecycle constraints to distributed database behavior.

---

# 30. Separate Normal Reconciliation From Complex Migrations

Do not put every special migration into the main controller.

ClickHouse Cloud's published architecture is a useful lesson here: complex live migrations were separated into a dedicated migration controller rather than making the normal reconcile loop responsible for every migration step.

A good principle:

```text
normal level-based reconciliation
        !=
long, ordered, reversible maintenance workflow
```

Examples that may deserve a separate workflow/controller:

- storage migration
- SingleSTS -> MultiSTS migration
- major topology transformation
- cross-cluster move
- disaster-recovery promotion

---

# 31. Workflow Engine Strategy

## Phase 1

No workflow engine.

Use:

```text
PostgreSQL operation table
+
worker loop
+
idempotent handlers
```

Learn the fundamentals yourself.

## Phase 2

Introduce durable workflow semantics:

- retries
- timeouts
- compensation
- resume after crash
- operation history

## Phase 3

Evaluate Temporal.

Temporal is highly relevant because ClickHouse Cloud has publicly described using it for complex maintenance/migration orchestration.

Do not introduce Temporal before you have a workflow complex enough to justify it.

Otherwise you learn Temporal instead of learning your own failure semantics.

---

# 32. Kubernetes vs Workflow Responsibility

Use Kubernetes reconciliation for:

```text
"this resource should continuously exist in this form"
```

Use workflow orchestration for:

```text
"perform these ordered steps with retries, waiting, and compensation"
```

Example:

Kubernetes:

```text
ClickHouseService should have three healthy replicas.
```

Workflow:

```text
Put service into maintenance mode.
Create replacement replicas.
Wait for replication.
Move traffic.
Delete old replicas.
Exit maintenance mode.
```

This distinction is fundamental.

---

# 33. Observability

Every layer should be observable.

## API metrics

```text
http_requests_total
http_request_duration_seconds
api_errors_total
```

## Control plane

```text
operations_total
operations_duration_seconds
operations_failed_total
placement_failures_total
```

## Operator

```text
reconcile_total
reconcile_duration_seconds
reconcile_errors_total
reconcile_requeues_total
```

## Database platform

```text
clickhouse_service_ready
clickhouse_ready_replicas
clickhouse_replication_lag_seconds
keeper_members_ready
backup_last_success_timestamp
upgrade_in_progress
```

## Logs

Use structured logs with:

```text
request_id
organization_id
service_id
operation_id
namespace
resource_name
reconcile_id
```

Do not log secrets.

---

# 34. Kubernetes Events

Emit human-readable events for significant transitions:

```text
ServiceProvisioningStarted
KeeperReady
ReplicaCreated
ReplicaReady
ClusterReady
UpgradeStarted
ReplicaUpgradeStarted
UpgradeCompleted
BackupStarted
BackupCompleted
RestoreFailed
```

Avoid emitting an event every reconcile.

---

# 35. Testing Strategy

A top-tier operator project is defined as much by failure testing as by its happy path.

## Unit tests

Pure logic:

- topology planner
- placement planner
- upgrade planner
- version compatibility
- API validation
- resource naming
- configuration generation

## Controller tests

Use `envtest`.

Test:

- CR creation
- generated Kubernetes objects
- status updates
- ownership
- deletion/finalizer paths
- invalid updates

## API integration tests

Use real PostgreSQL via Testcontainers.

Test:

- idempotency
- transaction behavior
- operation creation
- authorization
- persistence

## End-to-end tests

Use Kind.

Test full flow:

```text
HTTP API
 -> control plane
 -> CR
 -> operator
 -> ClickHouse
 -> SQL query
```

## Failure tests

```text
delete ClickHouse pod
delete Keeper pod
drain worker
kill operator
kill API
restart worker
interrupt upgrade
interrupt backup
break Keeper connectivity
```

## Compatibility tests

Eventually build a small matrix:

```text
Kubernetes version
ClickHouse version
operator version
```

The official ClickHouse operator itself uses unit/functional tests, Kind E2E tests, fuzzing, and version compatibility testing.

---

# 36. Security Topics You Should Know

Do not make enterprise security an early scope requirement, but understand:

- RBAC
- namespace isolation
- ServiceAccounts
- least privilege
- Kubernetes Secrets
- secret encryption at rest
- TLS
- certificate rotation
- network policies
- Pod security contexts
- image pinning/digests
- supply-chain scanning
- audit logs
- API rate limiting
- tenant isolation

For a public provider, tenant boundaries eventually become one of the most important design areas.

---

# 37. Multi-Tenancy Model

Start with:

```text
one Kubernetes namespace per managed service
```

It is easy to understand and inspect.

Possible later models:

```text
namespace per organization
namespace per service
dedicated cluster per enterprise tenant
shared workload cluster with strong isolation
BYOC cluster
```

Do not claim hard multi-tenant security until you have explicitly designed and tested it.

---

# 38. Managed Provider Deployment Models

## Model A — Local developer cloud

```text
Kind
single machine
local storage
```

Purpose:

- development
- learning
- CI

## Model B — Single managed workload cluster

```text
API/control plane
        |
        v
one Kubernetes cluster
        |
many ClickHouse services
```

This is the right first real provider architecture.

## Model C — Multi-region managed service

```text
Global Control Plane
       |
       +-- eu-central cluster
       +-- eu-west cluster
       +-- us-east cluster
```

## Model D — BYOC

```text
Your SaaS control plane
        |
 outbound secure control channel
        |
customer cloud account
        |
Kubernetes + operator
        |
ClickHouse data
```

BYOC is a later-stage architecture.

---

# 39. Recommended Technology Choices

| Problem | Recommended first choice | Alternatives to understand |
|---|---|---|
| Operator framework | Kubebuilder + controller-runtime | Operator SDK, raw client-go |
| Public HTTP API | chi + net/http | Gin, Echo, Fiber |
| API contract | OpenAPI + oapi-codegen | handwritten DTOs, Connect/gRPC |
| Control-plane DB | PostgreSQL | CockroachDB, etcd |
| DB access | pgx + sqlc | GORM, Ent |
| Migrations | goose | Atlas, migrate |
| Local Kubernetes | Kind | k3d, minikube |
| Deployment manifests | Kustomize first | Helm |
| Operator packaging | Helm later | raw manifests / OLM |
| Metrics | Prometheus | OpenTelemetry metrics |
| Logs | slog | zap, zerolog |
| Tracing | OpenTelemetry later | vendor SDK |
| Async work | DB-backed worker first | Temporal |
| Object storage | S3-compatible API | cloud-specific clients |
| Certificates | cert-manager | manual PKI |
| E2E | Kind | managed test cluster |
| Controller test | envtest | fake client only |

---

# 40. Research: What to Learn From Top-Tier Projects

The goal is not to imitate one repository.

Take one architectural lesson from each.

---

## 40.1 Official ClickHouse Kubernetes Operator

### Study

- repository structure
- Kubebuilder/controller-runtime
- `ClickHouseCluster`
- `KeeperCluster`
- reconcile step pipeline
- templates/resource rendering
- validation/defaulting webhooks
- status conditions
- E2E test layout
- compatibility testing
- one StatefulSet per replica

### Most important lesson

The official operator deliberately uses a StatefulSet per ClickHouse replica for fine-grained lifecycle control.

It also follows the principle:

> if database behavior can live safely in ClickHouse itself, prefer that over rebuilding it in Kubernetes orchestration.

### What not to copy blindly

Its current CRD API is still evolving. Your customer-facing cloud API should be a separate abstraction.

Reference:
- https://github.com/ClickHouse/clickhouse-operator
- https://clickhouse.com/blog/clickhouse-kubernetes-operator

---

## 40.2 ClickHouse Cloud Make-Before-Break Architecture

### Study

- SingleSTS limitations
- MultiSTS
- make-before-break
- graceful replica removal
- active query/backup awareness
- live migrations
- maintenance mode
- separate migration controller
- Temporal workflows
- synchronization between controllers
- topology problems during transitions

### Most important lesson

Stateful database operations often require two different mechanisms:

1. a steady-state reconciler
2. durable ordered maintenance workflows

Trying to make one controller do both can create dangerous complexity.

Reference:
- https://clickhouse.com/blog/make-before-break-faster-scaling-mechanics-for-clickhouse-cloud

---

## 40.3 Altinity ClickHouse Operator

### Study

Altinity is valuable because it represents years of production ClickHouse-on-Kubernetes experience.

Feature areas:

- storage templates
- Pod templates
- Service templates
- settings/configuration
- users
- shards/replicas
- upgrades
- schema propagation
- monitoring
- Keeper/ZooKeeper integration

### Most important lesson

Use Altinity as a **feature checklist**.

Do not use its large, highly flexible CRD as the model for your first API.

A learning project benefits from a smaller opinionated API.

Reference:
- https://github.com/Altinity/clickhouse-operator

---

## 40.4 CloudNativePG

CloudNativePG is one of the best projects to study for database-aware Kubernetes design.

### Study

- application-aware status
- direct Kubernetes API integration
- custom Pod controller
- direct PVC ownership
- failover
- rolling upgrades
- backup resources
- scheduled backup
- multi-zone recommendations
- immutable image strategy
- explicit database lifecycle management

### Most important lesson

A database operator does not have to use StatefulSets.

CloudNativePG intentionally manages Pods and PVCs directly because database instance identity/lifecycle can require finer control.

This is the strongest alternative to your MultiSTS design.

Reference:
- https://github.com/cloudnative-pg/cloudnative-pg
- https://cloudnative-pg.io/

---

## 40.5 Crunchy Postgres Operator (PGO)

### Study

- one high-level `PostgresCluster` resource
- integrated HA
- backup and disaster recovery
- monitoring
- topology configuration
- operator as collection of controllers
- database-as-a-service orientation

### Most important lesson

A good CRD can expose a simple user contract while generating many low-level Kubernetes resources.

Also notice the principle of delegating HA responsibilities to database-aware components rather than making the central operator the only source of failover logic.

Reference:
- https://github.com/CrunchyData/postgres-operator
- https://access.crunchydata.com/documentation/postgres-operator/latest/architecture

---

## 40.6 Strimzi Kafka Operator

### Study

Strimzi separates responsibilities:

```text
Cluster Operator
Topic Operator
User Operator
```

### Most important lesson

Not every domain belongs in the cluster reconciler.

If ClickHouse users, databases, backups, ingestion pipelines, or access policies become large domains, separate controllers/resources can be cleaner than one giant CRD/reconciler.

Reference:
- https://github.com/strimzi/strimzi-kafka-operator
- https://strimzi.io/docs/operators/latest/

---

## 40.7 TiDB Operator v2

TiDB Operator v2 is especially interesting for your design.

It moved from a large cluster resource toward:

```text
Cluster
  |
ComponentGroup
  |
Instance
```

It also moved away from StatefulSet dependence and toward direct Pod management.

### Most important lesson

When per-instance lifecycle becomes sufficiently complicated, introducing an explicit instance resource can make orchestration easier to reason about.

This is a credible future evolution for your ClickHouse project.

Do not start there.

Reference:
- https://github.com/pingcap/tidb-operator
- https://docs.pingcap.com/tidb-in-kubernetes/dev/architecture/

---

## 40.8 Crossplane

Crossplane is not a database operator, but it is extremely relevant to the managed-provider part.

Core idea:

```text
high-level platform API
        |
        v
managed resource
        |
        v
provider/external system
```

### Most important lesson

Separate the **consumer API** from the **provider-specific execution representation**.

Your equivalent:

```text
Public ClickHouse Service API
        |
        v
control-plane desired state
        |
        v
Kubernetes ClickHouseService
        |
        v
ClickHouse Operator
```

This makes Kubernetes an implementation target rather than your public product API.

Reference:
- https://github.com/crossplane/crossplane
- https://docs.crossplane.io/

---

## 40.9 Vitess Operator

Vitess is useful for studying:

- compatibility matrices
- distributed database topology
- lifecycle automation
- version support policies

The key portfolio lesson is that version compatibility should eventually become an explicit tested contract, not "whatever image tag the user typed."

Reference:
- https://github.com/planetscale/vitess-operator

---

# 41. Architecture Decision Summary

Use these decisions unless the project provides evidence to change them.

| Decision | Choice |
|---|---|
| Repository | Monorepo |
| Language | Go |
| Operator base | controller-runtime / Kubebuilder |
| Initial Kubernetes environment | Kind |
| Final replica primitive | StatefulSet per replica |
| Alternative to study | direct Pod/PVC controller |
| Public API | REST + OpenAPI |
| Public API version | `/v1` |
| Control-plane persistence | PostgreSQL |
| HTTP -> K8s coupling | temporary only |
| Async provisioning | operation worker |
| Complex workflows | Temporal later |
| Keeper | first-class managed component |
| Backup | first-class resource |
| Restore | first-class resource |
| Customer abstraction | Service, not Kubernetes resources |
| Multi-cluster | later |
| BYOC | much later |

---

# 42. Roadmap

Each phase has four questions:

1. What should I build?
2. What should I learn?
3. What alternatives should I understand?
4. What is the exit criterion?

Do not advance because the code "mostly works."

Advance when you can explain why it works.

---

# Phase 0 — Kubernetes Fundamentals Without an Operator

## Build

Create a local Kind cluster.

Deploy manually:

- ClickHouse container
- ConfigMap
- Secret
- Service
- headless Service
- PVC
- StatefulSet
- readiness/liveness probes

Insert data.

Delete the Pod.

Verify persistence.

Inspect DNS.

Inspect mounted volumes.

## Learn

Be able to explain:

- Pod vs Deployment vs StatefulSet
- PV vs PVC
- StorageClass
- Service vs headless Service
- labels/selectors
- owner references
- probes
- scheduler decisions
- namespaces
- RBAC basics

## Alternatives

Know:

- Deployment
- direct Pod
- StatefulSet
- local-path storage
- network block storage

## Exit criterion

You can rebuild the ClickHouse deployment from memory and explain why every Kubernetes object exists.

---

# Phase 1 — Minimal Go Operator

## Build

Scaffold controller-runtime/Kubebuilder.

Create one high-level resource:

```text
ClickHouseService
```

It should reconcile:

- one StatefulSet
- one PVC
- one Service
- one ConfigMap

Implement:

- spec
- status
- Ready condition
- observed generation
- ownership
- deletion

## Learn

- Kubernetes API machinery
- scheme
- GVK
- informer/cache
- watch
- work queue
- reconcile request
- cached reads
- optimistic concurrency
- controller manager
- leader election basics

## Alternatives

Understand:

- Operator SDK
- raw client-go
- Helm operator
- GitOps without custom controller

## Exit criterion

Deleting a managed child resource causes the operator to recreate it.

Restarting the operator does not damage the service.

---

# Phase 2 — Production-Style Reconciliation Structure

## Build

Refactor into:

```text
controller
planner
resource builders
ClickHouse client
status writer
```

Add:

- deterministic naming
- labels
- spec hashes/revisions
- validation
- defaulting
- Kubernetes Events
- metrics

## Learn

- idempotency
- level-based control
- eventual consistency
- stale cache behavior
- resourceVersion conflicts
- status subresource
- finalizers
- conditions

## Alternatives

Study:

- monolithic reconcile function
- pipeline/step reconciler
- separate subcontrollers

## Exit criterion

You can deliberately crash the controller between reconcile steps and it recovers safely.

---

# Phase 3 — Keeper

## Build

Add a `KeeperCluster` resource or an internally managed Keeper component.

Provision three Keeper members.

Spread them across Kind worker topology labels.

Expose Keeper health in service status.

## Learn

- quorum
- consensus basics
- leader/follower
- majority
- split brain prevention
- persistent coordination state
- failure domains

## Alternatives

Understand:

- ZooKeeper
- external Keeper
- operator-managed Keeper
- shared Keeper vs dedicated Keeper

## Exit criterion

Kill one Keeper member and ClickHouse remains functional.

Explain what happens if two of three are unavailable.

---

# Phase 4 — Replicated ClickHouse

## Build

Move to:

```text
one shard
three replicas
```

Introduce StatefulSet-per-replica.

Generate ClickHouse topology/configuration.

Make replication health part of Ready status.

## Learn

- ReplicatedMergeTree
- Keeper paths
- replica identity
- ClickHouse macros
- distributed DDL / replicated database behavior
- service discovery

## Alternatives

Understand:

- single StatefulSet
- MultiSTS
- direct Pod ownership

## Exit criterion

Delete any ClickHouse replica.

It returns, catches up, and the operator reports healthy state again.

---

# Phase 5 — Shards and Topology

## Build

Support:

```text
N shards x M replicas
```

Add topology-aware placement.

Validate unsafe layouts.

## Learn

- ClickHouse sharding
- replicas vs shards
- Distributed engine
- local vs distributed tables
- network topology
- failure domains

## Alternatives

Understand:

- one large shard
- many shards
- SharedMergeTree architecture conceptually
- compute/storage separation

## Exit criterion

Deploy a two-shard/two-replica cluster and demonstrate distribution + replica recovery.

---

# Phase 6 — Public Go API

## Build

Add `cmd/api`.

Define OpenAPI.

Implement:

```text
POST /v1/.../services
GET  /v1/.../services
GET  /v1/.../services/{id}
PATCH /v1/.../services/{id}
DELETE /v1/.../services/{id}
```

Initially it may directly create/patch the internal Kubernetes resource.

## Learn

- HTTP semantics
- API contracts
- DTO vs domain model
- OpenAPI
- generated clients
- authentication middleware
- idempotency
- structured errors
- request IDs

## Alternatives

Understand:

- REST
- gRPC
- Connect RPC
- GraphQL
- exposing Kubernetes API directly

## Exit criterion

A user can create a working ClickHouse service without running `kubectl`.

---

# Phase 7 — Real Control Plane

## Build

Add PostgreSQL.

Persist:

- organizations
- API keys
- services
- operations

Change the flow:

```text
API -> PostgreSQL -> worker -> Kubernetes
```

Return operation IDs.

## Learn

- transactions
- async APIs
- durable operations
- retry semantics
- idempotency across process crashes
- desired vs observed state
- control-plane persistence

## Alternatives

Understand:

- Kubernetes as the only database
- event sourcing
- message broker
- DB-backed queue
- Temporal

## Exit criterion

Shut down the worker after an operation is created.

Restart it.

The operation resumes safely and creates exactly one service.

---

# Phase 8 — Authentication, Organizations, Credentials

## Build

Add:

- organizations
- API keys
- scoped authorization
- ClickHouse generated credentials
- audit events

## Learn

- API-key storage
- hashing
- authorization boundaries
- secret handling
- auditability

## Alternatives

Understand:

- OIDC
- JWT
- mTLS
- service accounts

## Exit criterion

Two organizations cannot view or mutate each other's services.

---

# Phase 9 — Managed Networking

## Build

Expose a stable ClickHouse endpoint.

Add:

- TLS
- DNS
- public endpoint abstraction
- optional IP allow list

## Learn

- Kubernetes Services
- LoadBalancer
- ingress/gateway concepts
- cert-manager
- DNS automation
- network policy

## Alternatives

Understand:

- direct LoadBalancer per service
- shared proxy
- Gateway API
- private endpoints

## Exit criterion

Create service through API and connect using only endpoint + generated credentials.

---

# Phase 10 — Backups and Restores

## Build

Add API + Kubernetes resources for:

- Backup
- Restore

Backups go to S3-compatible storage.

Expose progress/status.

## Learn

- object storage
- retention
- backup consistency
- restore semantics
- long-running operation design
- data-loss boundaries

## Alternatives

Understand:

- native ClickHouse backup
- clickhouse-backup ecosystem
- filesystem snapshots
- volume snapshots

## Exit criterion

Create data, backup, destroy service, recreate, restore, verify data.

---

# Phase 11 — Safe Upgrades

## Build

Implement database-aware rolling upgrades.

Add compatibility rules.

Block upgrade while unsafe operations are active.

Expose progress.

## Learn

- disruption budgets
- graceful shutdown
- rolling replacement
- version compatibility
- database readiness
- maintenance operations

## Alternatives

Understand:

- Kubernetes RollingUpdate
- OnDelete
- direct pod replacement
- blue/green
- make-before-break

## Exit criterion

Run continuous queries/inserts while upgrading a replicated cluster with bounded disruption.

---

# Phase 12 — Failure and Chaos Engineering

## Build

Create automated scenarios:

```text
kill replica
kill Keeper
kill operator
kill worker
drain node
interrupt upgrade
interrupt backup
break DNS/network
```

## Learn

- distributed failure behavior
- retry storms
- backoff
- reconciliation convergence
- failure domains
- recovery objectives

## Alternatives

Understand:

- Chaos Mesh
- LitmusChaos
- custom test scripts

## Exit criterion

Produce a documented failure matrix with expected behavior and measured recovery.

---

# Phase 13 — Multi-Cluster Control Plane

## Build

Add workload cluster registry.

Implement placement.

Initially:

```text
provider + region
```

Later:

```text
provider + region + capacity + health
```

## Learn

- global vs regional control planes
- fleet management
- cluster credentials
- version skew
- rollout strategy
- regional failure

## Alternatives

Understand:

- one giant Kubernetes cluster
- federation
- Crossplane-style providers
- agent-based execution
- direct Kubernetes API access

## Exit criterion

The same public API creates services in two independent Kubernetes clusters.

---

# Phase 14 — Make-Before-Break

## Build

Implement a safe replica replacement workflow.

Possible use cases:

- vertical resize
- version upgrade
- node migration

Flow:

```text
make replacement
wait healthy
sync
shift traffic
break old replica
```

## Learn

- StatefulSet limitations
- instance identity
- capacity preservation
- ordered maintenance
- reversible migration
- controller synchronization

## Alternatives

Understand:

- rolling restart
- surge semantics
- direct Pod controller
- instance CRDs

## Exit criterion

Demonstrate a resize/upgrade where replacement capacity becomes healthy before old capacity disappears.

---

# Phase 15 — Durable Workflows / Temporal

## Build

Only now consider moving complex maintenance operations into Temporal.

Candidate workflow:

```text
maintenance mode
capture state
create replacements
wait for K8s events
verify ClickHouse
switch traffic
remove old replicas
restore service settings
```

## Learn

- durable execution
- activities
- workflow determinism
- retries
- timeouts
- compensation
- workflow versioning

## Alternatives

Understand:

- DB state machines
- message broker + consumers
- Kubernetes Jobs
- dedicated migration controller

## Exit criterion

Kill workflow workers during a live migration and successfully resume the operation.

---

# Phase 16 — BYOC / Agent Model

Optional advanced phase.

## Build

Instead of opening customer Kubernetes API to your SaaS:

```text
customer cluster agent
        |
 outbound authenticated connection
        |
your control plane
```

Agent receives desired operations and reports status.

## Learn

- trust boundaries
- outbound-only control channels
- identity
- credential rotation
- multi-tenant SaaS control plane
- customer-owned infrastructure

## Exit criterion

Provision a service into a second Kubernetes cluster without storing a broad administrator kubeconfig in the public API service.

---

# 43. First Four Milestones I Would Actually Target

Do not think about sixteen phases every day.

Use these portfolio milestones.

## Milestone A — Operator Core

```text
Kind
ClickHouseService
1 ClickHouse
persistent storage
reconciliation
status
tests
```

## Milestone B — Distributed Database

```text
Keeper
3 replicas
MultiSTS
topology spread
database-aware health
failure recovery
```

## Milestone C — Managed Provider

```text
REST API
PostgreSQL control plane
operations worker
auth
endpoint + credentials
```

## Milestone D — Production Mechanics

```text
backup/restore
safe upgrade
chaos tests
make-before-break
multi-cluster optional
```

Milestone C is already enough for a very strong portfolio project.

Milestone D makes it exceptional.

---

# 44. What Not to Build Early

Avoid:

- custom Kubernetes scheduler
- custom CSI driver
- your own consensus implementation
- your own object storage
- a full billing platform
- frontend console
- Terraform provider
- multi-cloud
- autoscaling recommender
- BYOC
- service mesh
- custom SQL proxy
- custom backup file format

These can distract from the core learning goals.

---

# 45. Frontend?

Do not build a React console initially.

For a portfolio infrastructure project, this is stronger:

```text
curl
CLI
OpenAPI
Grafana
kubectl
architecture docs
```

than spending two weeks on dashboards and forms.

A UI can come after the managed-service API is stable.

---

# 46. CLI

A small CLI is valuable later.

Example:

```bash
chp service create \
  --name analytics \
  --provider local \
  --region dev \
  --replicas 3

chp service get svc_123

chp service scale svc_123 --replicas 4

chp backup create svc_123

chp operation get op_456
```

The CLI should use the public API, never Kubernetes directly.

This proves that the API is a real product boundary.

---

# 47. Terraform Provider

Do not build it initially.

It becomes useful once `/v1` is stable.

Then it should live in its own repository because:

- independent releases
- Terraform SDK dependencies
- separate compatibility contract
- external consumers

A Terraform provider is a strong final proof that the public API is automation-friendly.

---

# 48. Portfolio Demo

Create a scripted demo.

```text
1. Start control plane.
2. Start Kind workload cluster.
3. POST /services.
4. Watch operation progress.
5. Connect to returned endpoint.
6. Insert several million rows.
7. Kill a ClickHouse replica.
8. Show recovery.
9. Kill a Keeper member.
10. Show service remains available.
11. Trigger upgrade.
12. Show health gating.
13. Create backup.
14. Destroy service.
15. Restore it.
16. Validate row count/checksum.
```

Later:

```text
17. Trigger make-before-break resize.
18. Show old + new replicas overlap.
19. Show traffic remains available.
20. Remove old replicas.
```

Record this as a short terminal demo.

---

# 49. README Story

Your README should answer in the first screen:

## What is this?

An experimental managed ClickHouse platform built in Go on Kubernetes.

## Why does it exist?

To explore how database control planes handle lifecycle operations Kubernetes cannot safely infer on its own.

## What does it demonstrate?

- Kubernetes operators
- distributed database lifecycle
- HA
- upgrades
- backup/restore
- control-plane APIs
- asynchronous operations
- failure recovery

## Architecture

One diagram.

## Demo

One command sequence.

## Engineering decisions

Links to ADRs.

---

# 50. Architecture Decision Records

Write ADRs for decisions that have meaningful alternatives.

Suggested:

```text
ADR-001 monorepo
ADR-002 controller-runtime
ADR-003 public API separate from CRD
ADR-004 PostgreSQL control-plane state
ADR-005 StatefulSet per replica
ADR-006 Keeper ownership model
ADR-007 backup orchestration
ADR-008 async operation model
ADR-009 workflow engine threshold
ADR-010 multi-cluster placement
```

Each ADR:

```text
Context
Decision
Alternatives
Consequences
```

This is extremely valuable for interviews.

---

# 51. Interview Questions You Should Be Able to Answer

By the end, you should answer:

### Kubernetes

- Why an operator instead of Helm?
- What happens if the operator is down?
- Why use a cache?
- How do you prevent duplicate side effects?
- What does `observedGeneration` tell you?
- Why use a finalizer?
- What happens during node drain?
- Why is a StatefulSet useful?
- Why might StatefulSet become restrictive?

### ClickHouse

- What is a shard?
- What is a replica?
- What does Keeper coordinate?
- What does replication health mean?
- How do you safely add/remove a replica?
- What changes during an upgrade?
- How do you verify a backup?

### Control planes

- Why not let the REST API call Kubernetes synchronously?
- What is the source of truth?
- How do retries avoid creating two services?
- How is placement decided?
- What happens if the worker crashes halfway through provisioning?
- How would you add another region?
- How would you implement BYOC?

### Distributed systems

- What if desired state changes while reconciliation is in progress?
- What if two controllers act on the same resource?
- What if the API write succeeds but the response is lost?
- What if Kubernetes accepts a resource but the operator never converges?
- What does eventual consistency mean here?

---

# 52. Definition of Done for a Strong Portfolio Version

You do not need every advanced phase.

A strong version should include:

```text
[ ] Monorepo with clear binary/package boundaries
[ ] Go public REST API
[ ] OpenAPI spec
[ ] PostgreSQL control-plane database
[ ] asynchronous operation model
[ ] Kubernetes operator using controller-runtime
[ ] ClickHouseService custom resource
[ ] Keeper management
[ ] one StatefulSet per ClickHouse replica
[ ] persistent storage
[ ] multiple replicas
[ ] shards
[ ] topology-aware scheduling
[ ] database-aware health
[ ] status conditions
[ ] TLS endpoint
[ ] backup
[ ] restore
[ ] safe rolling upgrade
[ ] Prometheus metrics
[ ] structured logging
[ ] unit tests
[ ] envtest controller tests
[ ] Kind E2E tests
[ ] failure tests
[ ] CI
[ ] architecture documentation
[ ] ADRs
[ ] scripted demo
```

Exceptional extras:

```text
[ ] make-before-break replacement
[ ] automatic restore verification
[ ] multi-cluster placement
[ ] Temporal maintenance workflows
[ ] BYOC agent
[ ] Terraform provider
[ ] upstream contribution
```

---

# 53. Recommended Starting Sequence

Your first concrete implementation sequence should be:

```text
01. Create monorepo.
02. Create Kind cluster with three workers.
03. Manually deploy one ClickHouse StatefulSet.
04. Add persistent storage.
05. Verify persistence after Pod deletion.
06. Create operator binary.
07. Define ClickHouseService resource.
08. Reconcile Service + ConfigMap + PVC + StatefulSet.
09. Add status/conditions.
10. Add envtest.
11. Add three-member Keeper.
12. Add replicated ClickHouse.
13. Move to StatefulSet-per-replica.
14. Add database-aware health.
15. Add first Kind E2E test.
16. Only then start the public API.
```

The key rule:

> **Learn the Kubernetes primitive manually before hiding it behind your controller.**

---

# 54. Public API Starting Sequence

After the operator is healthy:

```text
01. Write OpenAPI service schema.
02. Generate Go server interfaces/types.
03. Implement POST /services.
04. Initially create CR directly.
05. Implement GET /services/{id}.
06. Map Kubernetes status to API response.
07. Add PostgreSQL.
08. Persist service desired state.
09. Add operations table.
10. Move Kubernetes mutation into worker.
11. Add idempotency key.
12. Add API-key authentication.
13. Add organization isolation.
```

This sequence intentionally lets the architecture evolve instead of front-loading infrastructure.

---

# 55. When to Introduce a Message Broker

Probably later than you think.

A PostgreSQL operations table is sufficient when:

- one worker class
- moderate operation volume
- simple retries
- no streaming fan-out requirement

Introduce a queue when you can state the problem it solves.

Possible later choices:

- NATS JetStream
- Kafka
- cloud queues
- Temporal task queues

Do not add Kafka merely because this is a distributed systems project.

---

# 56. When to Introduce Temporal

Use Temporal when you have workflows like:

```text
step 1
wait 15 minutes
step 2
retry external call
wait for event
step 3
on failure compensate
resume safely after restart
```

A ClickHouse MBB migration is a legitimate Temporal use case.

Simple provisioning is not.

---

# 57. A Possible Future Production Architecture

```text
                            Global API
                               |
                  +------------+------------+
                  |                         |
               AuthN/Z                  Audit Log
                  |
            Service Catalog
                  |
             PostgreSQL
                  |
           Operation Service
                  |
             Placement
                  |
              Temporal
                  |
       +----------+----------+
       |                     |
       v                     v
 Europe Control         US Control
       |                     |
       v                     v
 Cluster Agent          Cluster Agent
       |                     |
       v                     v
 ClickHouse Operator    ClickHouse Operator
       |                     |
       v                     v
 CH / Keeper / S3       CH / Keeper / S3
```

Do not build this first.

Use it as the architectural direction.

---

# 58. Best-Practice Principle: Separate Product State From Infrastructure State

Example:

## Product state

```text
Service:
  replicas = 3
  memory = 16GB
  release_channel = stable
```

## Kubernetes state

```text
3 StatefulSets
3 PVCs
Services
ConfigMaps
Secrets
PDB
Keeper
```

## ClickHouse state

```text
replication healthy
catalog synchronized
Keeper connected
```

Your system continuously maps:

```text
Product intent
     ->
Kubernetes desired state
     ->
Database actual state
     ->
Product observed state
```

That is the core control-plane design.

---

# 59. Best-Practice Principle: The Operator Should Not Be the Whole Product

The public API should eventually work with:

```text
Kubernetes implementation A
Kubernetes implementation B
BYOC
future non-Kubernetes execution
```

Even if you never build those alternatives, keeping the boundary produces a better architecture.

---

# 60. Best-Practice Principle: Prefer Database-Native Mechanisms

Do not teach Kubernetes to implement ClickHouse features ClickHouse already provides correctly.

Examples:

Prefer:

```text
ClickHouse replication
Keeper coordination
ClickHouse backup primitives
ClickHouse health/system tables
```

over inventing:

```text
file synchronization controller
custom consensus
custom table copying
custom catalog replication
```

The operator orchestrates.

The database owns database semantics.

---

# 61. Best-Practice Principle: Fail Safe

For destructive operations:

```text
uncertain health
=>
do less
```

Examples:

- do not remove another replica if replication is unhealthy
- do not continue an upgrade after a failed replacement
- do not delete backup metadata if external backup cleanup is uncertain
- do not scale below an HA safety threshold without explicit policy

Database control planes should bias toward preserving data and capacity.

---

# 62. Research References

The following are the most useful projects and documents for this roadmap.

## ClickHouse

- Official ClickHouse Kubernetes Operator  
  https://github.com/ClickHouse/clickhouse-operator

- Introducing the Official ClickHouse Kubernetes Operator  
  https://clickhouse.com/blog/clickhouse-kubernetes-operator

- Make Before Break — Faster Scaling Mechanics for ClickHouse Cloud  
  https://clickhouse.com/blog/make-before-break-faster-scaling-mechanics-for-clickhouse-cloud

- ClickHouse Cloud  
  https://clickhouse.com/cloud

## Altinity

- Altinity ClickHouse Operator  
  https://github.com/Altinity/clickhouse-operator

## PostgreSQL Operators

- CloudNativePG  
  https://github.com/cloudnative-pg/cloudnative-pg

- CloudNativePG documentation  
  https://cloudnative-pg.io/

- Crunchy Postgres Operator  
  https://github.com/CrunchyData/postgres-operator

- Crunchy PGO Architecture  
  https://access.crunchydata.com/documentation/postgres-operator/latest/architecture

## Distributed Systems Operators

- Strimzi Kafka Operator  
  https://github.com/strimzi/strimzi-kafka-operator

- Strimzi documentation  
  https://strimzi.io/docs/operators/latest/

- TiDB Operator  
  https://github.com/pingcap/tidb-operator

- TiDB Operator architecture  
  https://docs.pingcap.com/tidb-in-kubernetes/dev/architecture/

- Vitess Operator  
  https://github.com/planetscale/vitess-operator

## Control Plane Design

- Crossplane  
  https://github.com/crossplane/crossplane

- Crossplane documentation  
  https://docs.crossplane.io/

## Kubernetes

- Custom Resources  
  https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/

- Finalizers  
  https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/

- Kubernetes API conventions  
  https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md

---

# 63. Final Recommended Project Definition

Use this as the working one-sentence definition:

> **A Go-based managed ClickHouse control plane that exposes a public cloud API and reconciles customer service intent through a custom Kubernetes operator into highly available ClickHouse and Keeper deployments.**

And this as the engineering focus:

> **The project explores the boundary between cloud control-plane APIs, Kubernetes reconciliation, and database-aware lifecycle management, with an emphasis on failure recovery, safe upgrades, backup/restore, and eventually make-before-break operations.**

That scope is coherent, technically deep, and directly useful as a portfolio project.
