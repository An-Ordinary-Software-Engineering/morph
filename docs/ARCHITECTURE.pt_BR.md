# Arquitetura do morph

> English version: [ARCHITECTURE.en_US.md](ARCHITECTURE.en_US.md)

Este documento descreve a arquitetura **alvo** do morph como biblioteca Go pública.
As decisões que levaram a ela estão registradas em [`docs/adr/pt-BR/`](adr/pt-BR/).

## Visão geral

O morph segue o mesmo modelo do Fluent Migrator e separa três responsabilidades:

- **Declarar:** o que o cliente quer fazer com o esquema. É público e fica em `pkg/`.
- **Gerar:** como essa declaração vira SQL em cada banco. É interno e fica em `internal/src/builder` e `internal/src/provider`.
- **Executar:** quando o SQL é aplicado. Também é interno, fica em `internal/src/runner` e é acionado pela CLI `mph`.

```
 projeto do cliente                  morph (público)                 morph (interno)
┌──────────────────────┐      ┌──────────────────────────┐      ┌──────────────────────────────┐
│ migrations/          │      │ pkg/migrator             │      │ internal/src/runner          │
│  Up() / Down()       │─────▶│ pkg/constraints          │      │  ordena, tabela de versões   │
│                      │      │ pkg/types                │      │            │                 │
└──────────────────────┘      │  (declarações neutras)   │─────▶│            ▼                 │
           ▲                  └──────────────────────────┘      │ internal/src/builder         │
           │ analisa e gera                ▲                    │  script_builder              │
┌──────────────────────┐                   │                    │            │ usa             │
│ mph up (cmd/mph)     │──▶ .morph/main.go ─┴─ morph.Main ─────▶│            ▼                 │
└──────────────────────┘     (go run)        (pkg/runner)       │ internal/src/provider        │
                                                                │  Factory: mysql, postgres... │──▶ SQL ──▶ banco
                                                                └──────────────────────────────┘
```

### Camadas

| Camada | Pacote | Responsabilidade | Conhece SQL? |
| ------ | ------ | ---------------- | ------------ |
| **Alias** | `morph` (raiz) | Reexporta `pkg/*` para que o cliente importe um único pacote. | Não |
| **Declaração** | `pkg/migrator`, `pkg/constraints`, `pkg/types` | Acumula de forma incremental o que o cliente declara, com campos tipados. | Não |
| **Entrada do runner** | `pkg/runner` | `Main(args, migrations...)`, chamado pelo programa gerado. Só delega. | Não |
| **Builder** | `internal/src/builder` | O `script_builder` percorre as declarações e monta o DDL. | Via provider |
| **Providers** | `internal/src/provider/*` | Abstract Factory por banco: tipos, modificadores, defaults, instruções e driver. | Sim |
| **Runner** | `internal/src/runner` | Valida, ordena por versão, consulta a tabela de versões e executa. | Não (delega) |
| **CLI** | `cmd/mph` | Lê o `morph.yaml`, descobre as migrations, gera e roda o programa. | Não |

### Regras de dependência

```
 morph (raiz) ──▶ pkg/*
                   │
      pkg/runner ──┼──▶ internal/src/runner ──▶ internal/src/builder ──▶ internal/src/provider
                   │                                   │
                   ◀───────────────────────────────────┘  (lê as declarações)
 cmd/mph ──▶ internal/src/provider (registro e drivers) + geração de código
```

- **`internal/src/*` pode importar `pkg/*`.** É assim que o builder lê as declarações.
- **`pkg/*` nunca expõe tipos de `internal/` em assinaturas exportadas.** Só o `pkg/runner` importa `internal/`, e apenas para delegar.
- **A raiz importa só `pkg/*`, e nenhum pacote de `pkg/` importa a raiz.** Isso evita ciclos.
- **Nada em `pkg/` produz SQL.** Os textos SQL vêm exclusivamente dos providers.
- **O builder depende só da interface `Factory`**, nunca de um provider concreto.

## Layout de pastas (alvo)

```
morph/
├── go.mod                  module github.com/An-Ordinary-Software-Engineering/morph
├── morph.go                alias único: reexporta pkg/* (tipos, funções, constantes) e Main
├── doc.go                  // Package morph ... (documentação do pacote no pkg.go.dev)
├── example_test.go         exemplos testáveis (aparecem no pkg.go.dev)
│
├── pkg/                    API declarativa pública e neutra em relação ao banco
│   ├── migrator/           Migrator (CreateTable, DropTable, ...) e interface Migration
│   ├── constraints/
│   │   ├── table/          declaração de tabela (acúmulo incremental de colunas e opções)
│   │   └── columns/        declaração de coluna (campos tipados)
│   ├── types/              tipos de coluna e defaults neutros (Integer, VarChar, CurrentTimestamp)
│   └── runner/             Main(args, migrations...): entrada para o programa gerado
│
├── internal/
│   └── src/
│       ├── builder/        script_builder: declarações → DDL, usando a Factory
│       ├── provider/       interface Factory e registro
│       │   ├── mysql/
│       │   ├── postgres/   (futuro)
│       │   └── sqlite/     (futuro)
│       └── runner/         ordenação, tabela de versões, execução
│
├── cmd/
│   └── mph/                CLI: new, up, down, status, script, build
│
├── docs/
│   ├── ARCHITECTURE.pt_BR.md / ARCHITECTURE.en_US.md
│   └── adr/                README.md, template.md (en), pt-BR/, en-US/
│
├── .github/workflows/      CI: go vet, golangci-lint, go test -race
├── README.md
├── CHANGELOG.md
└── LICENSE
```

Por que esse layout? Veja a [ADR-0002](adr/pt-BR/0002-layout-de-biblioteca-go.md). Em resumo:

- **`pkg/` é o que o cliente usa.** Ali ficam as declarações, e só as declarações.
- **`internal/src/` é o que a CLI executa.** O compilador impede o import de fora do módulo, então a geração do script pode mudar sem quebrar ninguém.
- **A raiz é um alias único.** O cliente importa `morph` e tem acesso a toda a API declarativa.

## Projeto do cliente e fluxo do `mph`

Detalhes na [ADR-0006](adr/pt-BR/0006-cli-mph-e-programa-gerado.md).

```
projeto-do-cliente/
├── go.mod                  require .../morph vX.Y.Z (e o driver do banco)
├── morph.yaml              provider: mysql · dir: ./migrations · dsn: ${DATABASE_URL}
└── migrations/             package migrations
    └── 20261011120000_create_user.go
```

1. `mph up` lê o `morph.yaml`.
2. Analisa `./migrations` com `go/packages` e encontra os tipos que implementam `Migration`.
3. Gera `.morph/main.go`, que importa o pacote do cliente e o driver do provider, e chama `morph.Main(os.Args[1:], ...)`.
4. Executa `go run ./.morph up` no módulo do cliente. A versão do morph usada é a declarada no `go.mod` do cliente.
5. O runner valida as declarações, gera o script com o builder e a `Factory` escolhida e o executa.

Sem o toolchain Go no ambiente de destino, use `mph build`, que gera um binário autônomo.

## Providers (Abstract Factory)

Detalhes na [ADR-0007](adr/pt-BR/0007-providers-abstract-factory.md).

```go
type Factory interface {
    Name() string
    Types() TypeMapper            // types.Integer → "INT"
    Columns() ColumnWriter        // PrimaryKey, Identity, NotNull, Unique
    Defaults() DefaultMapper      // CurrentTimestamp → "CURRENT_TIMESTAMP"
    Statements() StatementWriter  // CREATE/DROP/ALTER, opções de tabela
    Driver() Driver               // driver database/sql e import path
}
```

O conjunto de providers é fechado e interno. O `morph.yaml` escolhe o provider pelo nome, e um registro interno resolve a `Factory`.

## Estado atual → alvo

O layout de diretórios já segue o alvo. O que ainda falta:

| Onde | Hoje | Alvo | Referência |
| ---- | ---- | ---- | ---------- |
| `pkg/types` | `Integer.Get()` retorna `"INT"`, e `Defaults.CurrentTimestamp` é `"CURRENT_TIMESTAMP()"` | Tipos e defaults neutros, que o `TypeMapper` e o `DefaultMapper` traduzem | [ADR-0003](adr/pt-BR/0003-declaracoes-e-geracao.md) |
| `pkg/types` | `VarChar.Length uint8` limita a 255 | Tipo inteiro maior, validado pelo provider | |
| `pkg/constraints/columns` | `Constraints []string` com SQL | Campos tipados (`PrimaryKey`, `Identity`, `Nullable`, `Unique`, `Default`) | [ADR-0003](adr/pt-BR/0003-declaracoes-e-geracao.md) |
| `pkg/constraints/table` | `Builder strings.Builder` e `Build()`, provisórios para validar a query | Só acumula a declaração. A montagem vai para `internal/src/builder` | [ADR-0003](adr/pt-BR/0003-declaracoes-e-geracao.md) |
| `pkg/constraints/table` | Um modificador antes de `Column` causa panic (`payload[-1]`) | Erros acumulados, sem panic | [ADR-0004](adr/pt-BR/0004-api-fluente-e-erros.md) |
| `pkg/migrator` | `CreateTable` escreve `DROP`/`CREATE` direto | Só registra a declaração | [ADR-0003](adr/pt-BR/0003-declaracoes-e-geracao.md) |
| `internal/src/builder` | `script_builder.go` vazio | Montagem do DDL via `Factory` | [ADR-0007](adr/pt-BR/0007-providers-abstract-factory.md) |
| raiz | Não existe | `morph.go` reexportando `pkg/*` | [ADR-0002](adr/pt-BR/0002-layout-de-biblioteca-go.md) |
| `cmd/main.go` | Harness de teste | `cmd/mph` (CLI). O exemplo vai para `example_test.go` | [ADR-0006](adr/pt-BR/0006-cli-mph-e-programa-gerado.md) |

## Qualidade e publicação

- **Versionamento:** SemVer com tags `vX.Y.Z`. O projeto fica em `v0.x` até a API estabilizar ([ADR-0005](adr/pt-BR/0005-versionamento-e-estabilidade.md)).
- **Documentação:** todo identificador exportado tem doc comment. Os exemplos são escritos como `Example*` testáveis.
- **Testes:** golden files por provider (`testdata/<provider>/*.sql`) e um teste que garante que a raiz reexporta tudo o que `pkg/` exporta.
- **CI:** `go vet`, `golangci-lint`, `go test -race ./...` e `govulncheck`.
- **Dependências:** apenas a biblioteca padrão em `pkg/`. Drivers de banco entram só no programa gerado, e apenas o driver do provider escolhido.
