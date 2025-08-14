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

- New
  - Input: `in <-chan T` (user channel)
  - Output: `object[T]` forwarding from `in`
  - Order: preserved; forwards sequentially
  - Cancellation: stops forwarding when `ctx.Done()`
  - Errors: none
  - Buffering: `WithSize` on the forwarded output

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
| New       | N/A               | N/A                     | N/A                      |
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
