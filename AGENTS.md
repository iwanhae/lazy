# Repository Guidelines

## Project Structure & Module Organization
- Root Go package `lazy`: `new.go`, `map.go`, `filter.go`, `with.go`, `consume.go`.
- Tests in root as `*_test.go` (both `lazy_test` black-box and a small white-box test in `package lazy`).
- Examples by case: `examples/{case}/main.go` (default `simple/`). Demonstrates `NewSlice → Filter → Map → Consume`. Available cases: `simple`, `errors`, `cancellation`.
- Module: `github.com/iwanhae/lazy` (see `go.mod`). Target Go 1.24.

## Example Outputs
- Each example case has an `OUTPUT` file under `examples/{case}/OUTPUT` with the expected stdout.
- Run `make test-examples` to execute all cases and verify actual output matches `OUTPUT` (uses `diff`).

## Contribution Workflow
- Implement changes in small, focused commits.
- Always add or update tests for your changes, then run `make test` locally and ensure it passes before pushing.
- Add or update `examples/{case}/main.go` (e.g., `examples/simple/main.go`) when user-visible behavior changes.
- When adding a new operator (Map/Filter-level change), update AGENTS.md: Operator API Contracts, Error Policy Matrix, New Operator Checklist, and Examples guidance.

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

Each operator should document these properties succinctly:

- Input: upstream `object[T]` and user function signature
- Output: downstream `object[U]` (or `object[T]` for Filter)
- Order: preserves input order for emitted values
- Cancellation: must respect `ctx.Done()` before send
- Errors: handled by `WithErrHandler` → `DecisionStop|DecisionIgnore`
- Buffering: output channel size via `WithSize`

Current operators:

- NewSlice
  - Input: `slice []T`
  - Output: `object[T]`
  - Order: preserved; emits sequentially
  - Cancellation: stops emission when `ctx.Done()`
  - Errors: none
  - Buffering: `WithSize`

- Map
  - Input: `object[IN]`, `mapper(IN) (OUT, error)`
  - Output: `object[OUT]`
  - Order: preserved for successful results
  - Cancellation: select on `ctx.Done()` before send
  - Errors: handler decides stop vs. drop
  - Buffering: `WithSize`

- Filter
  - Input: `object[T]`, `predicate(T) (bool, error)`
  - Output: `object[T]` (passes through accepted values)
  - Order: preserved for emitted values
  - Cancellation: select on `ctx.Done()` before send
  - Errors: handler decides stop vs. drop
  - Buffering: `WithSize`

## New Operator Checklist

- Accept `context.Context` as first arg.
- Build options via `buildOpts(opts)`.
- Allocate output channel with `make(chan X, opt.size)`.
- Launch a goroutine; at top: `defer recover()` and `defer close(ch)`.
- Iterate `for v := range obj.ch { ... }`.
- On error from user func: `if opt.onError(err) == DecisionStop { return } else { continue }`.
- Before sending: `select { case <-ctx.Done(): return; case ch <- out: }`.
- Do not leak goroutines on cancellation or stop.

Minimal skeleton:

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

## Error Handling Policy Matrix

| Operator  | Error Source      | DecisionStop            | DecisionIgnore           |
|-----------|-------------------|-------------------------|--------------------------|
| NewSlice  | N/A               | N/A                     | N/A                      |
| Map       | mapper error      | stop pipeline, close ch | drop value, continue     |
| Filter    | predicate error   | stop pipeline, close ch | drop value, continue     |

## Concurrency Invariants

- Always guard sends with `select` on `ctx.Done()`.
- Always `defer close(out)` in the goroutine producing `out`.
- Treat each operator as the sole owner of its output channel.
- Never block indefinitely on downstream; cancellation must unblock sends.
- Preserve order; do not fan-out/fan-in unless explicitly documented.

## Examples: Rules of Thumb

- Add a new example when behavior is user-visible or educational (errors, cancellation, new operator).
- Each example must include `main.go` and an `OUTPUT` with expected stdout.
- Use `make test-examples` to verify outputs via `diff`.
- Keep outputs small and deterministic; avoid time-based flakiness.

## Make Targets Guide

- fmt: run `gofmt -s -w .`
- vet: run `go vet ./...`
- test-examples: run all examples and diff against `OUTPUT`
- test: tidy → fmt → vet → test-examples → `go test -v -cover ./...`

## Running in Restricted Environments

- Set absolute caches to avoid sandbox issues:
  - `GOCACHE=$(pwd)/.gocache`
  - `GOMODCACHE=$(pwd)/.gocache-mod`
  - `GOTOOLCACHE=$(pwd)/.gocache-tools`
- Ensure `go mod tidy` runs before tests to vendor deps in cache.

## PR Checklist

- Code formatted, `go vet` clean.
- Unit tests updated/added; goleak checks present and pass.
- Examples updated and `OUTPUT` verified via `make test-examples`.
- Docs updated (AGENTS.md and example descriptions) if behavior changed.
- For new operators (Map/Filter-level), ensure AGENTS.md sections are updated (contracts, error matrix, checklist, examples).
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
- Exported APIs include a concise doc comment with the contract above.
