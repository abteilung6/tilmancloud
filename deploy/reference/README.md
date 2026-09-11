# Study tooling (not the product data plane)

Lab ClickHouse stays the hand-written StatefulSet in `clickhouse-lab`. Nothing here should own that STS. Do not fold these targets into `cluster-up`.

## Official ClickHouse operator

Install exists so you can inspect CRDs, webhooks, and status before writing our operator. Do not apply a `ClickHouseCluster` into `clickhouse-lab`.

```bash
eval "$(make env)"
make apply-reference

kubectl get crd | grep clickhouse.com
kubectl explain clickhousecluster.spec
kubectl explain keepercluster.spec
kubectl get pods -n clickhouse-operator-system -o wide
kubectl api-resources | grep clickhouse

# must print nothing — our STS is not owned by a CR
kubectl -n clickhouse-lab get sts clickhouse -o jsonpath='{.metadata.ownerReferences}{"\n"}'

make verify-clickhouse
```

Pins are Makefile variables `CERT_MANAGER_VERSION` and `CLICKHOUSE_OPERATOR_VERSION` (release URLs, not `/latest/`). YAML is fetched to a temp file, not vendored. Operator apply is `--server-side` because combined CRDs exceed client-side apply size. Wait for cert-manager before the operator or the webhook cert never issues. If fetch TLS fails because the host clock is behind the cert `notBefore`, the script retries insecure and prints a warning.

Optional, only if Docker has plenty of RAM (second ClickHouse + Keeper). Never into `clickhouse-lab`:

```bash
kubectl create namespace clickhouse-reference
kubectl apply -n clickhouse-reference \
  -f https://raw.githubusercontent.com/ClickHouse/clickhouse-operator/v0.0.7/examples/minimal.yaml
```

## Headlamp

Kubernetes Dashboard is unmaintained. Headlamp is the in-cluster UI. Namespace `headlamp`, digest-pinned `v0.45.0`, ClusterIP only. The ServiceAccount is bound to `cluster-admin` so the UI can list every namespace on this Kind lab.

```bash
eval "$(make env)"
make apply-headlamp
# recipe prints a 24h token — paste it in the UI

kubectl -n headlamp port-forward svc/headlamp 8080:80
# open http://127.0.0.1:8080
```

Look at `clickhouse-lab` (our STS) vs `clickhouse-operator-system` (controller only) vs Custom Resources (empty until someone applies a `ClickHouseCluster`).

Do not expose Headlamp on a NodePort or Ingress. `cluster-down` deletes Kind; the operator, cert-manager, and Headlamp go with it.
