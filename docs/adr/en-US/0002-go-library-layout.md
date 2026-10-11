# ADR-0002: Go library layout

- **Status:** Proposed
- **Date:** 2026-10-11

## Context

morph has two responsibilities with different audiences:

- **Declaring** migrations (`Up`/`Down`). Client code does this, so this part must be public and stable.
- **Generating and running** the DDL script from those declarations. This is an internal detail, triggered by the `mph` CLI ([ADR-0006](0006-mph-cli-and-generated-program.md)).

The official reference ([go.dev/doc/modules/layout](https://go.dev/doc/modules/layout)) suggests keeping code at the module root.
Even so, we want the boundary between what the client uses and what is internal to show in the folder structure itself,
and we want the compiler to help enforce that boundary.

## Decision

1. **`pkg/` = public declarative API.** This is what clients use to write migrations:
   - `pkg/migrator`: the `Migrator` (`CreateTable`, `DropTable`, ...) and the `Migration` interface;
   - `pkg/constraints`: declarations accumulated incrementally (`table`, `columns`);
   - `pkg/types`: **database-neutral** column types and default values;
   - `pkg/runner`: a thin entry point (`Main`) used by the program the CLI generates.

   Nothing in `pkg/` produces SQL.
2. **`internal/src/` = generation and execution.** The compiler blocks imports from outside the module:
   - `internal/src/builder`: the `script_builder` walks the declarations and assembles the DDL script;
   - `internal/src/provider`: per-database factories ([ADR-0007](0007-providers-abstract-factory.md));
   - `internal/src/runner`: orders migrations, manages the version table and runs the script.
3. **Module root = `morph` package, a single alias.** It only re-exports what is in `pkg/` (type aliases, function wrappers and constants),
   so clients import a single package.
4. **`cmd/mph/` = CLI.** It lives in the same module and can therefore use `internal/`.
5. Package names are short, singular, and never `util`, `common` or `helpers`.

### Dependency rules

- `internal/src/*` may import `pkg/*`. This is how the builder reads the declarations.
- **`pkg/*` never exposes `internal/` types in exported signatures.** Otherwise clients receive a type they cannot name.
  Only `pkg/runner` imports `internal/`, and only to delegate execution.
- The root imports only `pkg/*`.
- No package in `pkg/` imports the root. This avoids import cycles.

## Alternatives considered

- **Everything at the root (minimal official layout):** the boundary between declaring and generating would rest only on identifier capitalization, with no help from the compiler.
- **Core in `internal/src` with a root facade and public dialects in `pkg/`:** this was the previous version of this ADR. Every exposed symbol
  would be declared twice, and the declarative API, which is the product itself, would be hidden behind aliases. Discarded.
- **Public builder:** would let third parties build their own dialects. Discarded because the set of providers is closed ([ADR-0007](0007-providers-abstract-factory.md)).

## Consequences

- Client import: `import "github.com/An-Ordinary-Software-Engineering/morph"`. The `pkg/` subpackages can also be imported directly.
- Script generation can change freely, since only `pkg/` and the root carry a compatibility commitment.
- The root must be kept in sync with `pkg/`. A test should check that every exported symbol in `pkg/` has an alias at the root.
- The current code still generates SQL inside `pkg/constraints` (`Builder strings.Builder`, `Build()`). This is temporary, only meant to validate
  the construction, and must leave `pkg/` ([ADR-0003](0003-declarations-and-generation.md)).
