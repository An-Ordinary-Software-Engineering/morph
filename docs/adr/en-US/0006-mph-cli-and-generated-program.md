# ADR-0006: `mph` CLI with a generated program

- **Status:** Proposed
- **Date:** 2026-10-11

## Context

Clients write migrations (`Up`/`Down`) in Go, in a dedicated directory of their project, and want to apply them with `mph up`.
Go is compiled, so an installed `mph` binary **cannot load client Go code at runtime**.
On top of that, client code lives in another module and cannot import morph's `internal/`.

## Decision

The `mph` CLI (`cmd/mph`) will **generate a temporary program and run it inside the client's module**. This is the same mechanism as `go test`,
which generates a `main` that calls `testing.MainStart`.

### Client project conventions

```
client-project/
├── go.mod                    require github.com/An-Ordinary-Software-Engineering/morph vX.Y.Z
├── morph.yaml                provider, migrations directory and DSN
└── migrations/               dedicated Go package (package migrations)
    ├── 20261011120000_create_user.go
    └── 20261012090000_add_role.go
```

```yaml
# morph.yaml
provider: mysql
dir: ./migrations
dsn: ${DATABASE_URL}
```

- A migration is an exported type that implements `morph.Migration`. The version comes from `Version()` and uses the `YYYYMMDDHHMMSS` format.
- `mph new <name>` creates the file with the version and an `Up`/`Down` skeleton already in place.

### `mph up` flow

1. Reads `morph.yaml`.
2. Analyzes the package in `dir` with `go/packages` and finds the types that implement `Migration`. No `init()` or manual registration is needed.
3. Generates `.morph/main.go` in the client's module:

   ```go
   package main

   import (
       "os"

       "github.com/An-Ordinary-Software-Engineering/morph"
       _ "github.com/go-sql-driver/mysql" // only the configured provider's driver
       "client-project/migrations"
   )

   func main() {
       morph.Main(os.Args[1:], migrations.CreateUser{}, migrations.AddRole{})
   }
   ```

4. Runs `go run ./.morph up` in the client's directory.
5. `morph.Main` (an alias of `pkg/runner.Main`) delegates to `internal/src/runner`. The runner validates the declarations, orders the migrations by version,
   checks the version table, generates the script with `internal/src/builder` and the selected provider, and runs it in a transaction when the database supports transactional DDL.

### Commands

| Command             | Effect |
| ------------------- | ------ |
| `mph new <name>`    | Creates a migration with the current version. |
| `mph up`            | Applies pending migrations. |
| `mph down [n]`      | Reverts the last `n` migrations (default: 1). |
| `mph status`        | Lists applied and pending migrations. |
| `mph script`        | Generates the SQL without running it (*dry-run*). |
| `mph build -o bin`  | Builds a standalone binary with the migrations, for environments without the Go toolchain. |

### Public entry point

`pkg/runner.Main` is the **only** place in `pkg/` that imports `internal/`. It is thin, documented as "for use by generated code"
and re-exported at the root as `morph.Main`. The builder, the providers and the runner stay internal.

## Alternatives considered

- **Public runner with a client-written `main`:** simpler, but requires boilerplate in every project and does not deliver `mph up` searching a directory.
- **Go `plugin` package:** works only on Linux/macOS, requires the same compiler and the same dependency versions, and is fragile.
- **Embedded interpreter (e.g. `yaegi`):** incomplete language support, slower execution and behavior that differs from compiled code.
- **Migrations in SQL or YAML instead of Go:** loses the fluent API, which is the point of the project.

## Consequences

- **Consistent version:** the generated program compiles against the morph version declared in the client's `go.mod`, so the CLI and the library never drift apart.
  Even so, the CLI should check that the `Main` protocol is compatible and warn when it is not.
- `mph up`, `down`, `status` and `script` require the Go toolchain. For production without Go, the path is `mph build`.
- The database driver must be in the client's `go.mod`. The CLI should detect when it is missing and suggest the matching `go get`.
- `.morph/` is a generated directory that clients should keep out of version control. `mph new` can add it to `.gitignore`.
- The first run compiles the program. Later runs reuse Go's build cache.
