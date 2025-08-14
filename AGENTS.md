# Repository Guidelines

This document is designed to be evergreen. It focuses on invariant policies,
checklists, and small reusable templates rather than enumerating current files
or APIs that frequently change. The code and tests are the source of truth for
implementation details; AGENTS.md describes contracts each change must honor.

## Source of Truth & Drift Policy
- Code comments on exported APIs + tests are the primary specification.
- AGENTS.md captures stable invariants (ordering, cancellation, error policy),
  category-level contracts (Source/Transform/Sink), and contributor checklists.
- Prefer patterns over lists: avoid enumerating files, operators, or example
  cases here. The repo layout and discovery rules make those discoverable.
- Update AGENTS.md only when invariants, policies, or contributor workflows
  change. Adding a new operator that follows the same contracts does not
  require editing this file beyond adding doc comments and tests in code.

## Project Structure & Module Organization
- Root Go package `lazy`. Operators live as focused files in the root (e.g.,
  `new.go`, `map.go`, `filter.go`, `with.go`, `consume.go`). Names may evolve,
  but each operator is in a single, small file.
- Tests live in root as `*_test.go` with a majority in `package lazy_test` and
  minimal white-box tests in `package lazy` only when necessary (e.g., buffers).
- Examples live under `examples/{case}` with `main.go` and `OUTPUT`.
- Module: `github.com/iwanhae/lazy` (see `go.mod`). Target Go 1.24.

## Example Outputs
- Each example case has an `OUTPUT` file under `examples/{case}/OUTPUT` with the expected stdout.
- Run `make test-examples` to execute all cases and verify actual output matches `OUTPUT` (uses `diff`).

## Contribution Workflow
- Implement changes in small, focused commits.
- Always add or update tests for your changes, then run `make test` locally and ensure it passes before pushing.
- Add or update `examples/{case}/main.go` and `OUTPUT` when user-visible behavior changes.
- When adding a new operator, ensure: operator doc comment (use template below),
  tests updated per matrix, and examples if behavior is user-visible. Update
  AGENTS.md only if you change invariants or contributor workflows.

## Documentation Expectations
- Briefly document user-visible behavior with operator comments and a minimal example.
- Update AGENTS.md whenever an operator is added/changed/removed so contracts and checklists stay current.
- Keep docs actionable: inputs/outputs, ordering, cancellation, error handling, buffering options.

## Coding Style & Naming Conventions
- Use standard Go formatting (tabs, gofmt). Keep imports organized.
- Package remains `lazy`; unexported internals (e.g., `object[T]`) stay unexported.
- Options follow `WithX` pattern (e.g., `WithSize`, `WithErrHandler`). Decisions are `lazy.DecisionStop` or `lazy.DecisionIgnore`.
- Prefer clear, single-purpose functions; keep generics type parameters short and meaningful (`IN`, `OUT`, `T`).

## Testing Guidelines
- Mandatory: before writing any test code, read `DEVELOPMENT_TEST_CODE.md` end-to-end and follow it.
- Canonical guide: see `DEVELOPMENT_TEST_CODE.md` for detailed patterns and rationale.
- Quick run: `make test` (add `make race` and `make cover` as needed).
- Prefer black-box tests in `package lazy_test`; use `package lazy` only for explicit white-box checks.
- Include `defer goleak.VerifyNone(t)` to detect goroutine leaks.

## Commit & Pull Request Guidelines
- Commits: short, imperative summaries (e.g., “Add tests for mapping errors”).
- PRs: include a concise description, linked issues, rationale for API changes, and test updates. Add or update `examples/main.go` when behavior changes.
- CI expectation: all `go vet`, formatting, race, and unit tests pass locally before requesting review.

## Concurrency & API Tips
- Always close output channels and exit goroutines on cancellation to avoid leaks (goleak enforced).
- Preserve input order in transformations; document any intentional deviations.
- When adding operators, accept `context.Context`, propagate errors via `WithErrHandler`, and expose buffering via `WithSize`.

---

## Operator API Contracts (Quick Reference)

Each exported operator must include a concise doc comment using this template.

Operator doc comment template:

```
// <Name> <short description>
//
// Input: <upstream object and user func signature>
// Output: <downstream object type>
// Order: preserves input order for emitted values
// Cancellation: guards sends with select on ctx.Done()
// Errors: handled via WithErrHandler → DecisionStop | DecisionIgnore
// Buffering: output channel capacity via WithSize
```

Operator categories and contracts:
- Source (e.g., from slice or user channel)
  - Input: user-owned data (`[]T` or `<-chan T`)
  - Output: `object[T]`
  - Order: preserved
  - Cancellation: stops emission/forwarding when `ctx.Done()`
  - Errors: none
  - Buffering: `WithSize`
- Transform (e.g., Map/Filter-family)
  - Input: `object[IN]` and user function
  - Output: `object[OUT]` (or `object[T]` for Filter)
  - Order: preserved for emitted values
  - Cancellation: guard sends with `select { case <-ctx.Done(): return; case out <- v: }`
  - Errors: `WithErrHandler` decides Stop (halt pipeline) vs Ignore (drop value)
  - Buffering: `WithSize`
- Sink (e.g., Consume)
  - Input: `object[T]` and consumer func
  - Behavior: drain values; return first consumer error
  - Cancellation: N/A (respects upstream closure)

## New Operator Checklist

- Accept `context.Context` as first arg.
- Build options via `buildOpts(opts)`.
- Allocate output channel with `make(chan X, opt.size)`.
- Launch a goroutine; at top: `defer recover()` and `defer close(ch)`.
- Iterate `for v := range obj.ch { ... }`.
- On error from user func: `if opt.onError(err) == DecisionStop { return } else { continue }`.
- Before sending: `select { case <-ctx.Done(): return; case ch <- out: }`.
- Do not leak goroutines on cancellation or stop.

Minimal skeleton (transform):

```go
func Op[IN any, OUT any](ctx context.Context, in object[IN], f func(IN) (OUT, error), opts ...optionFunc) object[OUT] {
    opt := buildOpts(opts)
    ch := make(chan OUT, opt.size)
    go func() {
        defer recover()
        defer close(ch)
        for v := range in.ch {
            out, err := f(v)
            if err != nil {
                if opt.onError(err) == DecisionStop { return }
                continue
            }
            select { case <-ctx.Done(): return; case ch <- out: }
        }
    }()
    return object[OUT]{ch: ch}
}
```

## Error Handling Policy

- Source operators do not produce user-function errors.
- Transform operators handle user-function errors via `WithErrHandler`:
  - DecisionStop: stop the pipeline and close the output channel.
  - DecisionIgnore: drop the failing value and continue.
- Sink operators must propagate consumer errors immediately (no wrapping unless intentional).

## Concurrency Invariants

- Always guard sends with `select` on `ctx.Done()`.
- Always `defer close(out)` in the goroutine producing `out`.
- Treat each operator as the sole owner of its output channel.
- Never block indefinitely on downstream; cancellation must unblock sends.
- Preserve order; do not fan-out/fan-in unless explicitly documented.

## Examples: Rules of Thumb

- Add a new example when behavior is user-visible or educational (errors, cancellation, new operator, or new source patterns).
- Each example must include `main.go` and an `OUTPUT` with expected stdout.
- Do not enumerate cases here; `make test-examples` discovers all `examples/*` with both files present.
- Keep outputs small and deterministic; avoid time-based flakiness.

## Make Targets Guide

- fmt: run `gofmt -s -w .`
- vet: run `go vet ./...`
- test-examples: run all examples and diff against `OUTPUT`
- test: format → vet → examples check → `go test -v -cover ./...` (offline by default via vendoring)
- vendor: populate `./vendor` from `go.mod`/`go.sum`
- tidy-vendor: `go mod tidy` then refresh `./vendor`

## Running in Restricted Environments

- Default is offline-friendly: the Makefile exports local caches and forces vendoring (`GOFLAGS=-mod=vendor`).
- Just run `make test`; no environment variables are required.
- When changing dependencies, run `make tidy-vendor` once (requires network) and commit updated `go.mod`, `go.sum`, and `vendor/`.

## PR Checklist

- Code formatted, `go vet` clean.
- Unit tests updated/added; goleak checks present and pass.
- Examples updated and `OUTPUT` verified via `make test-examples`.
- Docs updated (AGENTS.md and example descriptions) if behavior changed.
- If dependencies changed, run `make tidy-vendor` and commit `vendor/`.
- For new operators, ensure operator doc comments and tests are added; update AGENTS.md only if policies or checklists change.
- Coverage roughly maintained; call out intentional gaps.

## Testing Matrix (per operator)

- Happy path outputs (order preserved).
- Default ignore on user-func error (drops, continues).
- Stop on error with `WithErrHandler(DecisionStop)`.
- Empty input yields no outputs, no leaks.
- Buffer size honored (white-box `cap(out.ch)` in `package lazy`).
- Cancellation respected (stop mid-stream; no leaks).

## Naming & Docs Conventions

- Generics: short, meaningful (`IN`, `OUT`, `T`).
- Options: `WithX` pattern; decisions: `DecisionStop`, `DecisionIgnore`.
- Exported APIs include a concise doc comment using the template under Operator API Contracts.

## Drift Minimization Tips
- Avoid listing current operators or example names in docs. Let discovery rules and code comments speak for specifics.
- Place the full operator contract in code comments near the implementation.
- Prefer category-level guidance here (Source/Transform/Sink) and enforce via tests.
- If an operator deviates from standard contracts, call it out explicitly in its code comment and add targeted tests.

---

## Contributor Quick Guide

### Where To Look
- Source operators: `new.go` (e.g., `NewSlice`, `New`)
- Transform operators: `map.go`, `filter.go`
- Sink: `consume.go`
- Options/policies: `with.go` (`WithSize`, `WithErrHandler`, defaults)
- Examples/verification: `examples/{case}` and `Makefile` (`test-examples`)

### Quick Change Recipes
- Source add (New-family): edit `new.go` → add tests (`*_test.go` black-box + `size_test.go` white-box) → add example under `examples/{case}` with `OUTPUT` → update `AGENTS.md` contracts/matrix.
- Transform add (Map/Filter-family): implement operator file → honor error policy (Stop/Ignore) and cancellation → add test matrix (happy/error/cancel/buffer/empty) → update example + docs.
- Sink change/add (Consume-family): confirm consumer error propagation → add happy/error tests → verify example output.

### Contracts Cheat Sheet
- Cancellation: guard every send with `select { case <-ctx.Done(): return; case out <- v: }`.
- Channel ownership: each operator solely owns and closes its output (`defer close(out)`).
- Ordering: preserve input order; no fan-out/fan-in unless documented.
- Error default: absent `WithErrHandler`, behave as `DecisionIgnore`.
- Buffer default: absent `WithSize`, buffer size is `0`.
- Source errors: none (N/A in the error policy matrix).

### Testing Checklist
- Leak check: `defer goleak.VerifyNone(t)`.
- Happy path: values and order.
- Error policy: Ignore (drop) vs Stop (halt) branches.
- Cancellation: stop mid-stream; goroutines exit.
- Buffering: white-box `cap(out.ch)` where relevant.
- Empty input: zero outputs, no leaks.

### Examples Rules
- Structure: `main.go` + `OUTPUT` per case.
- Verification: `make test-examples` must diff-equal; avoid extra whitespace/newlines.
- Deterministic: avoid time/rand-driven output.

### PR/Commit Gate
- Formatting and vet: `gofmt`, `go vet` clean.
- Offline-friendly: `make test` (vendor enforced) passes.
- Docs: update `AGENTS.md` (contracts, matrix, checklists) for user-visible changes.
