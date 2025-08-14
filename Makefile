SHELL := /bin/bash
PKG := ./...
CASE ?= simple
EXAMPLE := examples/$(CASE)/main.go

.PHONY: all test race cover vet fmt example tidy test-examples

all: test

# One-stop command: tidy, format, and run tests with race and coverage
test: tidy fmt vet test-examples
	@go clean -testcache
	go test -v -cover $(PKG)

vet:
	go vet $(PKG)

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

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
