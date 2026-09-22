#!/usr/bin/env bash
# Same persist+DNS check as the lab, against the operator-owned StatefulSet.
# Fails if the lab STS picked up an ownerRef.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

owners="$(kubectl -n clickhouse-lab get sts clickhouse -o jsonpath='{.metadata.ownerReferences}')"
if [[ -n "${owners}" ]]; then
	echo "lab StatefulSet clickhouse has ownerReferences: ${owners}" >&2
	exit 1
fi

if ! kubectl -n clickhouse-managed get chs clickhouse >/dev/null 2>&1; then
	kubectl apply -f "${ROOT}/deploy/operator/sample-request.yaml"
fi

CLICKHOUSE_NS=clickhouse-managed \
	CLICKHOUSE_STS=clickhouse \
	CLICKHOUSE_HEADLESS_FQDN=clickhouse-0.clickhouse-headless.clickhouse-managed.svc.cluster.local \
	"${ROOT}/scripts/verify-clickhouse.sh"

ready="$(kubectl -n clickhouse-managed get chs clickhouse -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"
if [[ "${ready}" != "True" ]]; then
	echo "ClickHouseService Ready=${ready}" >&2
	exit 1
fi
