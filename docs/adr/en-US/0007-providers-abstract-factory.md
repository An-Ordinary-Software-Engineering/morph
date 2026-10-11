# ADR-0007: Providers as an Abstract Factory

- **Status:** Proposed
- **Date:** 2026-10-11

## Context

The same set of declarations ([ADR-0003](0003-declarations-and-generation.md)) must become SQL for different databases (MySQL, PostgreSQL, ...).
The script assembly algorithm is the same for all of them: walk tables and columns, build the definitions and join the statements.
What changes is a **family of details that must be consistent with one another**: type names, modifiers, default values,
the shape of each statement and the driver used to execute it. Mixing, for example, PostgreSQL's type mapping with MySQL's `AUTO_INCREMENT` would produce invalid SQL.

## Decision

Each database is a **provider**, implemented as an **Abstract Factory** that delivers that complete family:

```go
// internal/src/provider
type Factory interface {
    Name() string                 // "mysql", "postgres", ...
    Types() TypeMapper            // types.Integer → "INT" | "INTEGER"
    Columns() ColumnWriter        // PrimaryKey, Identity, NotNull, Unique → SQL fragments
    Defaults() DefaultMapper      // CurrentTimestamp → "CURRENT_TIMESTAMP" | "now()"
    Statements() StatementWriter  // CREATE/DROP/ALTER, table options, modifier order
    Driver() Driver               // database/sql name and import path, used by the CLI in the generated program
}
```

- Each provider lives in `internal/src/provider/<name>` (`mysql`, `postgres`, `sqlite`, ...).
- An internal **registry** maps the name to the factory (`"mysql"` → `mysql.Factory{}`). `morph.yaml` selects the provider by name ([ADR-0006](0006-mph-cli-and-generated-program.md)).
- The `script_builder` (`internal/src/builder`) depends only on the `Factory` interface, never on a concrete provider.
- Namespaced options (`"mysql.engine"`) are interpreted only by the provider that owns the namespace. The others ignore them.
- **The set of providers is closed:** everything lives in `internal/`, and third parties cannot register their own providers.

## Alternatives considered

- **A single `Dialect.Generate(expr) string` interface per database:** each dialect would reimplement the whole algorithm, duplicating the assembly logic.
- **A Strategy per product, with no factory**, injecting `TypeMapper`, `ColumnWriter`, etc. separately: allows inconsistent combinations across databases.
- **Public, extensible providers:** enlarge the compatibility surface ([ADR-0005](0005-versioning-and-stability.md)) with no concrete demand.
  If that demand appears, a new ADR can expose the `Factory` interface in `pkg/`.

## Consequences

- Adding a database means implementing a new `Factory` and registering it, without touching the builder or the public API.
- When a database diverges from the common flow (for example, SQLite's `ALTER TABLE` limitations), its `StatementWriter` can rewrite the whole statement.
- Golden-file tests run the same set of declarations against each `Factory` (`testdata/<provider>/*.sql`).
- Factory products are small and stateless, so each provider can be a plain value (`mysql.Factory{}`) with no initialization cost.
