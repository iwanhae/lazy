SHELL := /bin/bash
PKG := ./...
CASE ?= simple
EXAMPLE := examples/$(CASE)/main.go

# Use local caches and vendor by default so `make test` works offline
export GOCACHE := $(CURDIR)/.gocache
export GOMODCACHE := $(CURDIR)/.gocache-mod
export GOTOOLCACHE := $(CURDIR)/.gocache-tools
# Force go commands to use the vendor directory
export GOFLAGS := -mod=vendor

.PHONY: all test race cover vet fmt example tidy vendor tidy-vendor test-examples

all: test

# One-stop command: format, vet, run examples check, then tests.
# Note: "tidy" is not part of default test to allow offline runs.
test: fmt vet test-examples
	@go clean -testcache
	go test -v -cover $(PKG)

vet:
	go vet $(PKG)

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

# Ensure vendor directory is populated from go.mod/go.sum
vendor:
	go mod vendor

# Convenience: tidy go.mod and refresh vendor at once
tidy-vendor: tidy vendor

example:
	go run $(EXAMPLE)

# Verify each example's output matches its expected OUTPUT file
test-examples:
	@set -euo pipefail; \
	ok=0; fail=0; \
	for d in examples/*; do \
		[ -d "$$d" ] || continue; \
		c=$$(basename "$$d"); \
		if [ ! -f "$$d/main.go" ] || [ ! -f "$$d/OUTPUT" ]; then \
			echo "skip $$c (missing main.go or OUTPUT)"; \
			continue; \
		fi; \
		echo "==> $$c"; \
		if diff -u "$$d/OUTPUT" <(go run "$$d/main.go"); then \
			echo "OK $$c"; ok=$$((ok+1)); \
		else \
			echo "FAIL $$c"; fail=$$((fail+1)); \
		fi; \
	done; \
	echo "Examples OK: $$ok, FAIL: $$fail"; \
	test $$fail -eq 0
