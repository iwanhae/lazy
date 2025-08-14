# Testing Strategy for `lazy`

This document explains how we approach testing for the `lazy` package. It is intended for maintainers and contributors to quickly understand what to test, how to test it, and common pitfalls when validating concurrent, channel-based code.

## Goals

- Validate correctness of the lazy pipeline (source → transform → consume).
- Ensure robust error handling with clear, intentional behavior.
- Prove that cancellation propagates properly and goroutines terminate.
- Keep ordering guarantees clear and enforced.
- Avoid goroutine leaks and hidden deadlocks.

## Principles

- Prefer deterministic synchronization over sleeps; if a sleep is unavoidable, keep it tiny and document why.
- Test one concern per test; combine concerns only when it improves clarity.
- Default to black-box tests; use white-box only when necessary (e.g., verifying buffer size via channel capacity).
- Always verify no leaked goroutines via goleak to catch subtle lifecycle issues.
- Keep tests readable and small; assert on minimal, essential outcomes.

## Test Categories

1. Mapping behavior
   - Identity and transformation mapping (`v -> v`, `v -> f(v)`)
   - Multiple elements, verify all outputs are produced.

2. Error handling
   - Default handler (ignore): elements that fail mapping are dropped, pipeline continues.
   - Stop-on-error: with `WithErrHandler(DecisionStop)` the pipeline halts at the first error; only preceding outputs appear.

3. Consumer error propagation
   - `Consume` must return consumer errors immediately (no swallowing or wrapping unless intentional).

4. Context cancellation
   - Cancellation during consumption should stop the pipeline quickly (both upstream source and transforms terminate).
   - Use explicit `cancel()` in the consumer callback after some items are seen.

5. Order preservation
   - When errors are ignored, successful outputs retain input order.

6. Empty input
   - No outputs, no errors, no leaks.

7. Buffer size (white-box)
   - With `WithSize(n)` the internal channel capacity matches `n` for both `NewSlice` and `Map`.

## Goleak Usage

- Add `defer goleak.VerifyNone(t)` at the top of each test.
- Ensure `cancel()` happens before goleak verification runs. Use `defer cancel()` (not `t.Cleanup(cancel)`) so that all goroutines are signaled to stop before goleak checks execute.
- If a test legitimately starts background goroutines that outlive the test (discouraged here), use goleak filters (e.g., `goleak.IgnoreTopFunction(...)`). Prefer redesigning the test to avoid this.

## Concurrency/Channels Gotchas

- Avoid blocking sends/receives: prefer bounded buffers via `WithSize` when a test needs throughput.
- Rely on `context.Context` to stop producers/transforms; do not rely on closing input channels from tests (the library owns them).
- Do not use `t.Parallel()` with goleak unless you are confident that parallel scheduling will not confound leak detection.
- Never ignore `ctx.Done()` paths in helper code used by tests.

## Cancellation Determinism

- Always gate sends with context: when an operator forwards a result, use `select { case <-ctx.Done(): return; case ch <- result: }` instead of a bare `ch <- result`. Bare sends can slip one extra value through after cancellation, leading to nondeterministic counts.
- DecisionIgnore still drops failing items, but should `continue` the loop rather than fall through to a send.
- The `examples/cancellation` case cancels after consuming 5 items; it must consistently report `consumed before cancel: 5`. A previous race allowed 6 due to a non-selecting send in `Map`. This was fixed by wrapping the send with a `select` on `ctx.Done()`.
- When adding new operators, follow the same pattern: respect `ctx.Done()` before compute and before send; close output channels and exit goroutines promptly on cancellation.

## Structure and Naming

- One behavior per test function, named as: `Test<Subject>_<Behavior>`
  - Examples: `TestMap_DefaultIgnoreError`, `TestMap_StopOnError`, `TestContextCancellationStopsPipeline`.
- Keep test input small and explicit; use slices for clarity over generated ranges unless size is integral to the behavior.

## Running Tests

`make test`

## Future Enhancements

- Property-based tests for mapping and cancellation (e.g., using `testing/quick`).
- Fuzzing around error handlers and panics to validate `recover`-guards in goroutines.
- Benchmarks for `NewSlice`, `Map`, and `Consume` under different buffer sizes to guide default tuning.
- Additional operators (e.g., Filter, FlatMap) should adopt the same testing patterns.

## Example Pattern

```go
func TestMap_StopOnError(t *testing.T) {
    defer goleak.VerifyNone(t)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    src := lazy.NewSlice(ctx, []int{1,2,3,4})
    out := lazy.Map(ctx, src, func(v int) (int, error) {
        if v == 3 { return 0, errors.New("boom") }
        return v, nil
    }, lazy.WithErrHandler(func(err error) lazy.Decision { return lazy.DecisionStop }))

    var got []int
    if err := lazy.Consume(out, func(v int) error { got = append(got, v); return nil }); err != nil {
        t.Fatalf("consume: %v", err)
    }

    want := []int{1,2}
    if !reflect.DeepEqual(got, want) {
        t.Fatalf("got=%v want=%v", got, want)
    }
}
```
