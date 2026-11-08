#!/usr/bin/env sh
set -eu

DOCS_DIR="$(CDPATH= cd -- "$(dirname "$0")" && pwd)"
HOST_DIR="$(CDPATH= cd -- "${DOCS_DIR}/.." && pwd)"
TMP_V3_DIR="${DOCS_DIR}/.tmp-v3"

run_swag_v1() {
	go run github.com/swaggo/swag/cmd/swag@v1.16.4 "$@"
}

run_swag_v2() {
	go run github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc5 "$@"
}

echo "[1/3] Generating Swagger 2.0 docs"
(
	cd "${HOST_DIR}"
	run_swag_v1 init --parseDependencyLevel 3 --output "${DOCS_DIR}"
)

echo "[2/3] Generating OpenAPI 3.1 docs"
rm -rf "${TMP_V3_DIR}"
(
	cd "${HOST_DIR}"
	run_swag_v2 init --parseDependencyLevel 3 --v3.1 --output "${TMP_V3_DIR}" --outputTypes json,yaml
)

echo "[3/3] Writing .v3 files"
mv "${TMP_V3_DIR}/swagger.json" "${DOCS_DIR}/swagger.v3.json"
mv "${TMP_V3_DIR}/swagger.yaml" "${DOCS_DIR}/swagger.v3.yaml"
rm -rf "${TMP_V3_DIR}"

echo "Done"
