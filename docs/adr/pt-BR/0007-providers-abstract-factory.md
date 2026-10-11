# ADR-0007: Providers como Abstract Factory

- **Status:** Proposto
- **Data:** 2026-10-11

## Contexto

O mesmo conjunto de declarações ([ADR-0003](0003-declaracoes-e-geracao.md)) precisa virar SQL em bancos diferentes (MySQL, PostgreSQL, ...).
O algoritmo de montagem do script é o mesmo para todos: percorrer as tabelas e colunas, montar as definições e juntar as instruções.
O que muda é uma **família de detalhes que precisam ser coerentes entre si**: nomes de tipo, modificadores, valores padrão,
a forma de cada instrução e o driver usado na execução. Misturar, por exemplo, o mapeamento de tipos do PostgreSQL com o `AUTO_INCREMENT` do MySQL produziria SQL inválido.

## Decisão

Cada banco é um **provider**, implementado como uma **Abstract Factory** que entrega essa família completa:

```go
// internal/src/provider
type Factory interface {
    Name() string                 // "mysql", "postgres", ...
    Types() TypeMapper            // types.Integer → "INT" | "INTEGER"
    Columns() ColumnWriter        // PrimaryKey, Identity, NotNull, Unique → trechos de SQL
    Defaults() DefaultMapper      // CurrentTimestamp → "CURRENT_TIMESTAMP" | "now()"
    Statements() StatementWriter  // CREATE/DROP/ALTER, opções de tabela, ordem dos modificadores
    Driver() Driver               // nome no database/sql e import path, usado pela CLI no programa gerado
}
```

- Cada provider fica em `internal/src/provider/<nome>` (`mysql`, `postgres`, `sqlite`, ...).
- Um **registro** interno mapeia o nome para a fábrica (`"mysql"` → `mysql.Factory{}`). O `morph.yaml` escolhe o provider pelo nome ([ADR-0006](0006-cli-mph-e-programa-gerado.md)).
- O `script_builder` (`internal/src/builder`) depende só da interface `Factory`, e nunca de um provider concreto.
- Opções com namespace (`"mysql.engine"`) só são interpretadas pelo provider dono do namespace. Os demais as ignoram.
- **O conjunto de providers é fechado:** tudo fica em `internal/`, e terceiros não podem registrar providers próprios.

## Alternativas consideradas

- **Interface única `Dialect.Generate(expr) string` por banco:** cada dialeto reimplementaria o algoritmo inteiro, duplicando a lógica de montagem.
- **Strategy por produto, sem fábrica**, injetando `TypeMapper`, `ColumnWriter` etc. separadamente: permite combinações incoerentes entre bancos.
- **Providers públicos e extensíveis:** aumentam a superfície de compatibilidade ([ADR-0005](0005-versionamento-e-estabilidade.md)) sem uma demanda concreta.
  Se essa demanda surgir, uma nova ADR pode expor a interface `Factory` em `pkg/`.

## Consequências

- Adicionar um banco significa implementar uma nova `Factory` e registrá-la, sem tocar no builder nem na API pública.
- Quando um banco divergir do fluxo comum (por exemplo, as limitações de `ALTER TABLE` do SQLite), o `StatementWriter` dele pode reescrever a instrução inteira.
- Os testes de golden file rodam o mesmo conjunto de declarações contra cada `Factory` (`testdata/<provider>/*.sql`).
- Os produtos da fábrica são pequenos e sem estado, então cada provider pode ser um valor (`mysql.Factory{}`) sem custo de inicialização.
