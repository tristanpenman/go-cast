# Conventions

These conventions complement standard Go guidance.

## Core Guidelines

### Structure

- Keep reusable logic in packages. Limit `cmd/<name>` to arguments, dependency construction, lifecycle, and UI wiring.
- Give cohesive, independently testable capabilities their own packages.
- Keep focused tests beside their package. Use a separate test package for black-box API tests. Top-level `test/` should only contain cross-package black-box coverage.
- Treat directories with their own manifest or instructions as separate components and follow their local guidance.

### Packaging

- Use `pkg/` for publicly importable code intended for consumers outside the repository. Treat its exported API as a compatibility commitment.
- Use `internal/` for code that supports the repository's `cmd/` programs and is not intended as a public library.
- Keep command-specific wiring in `cmd/`; move shared command implementation into focused `internal/` packages.
- For repositories that are pure libraries rather than collections of programs, prefer top-level packages over `pkg/` to keep import paths short.
- Choose package boundaries by responsibility and dependency direction, not merely by file size.

### Errors and logging

- Reusable packages return errors with concise operation context:

  ```go
  return nil, fmt.Errorf("discover devices: %w", err)
  ```

- Commands report actionable errors and clean up. Panic only for unrecoverable process initialization.
- Use short structured log events with key/value context. Log full payloads only at debug level.
- Never log private keys, authentication secrets, salts, or similar data.
- Accept a logger from the owning layer where practical. Preserve the logging approach already established in an area.

### APIs and implementation

- Constructors normally use `New<Type>`. Preserve established names when compatibility matters; new fallible constructors should return an error.
- Export only what another package needs and document new exported APIs.
- Inject external effects through narrow function or interface seams.
- Return stable, presentation-ready collections when promised by the API. Preserve useful partial results alongside errors.
- Keep parsing and conversion deterministic and side-effect free.

### Formatting and tests

- Preserve a file's existing import grouping and run `gofmt` on changed Go files.
- Exclude generated code, reference documents, diagrams, and binary assets from broad formatting or replacement.
- Unit tests must not require devices, multicast networking, displays, hosted services, or cloud accounts.
- Cover success, failure, closure, timeout, fallback, deduplication, and stable ordering as relevant.
- Run focused package tests first, then `go test ./...` when integration dependencies permit.

## Idioms

These are primarily idioms that apply to this project. They might not apply elsewhere.

### Data and protocols

- Preserve domain vocabulary at boundaries. Use Go initialism casing in new code (`appID`, `sourceID`, `SSRC`) without renaming established APIs incidentally.
- Keep shared constants in the shared protocol layer and feature-specific identifiers with their feature.
- Use small wire structs with explicit tags. Keep envelopes and request types private unless callers must construct them.
- Pass routing, request, and session identifiers explicitly.
- Dispatch by message category before decoding its body. Keep control and application traffic separate.
- Validate required fields at boundaries. Ignore unknown messages only when continuing cannot corrupt framing or state.

### Concurrency and lifecycle

- A channel producer owns closing its channel; consumers range over it.
- Network readers publish complete messages. Close the connection to unblock reads during shutdown.
- Long-lived components expose `Close`, `Stop`, or an equivalent lifecycle operation. Make cleanup safe on error paths and wait for required goroutines.
- Protect mutable component state with its owning mutex. Return copied snapshots rather than mutable slices or maps.
- Signal condition variables while holding their mutex.
- Time-bounded waits distinguish timeout, remote failure, and local closure.

### Logging

- Core components use an injected logger or `common.NewLogger`; small commands may retain their established standard-library logger.
- Do not add a logging dependency solely to standardize existing areas.

### Desktop and frontend

- Keep framework wiring in `cmd/remote` and reusable logic in packages.
- `cmd/remote/frontend/dist` is dependency-free and embedded. Do not add a JavaScript build step for small changes.

## Integrations

### Protobuf

- Treat `.proto` files as authoritative. Never hand-edit the corresponding `.pb.go` files, regenerate them using `protoc`.
- Explicitly set required outbound message fields. Contain generated pointer semantics at the integration boundary.
- Keep binary generated messages separate from UTF-8 JSON messages.

### Discovery

- Convert discovery-library results into the project model at the boundary.
- Deduplicate advertisements, prefer advertised addresses, apply metadata fallbacks, sort by friendly name, and preserve partial results with errors.
- Inject discovery in tests; never perform live multicast queries.

### Graphics and media

- Keep graphics context work on the required OS thread and preserve receiver thread locking.
- Keep headless media handling independent of graphics.
- Full builds may require native RTP, codec, and graphics dependencies even when protocol package tests do not.
