#!/usr/bin/env bash
# Prove Kind local-path keeps MergeTree data across a pod delete, and that
# headless DNS names the ordinal. This is Kubernetes restart, not ClickHouse
# replication. Kind standard is node-local disk, not cloud durability.
set -euo pipefail

NS="${CLICKHOUSE_NS:-clickhouse-lab}"
STS="${CLICKHOUSE_STS:-clickhouse}"
PASSWORD="${CLICKHOUSE_PASSWORD:-clickhouse-lab}"
HEADLESS_FQDN="${CLICKHOUSE_HEADLESS_FQDN:-clickhouse-0.clickhouse-headless.clickhouse-lab.svc.cluster.local}"

ch() {
	kubectl -n "${NS}" exec "sts/${STS}" -- \
		clickhouse-client --password "${PASSWORD}" --query "$1"
}

# Readiness is HTTP GET /ping on 8123. clickhouse-client uses native 9000, which
# can still refuse connections for a few seconds after the pod is Ready.
wait_native() {
	echo "waiting for native protocol :9000"
	local i
	for i in $(seq 1 60); do
		if kubectl -n "${NS}" exec "sts/${STS}" --request-timeout=15s -- \
			clickhouse-client --password "${PASSWORD}" --query "SELECT 1" >/dev/null 2>&1; then
			return 0
		fi
		sleep 2
	done
	echo "clickhouse-client never accepted connections on localhost:9000" >&2
	exit 1
}

echo "waiting for ${STS} in ${NS}"
kubectl rollout status "sts/${STS}" -n "${NS}" --timeout=10m
wait_native

echo "creating lab.persist"
ch "CREATE DATABASE IF NOT EXISTS lab"
ch "CREATE TABLE IF NOT EXISTS lab.persist (id UInt32) ENGINE = MergeTree ORDER BY id"
ch "TRUNCATE TABLE lab.persist"
ch "INSERT INTO lab.persist VALUES (1), (2), (3)"

before="$(ch "SELECT count() FROM lab.persist" | tr -d '[:space:]')"
if [[ "${before}" != "3" ]]; then
	echo "expected 3 rows before delete, got ${before}" >&2
	exit 1
fi

echo "deleting pod ${STS}-0 (StatefulSet must recreate it on the same PVC)"
kubectl -n "${NS}" delete pod "${STS}-0" --wait=true
kubectl rollout status "sts/${STS}" -n "${NS}" --timeout=10m
wait_native

after="$(ch "SELECT count() FROM lab.persist" | tr -d '[:space:]')"
if [[ "${after}" != "3" ]]; then
	echo "expected 3 rows after pod recreate, got ${after}" >&2
	exit 1
fi

echo "resolving ${HEADLESS_FQDN} from inside the pod"
dns="$(kubectl -n "${NS}" exec "sts/${STS}" -- getent hosts "${HEADLESS_FQDN}")"
if [[ -z "${dns}" ]] || ! grep -q "clickhouse-0" <<<"${dns}"; then
	echo "headless DNS did not resolve ${HEADLESS_FQDN}: ${dns}" >&2
	exit 1
fi

echo "ok: lab.persist count=3 after pod delete; ${HEADLESS_FQDN} -> ${dns}"
