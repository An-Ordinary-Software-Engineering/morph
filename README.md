# morph

Gerador de migrations SQL para Go, com uma API fluente inspirada na semântica do [Fluent Migrator](https://fluentmigrator.github.io/) (.NET).

Em vez de escrever SQL à mão, você descreve o esquema encadeando chamadas de método, e o morph gera o script SQL correspondente.

> **Status:** em desenvolvimento inicial. A API ainda pode mudar.

## Requisitos

- Go 1.24+

## Instalação

```bash
go get github.com/An-Ordinary-Software-Engineering/morph
```

## Uso

```go
package main

import (
	"fmt"

	"github.com/An-Ordinary-Software-Engineering/morph/internal/constraints/table"
	"github.com/An-Ordinary-Software-Engineering/morph/internal/types"
	"github.com/An-Ordinary-Software-Engineering/morph/src/migrator"
)

func main() {
	constraint := migrator.CreateTable(
		"user",
		func(table *table.CreateTableConstraint) *table.CreateTableConstraint {
			return table.
				Column("id", types.Integer{}).PrimaryKey().AutoIncrement().
				Column("name", types.VarChar{Length: 50}).NotNull().
				Column("nickname", types.VarChar{Length: 50}).NotNull().Unique().
				Column("email", types.VarChar{Length: 50}).NotNull().Unique().
				Column("password", types.VarChar{Length: 255}).NotNull().
				Column("created_on", types.Timestamp{}).Default(types.Defaults.CurrentTimestamp).NotNull()
		},
		table.Options{
			DropIfExists: true,
			Engine:       "INNODB",
		},
	)

	fmt.Println(constraint.Builder.String())
}
```

Saída:

```sql
DROP TABLE IF EXISTS user;

CREATE TABLE user (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    nickname VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_on TIMESTAMP DEFAULT CURRENT_TIMESTAMP() NOT NULL
) ENGINE=INNODB;
```

Para rodar o exemplo:

```bash
go run ./cmd
```

## API

### `migrator.CreateTable(name, blueprint, options)`

Gera um `CREATE TABLE`. O `blueprint` recebe a tabela e define suas colunas de forma encadeada.

### Colunas

Cada `Column(name, type)` inicia uma nova coluna; os modificadores seguintes se aplicam a ela:

| Método            | SQL gerado       |
| ----------------- | ---------------- |
| `PrimaryKey()`    | `PRIMARY KEY`    |
| `AutoIncrement()` | `AUTO_INCREMENT` |
| `NotNull()`       | `NOT NULL`       |
| `Unique()`        | `UNIQUE`         |
| `Default(value)`  | `DEFAULT value`  |

### Tipos

| Tipo                          | SQL gerado    |
| ----------------------------- | ------------- |
| `types.Integer{}`             | `INT`         |
| `types.VarChar{Length: n}`    | `VARCHAR(n)`  |
| `types.Char{Length: n}`       | `CHAR(n)`     |
| `types.Timestamp{}`           | `TIMESTAMP`   |

Valores padrão prontos: `types.Defaults.CurrentTimestamp` (`CURRENT_TIMESTAMP()`).

### Opções da tabela

| Campo          | Efeito                                           |
| -------------- | ------------------------------------------------ |
| `DropIfExists` | Adiciona `DROP TABLE IF EXISTS` antes da criação |
| `Engine`       | Adiciona `ENGINE=<valor>` ao final               |

## Estrutura do projeto

```
cmd/          exemplo executável
src/migrator/ ponto de entrada da API (CreateTable, ...)
internal/     constraints de tabela/coluna e tipos SQL
pkg/builder/  construção de scripts
```

## Roadmap

- [ ] `AlterTable`, `DropTable`, `Rename`
- [ ] Chaves estrangeiras e índices
- [ ] Migrations versionadas com `Up` / `Down`
- [ ] Suporte a múltiplos dialetos (MySQL, PostgreSQL, SQLite)
- [ ] CLI para gerar e aplicar migrations
