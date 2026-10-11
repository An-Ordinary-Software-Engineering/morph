# morph architecture

> Versão em português: [ARCHITECTURE.pt_BR.md](ARCHITECTURE.pt_BR.md)

This document describes the **target** architecture of morph as a public Go library.
The decisions that led to it are recorded in [`docs/adr/en-US/`](adr/en-US/).

## Overview

morph follows the same model as Fluent Migrator and separates three responsibilities:

- **Declare:** what the client wants to do to the schema. It is public and lives in `pkg/`.
- **Generate:** how that declaration becomes SQL for each database. It is internal and lives in `internal/src/builder` and `internal/src/provider`.
- **Execute:** when the SQL is applied. It is also internal, lives in `internal/src/runner` and is triggered by the `mph` CLI.

```
 client project                      morph (public)                  morph (internal)
┌──────────────────────┐      ┌──────────────────────────┐      ┌──────────────────────────────┐
│ migrations/          │      │ pkg/migrator             │      │ internal/src/runner          │
│  Up() / Down()       │─────▶│ pkg/constraints          │      │  ordering, version table     │
│                      │      │ pkg/types                │      │            │                 │
└──────────────────────┘      │  (neutral declarations)  │─────▶│            ▼                 │
           ▲                  └──────────────────────────┘      │ internal/src/builder         │
           │ analyzes, generates           ▲                    │  script_builder              │
┌──────────────────────┐                   │                    │            │ uses            │
│ mph up (cmd/mph)     │──▶ .morph/main.go ─┴─ morph.Main ─────▶│            ▼                 │
└──────────────────────┘     (go run)        (pkg/runner)       │ internal/src/provider        │
                                                                │  Factory: mysql, postgres... │──▶ SQL ──▶ database
                                                                └──────────────────────────────┘
```

### Layers

| Layer | Package | Responsibility | Knows SQL? |
| ----- | ------- | -------------- | ---------- |
| **Alias** | `morph` (root) | Re-exports `pkg/*` so clients import a single package. | No |
| **Declaration** | `pkg/migrator`, `pkg/constraints`, `pkg/types` | Incrementally accumulates what the client declares, with typed fields. | No |
| **Runner entry** | `pkg/runner` | `Main(args, migrations...)`, called by the generated program. Only delegates. | No |
| **Builder** | `internal/src/builder` | The `script_builder` walks the declarations and assembles the DDL. | Through the provider |
| **Providers** | `internal/src/provider/*` | Per-database Abstract Factory: types, modifiers, defaults, statements and driver. | Yes |
| **Runner** | `internal/src/runner` | Validates, orders by version, checks the version table and executes. | No (delegates) |
| **CLI** | `cmd/mph` | Reads `morph.yaml`, discovers migrations, generates and runs the program. | No |

### Dependency rules

```
 morph (root) ──▶ pkg/*
                   │
      pkg/runner ──┼──▶ internal/src/runner ──▶ internal/src/builder ──▶ internal/src/provider
                   │                                   │
                   ◀───────────────────────────────────┘  (reads the declarations)
 cmd/mph ──▶ internal/src/provider (registry and drivers) + code generation
```

- **`internal/src/*` may import `pkg/*`.** This is how the builder reads the declarations.
- **`pkg/*` never exposes `internal/` types in exported signatures.** Only `pkg/runner` imports `internal/`, and only to delegate.
- **The root imports only `pkg/*`, and no package in `pkg/` imports the root.** This avoids cycles.
- **Nothing in `pkg/` produces SQL.** SQL text comes exclusively from the providers.
- **The builder depends only on the `Factory` interface**, never on a concrete provider.

## Folder layout (target)

```
morph/
├── go.mod                  module github.com/An-Ordinary-Software-Engineering/morph
├── morph.go                single alias: re-exports pkg/* (types, functions, constants) and Main
├── doc.go                  // Package morph ... (package documentation on pkg.go.dev)
├── example_test.go         testable examples (shown on pkg.go.dev)
│
├── pkg/                    public, database-neutral declarative API
│   ├── migrator/           Migrator (CreateTable, DropTable, ...) and Migration interface
│   ├── constraints/
│   │   ├── table/          table declaration (incremental accumulation of columns and options)
│   │   └── columns/        column declaration (typed fields)
│   ├── types/              neutral column types and defaults (Integer, VarChar, CurrentTimestamp)
│   └── runner/             Main(args, migrations...): entry point for the generated program
│
├── internal/
│   └── src/
│       ├── builder/        script_builder: declarations → DDL, using the Factory
│       ├── provider/       Factory interface and registry
│       │   ├── mysql/
│       │   ├── postgres/   (future)
│       │   └── sqlite/     (future)
│       └── runner/         ordering, version table, execution
│
├── cmd/
│   └── mph/                CLI: new, up, down, status, script, build
│
├── docs/
│   ├── ARCHITECTURE.en_US.md / ARCHITECTURE.pt_BR.md
│   └── adr/                README.md, template.md (en), en-US/, pt-BR/
│
├── .github/workflows/      CI: go vet, golangci-lint, go test -race
├── README.md
├── CHANGELOG.md
└── LICENSE
```

Why this layout? See [ADR-0002](adr/en-US/0002-go-library-layout.md). In short:

- **`pkg/` is what the client uses.** It holds the declarations, and only the declarations.
- **`internal/src/` is what the CLI runs.** The compiler blocks imports from outside the module, so script generation can change without breaking anyone.
- **The root is a single alias.** Clients import `morph` and get the whole declarative API.

## Client project and the `mph` flow

Details in [ADR-0006](adr/en-US/0006-mph-cli-and-generated-program.md).

```
client-project/
├── go.mod                  require .../morph vX.Y.Z (and the database driver)
├── morph.yaml              provider: mysql · dir: ./migrations · dsn: ${DATABASE_URL}
└── migrations/             package migrations
    └── 20261011120000_create_user.go
```

1. `mph up` reads `morph.yaml`.
2. It analyzes `./migrations` with `go/packages` and finds the types that implement `Migration`.
3. It generates `.morph/main.go`, which imports the client package and the provider's driver, and calls `morph.Main(os.Args[1:], ...)`.
4. It runs `go run ./.morph up` in the client's module. The morph version used is the one declared in the client's `go.mod`.
5. The runner validates the declarations, generates the script with the builder and the selected `Factory`, and runs it.

Without the Go toolchain in the target environment, use `mph build`, which produces a standalone binary.

## Providers (Abstract Factory)

Details in [ADR-0007](adr/en-US/0007-providers-abstract-factory.md).

```go
type Factory interface {
    Name() string
    Types() TypeMapper            // types.Integer → "INT"
    Columns() ColumnWriter        // PrimaryKey, Identity, NotNull, Unique
    Defaults() DefaultMapper      // CurrentTimestamp → "CURRENT_TIMESTAMP"
    Statements() StatementWriter  // CREATE/DROP/ALTER, table options
    Driver() Driver               // database/sql driver and import path
}
```

The set of providers is closed and internal. `morph.yaml` selects the provider by name, and an internal registry resolves the `Factory`.

## Current state → target

The directory layout already matches the target. What is still missing:

| Where | Today | Target | Reference |
| ----- | ----- | ------ | --------- |
| `pkg/types` | `Integer.Get()` returns `"INT"`, and `Defaults.CurrentTimestamp` is `"CURRENT_TIMESTAMP()"` | Neutral types and defaults, translated by `TypeMapper` and `DefaultMapper` | [ADR-0003](adr/en-US/0003-declarations-and-generation.md) |
| `pkg/types` | `VarChar.Length uint8` caps at 255 | A wider integer type, validated by the provider | |
| `pkg/constraints/columns` | `Constraints []string` holding SQL | Typed fields (`PrimaryKey`, `Identity`, `Nullable`, `Unique`, `Default`) | [ADR-0003](adr/en-US/0003-declarations-and-generation.md) |
| `pkg/constraints/table` | `Builder strings.Builder` and `Build()`, temporary, to validate the query | Only accumulates the declaration. Assembly moves to `internal/src/builder` | [ADR-0003](adr/en-US/0003-declarations-and-generation.md) |
| `pkg/constraints/table` | A modifier before `Column` panics (`payload[-1]`) | Accumulated errors, no panic | [ADR-0004](adr/en-US/0004-fluent-api-and-errors.md) |
| `pkg/migrator` | `CreateTable` writes `DROP`/`CREATE` directly | Only records the declaration | [ADR-0003](adr/en-US/0003-declarations-and-generation.md) |
| `internal/src/builder` | `script_builder.go` is empty | DDL assembly through the `Factory` | [ADR-0007](adr/en-US/0007-providers-abstract-factory.md) |
| root | Does not exist | `morph.go` re-exporting `pkg/*` | [ADR-0002](adr/en-US/0002-go-library-layout.md) |
| `cmd/main.go` | Test harness | `cmd/mph` (CLI). The example moves to `example_test.go` | [ADR-0006](adr/en-US/0006-mph-cli-and-generated-program.md) |

## Quality and publishing

- **Versioning:** SemVer with `vX.Y.Z` tags. The project stays at `v0.x` until the API stabilizes ([ADR-0005](adr/en-US/0005-versioning-and-stability.md)).
- **Documentation:** every exported identifier has a doc comment. Examples are written as testable `Example*` functions.
- **Tests:** per-provider golden files (`testdata/<provider>/*.sql`) and a test that ensures the root re-exports everything `pkg/` exports.
- **CI:** `go vet`, `golangci-lint`, `go test -race ./...` and `govulncheck`.
- **Dependencies:** standard library only in `pkg/`. Database drivers enter only in the generated program, and only the selected provider's driver.
