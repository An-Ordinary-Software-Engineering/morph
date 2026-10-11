# ADR-0001: Record architecture decisions

- **Status:** Accepted
- **Date:** 2026-10-11

## Context

morph will be published as a reusable package. Whoever comes later, whether contributors or the author
a few months from now, needs to understand why the structure and the API are the way they are, not only what they are.

## Decision

We will record relevant architecture decisions as ADRs in `docs/adr/`, in Michael Nygard's format, in English (`docs/adr/en-US/`) and Portuguese (`docs/adr/pt-BR/`).
The index and the template live at the root of `docs/adr/`, in English only.
The consolidated, current view of the architecture lives in `docs/ARCHITECTURE.en_US.md` and `docs/ARCHITECTURE.pt_BR.md`.

## Consequences

- A decision that affects the structure, the public API or dependencies requires an ADR.
- Accepted ADRs are immutable. Changes come as new ADRs.
- Both language versions (ADR and `ARCHITECTURE`) are updated together.
