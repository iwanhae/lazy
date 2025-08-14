# Repository Guidelines

## Project Structure & Module Organization
- Root Go package `lazy`: `new.go`, `map.go`, `with.go`, `consume.go`.
- Tests in root as `*_test.go` (both `lazy_test` black-box and a small white-box test in `package lazy`).
- Examples by case: `examples/{case}/main.go` (default `simple/`). Demonstrates `NewSlice → Map → Consume`. Available cases: `simple`, `errors`, `cancellation`.
- Module: `github.com/iwanhae/lazy` (see `go.mod`). Target Go 1.24.

## Example Outputs
- Each example case has an `OUTPUT` file under `examples/{case}/OUTPUT` with the expected stdout.
- Run `make test-examples` to execute all cases and verify actual output matches `OUTPUT` (uses `diff`).

## Contribution Workflow
- Implement changes in small, focused commits.
- Always add or update tests for your changes, then run `make test` locally and ensure it passes before pushing.
- Add or update `examples/{case}/main.go` (e.g., `examples/simple/main.go`) when user-visible behavior changes.

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
