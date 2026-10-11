# ADR-0005: Versioning and API stability

- **Status:** Proposed
- **Date:** 2026-10-11

## Context

Public Go modules are resolved through git tags and follow SemVer strictly. From `v2` on, the import path changes (`/v2`).
An incompatible change published by mistake in a `v1.x` release breaks every user.

## Decision

1. We will use **SemVer with `vX.Y.Z` tags** at the repository root.
2. The project stays at **`v0.x` until ADRs 0002–0004 are implemented and stable**. During this phase, incompatible changes bump the `minor`.
3. **`v1.0.0`** ships only when there are at least two dialects and `Create`, `Alter` and `Delete` are complete.
4. Relevant changes are recorded in `CHANGELOG.md`, following [Keep a Changelog](https://keepachangelog.com/en/).
5. The public surface stays minimal: we only export what has a use case. When in doubt, the code goes into `internal/`.
6. The `go` version declared in `go.mod` is the oldest one still supported by the Go team, that is, the last two *major releases*.

## Consequences

- Users can pin versions safely, and `go get -u` breaks nothing from `v1` on.
- Each release needs a tag and a CHANGELOG entry.
- CI must run on at least the two supported Go versions.
