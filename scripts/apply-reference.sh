#!/usr/bin/env bash
# Fetch pinned cert-manager + official ClickHouse operator YAML and apply them.
# Study only — does not own the lab StatefulSet. Do not call from cluster-up.
set -euo pipefail

CERT_MANAGER_URL="${CERT_MANAGER_URL:?}"
CLICKHOUSE_OPERATOR_URL="${CLICKHOUSE_OPERATOR_URL:?}"
CLICKHOUSE_OPERATOR_NS="${CLICKHOUSE_OPERATOR_NS:-clickhouse-operator-system}"

command -v curl >/dev/null || { echo "curl required to fetch pinned release YAML"; exit 1; }

fetch() {
	local dest="$1" url="$2"
	if curl -fsSL -o "${dest}" "${url}"; then
		return 0
	fi
	echo "warning: TLS verify failed fetching ${url} (is the host clock correct?); retrying insecure" >&2
	curl -fsSLk -o "${dest}" "${url}"
}

tmp="$(mktemp -d)"
trap 'rm -rf "${tmp}"' EXIT

fetch "${tmp}/cert-manager.yaml" "${CERT_MANAGER_URL}"
kubectl apply -f "${tmp}/cert-manager.yaml"
kubectl wait --for=condition=Available deploy --all -n cert-manager --timeout=10m

fetch "${tmp}/clickhouse-operator.yaml" "${CLICKHOUSE_OPERATOR_URL}"
kubectl apply --server-side --force-conflicts -f "${tmp}/clickhouse-operator.yaml"
kubectl wait --for=condition=Available deploy --all -n "${CLICKHOUSE_OPERATOR_NS}" --timeout=10m
kubectl get crd clickhouseclusters.clickhouse.com keeperclusters.clickhouse.com
kubectl get pods -n "${CLICKHOUSE_OPERATOR_NS}"
