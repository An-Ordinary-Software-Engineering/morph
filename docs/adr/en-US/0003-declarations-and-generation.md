# ADR-0003: Separate declarations from script generation

- **Status:** Proposed
- **Date:** 2026-10-11

## Context

Today `pkg/constraints/table` accumulates columns incrementally, which is already the desired behavior.
The problem is that this accumulation stores **MySQL SQL strings** (`"PRIMARY KEY"`, `"AUTO_INCREMENT"`, `"DEFAULT CURRENT_TIMESTAMP()"`),
and `types.Integer.Get()` returns `"INT"`. This creates three problems:

- Supporting another database would require changing the public API.
- A migration cannot be inspected or validated before SQL is generated. For example, to detect two `PRIMARY KEY`s.
- `Down` cannot be generated automatically from `Up`, as Fluent Migrator does with `AutoReversingMigration`.

Fluent Migrator solves this with a pipeline in which the fluent API produces **expressions**, and each database **generator** translates those expressions into SQL.

## Decision

We will use a two-stage pipeline:

1. **`pkg/` produces neutral declarations.** `pkg/constraints` accumulates structs with typed fields
   (`PrimaryKey bool`, `Identity bool`, `Nullable bool`, `Default`), and `pkg/types` describes types (`Integer`, `VarChar{Length}`)
   and default values (`CurrentTimestamp`) **with no SQL text at all**.
2. **`internal/src/builder` generates the script.** The `script_builder` walks the declarations and assembles the DDL incrementally.
   The parts that vary per database (type names, modifiers, defaults and the shape of each statement) come from the selected provider's factory ([ADR-0007](0007-providers-abstract-factory.md)).

Database-specific options (for example, MySQL's `ENGINE`) go into the declaration as a namespaced options map (`"mysql.engine"`),
and only the matching provider interprets them.

## Alternatives considered

- **Keep SQL strings in the declarations:** ties the public API to a single database.
- **A single builder with a `switch provider` inside:** scales poorly and couples everything.
- **`text/template` templates per database:** works for simple cases, but is hard to test and to handle special cases with.

## Consequences

- `CreateTableConstraint`'s `Build()` and `strings.Builder` leave `pkg/`. Today they exist only to validate query construction.
- Declarations are part of the public API and need care regarding compatibility ([ADR-0005](0005-versioning-and-stability.md)).
- Tests can run at two levels: the API produces the expected declaration, and the builder, with each provider, produces the expected SQL (golden files).
