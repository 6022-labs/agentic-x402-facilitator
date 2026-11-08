.PHONY: build test integration-test lint fmt mocks swagger abi help

# build compiles every Go package under src/.
build:
	go build ./src/...

# test runs the unit-test packages; the $(...) form propagates a failing exit code.
test:
	go test $$(go list ./tests/... | grep '_unit_tests')

# integration-test runs the integration packages; they start anvil, Redis and Vault containers, so Docker must run.
integration-test:
	go test $$(go list ./tests/... | grep '_integration_tests')

# lint requires golangci-lint v2 on PATH (go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest).
lint:
	golangci-lint run ./...

# fmt runs golines via golangci-lint's built-in formatter; no standalone golines needed.
fmt:
	golangci-lint fmt

# mocks requires the deepsearch-mockgen CLI (github.com/FournyP/deepsearch-mockgen-cli).
mocks:
	deepsearch-mockgen -S ./src -O ./tests -A -P

# swagger regenerates the host's API docs (needs the go toolchain; runs swag via go run).
swagger:
	./src/x402_facilitator_host/docs/generate_docs.sh

# abi regenerates the contract bindings (needs abigen on PATH).
abi:
	cd src/x402_facilitator_blockchain/abi && go generate .

help:
	@echo "build       - go build ./src/..."
	@echo "test        - go test unit-test packages"
	@echo "integration-test - go test integration packages (needs Docker)"
	@echo "lint        - golangci-lint run ./... (needs golangci-lint v2)"
	@echo "fmt         - golangci-lint fmt, 120-char cap"
	@echo "mocks       - regenerate mocks (needs deepsearch-mockgen)"
	@echo "swagger     - regenerate host API docs (needs go toolchain)"
	@echo "abi         - regenerate contract bindings (needs abigen)"
