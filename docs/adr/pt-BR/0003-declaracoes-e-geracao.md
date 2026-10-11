# ADR-0003: Separar declarações da geração do script

- **Status:** Proposto
- **Data:** 2026-10-11

## Contexto

Hoje `pkg/constraints/table` acumula as colunas incrementalmente, o que já é o comportamento desejado.
O problema é que esse acúmulo guarda **strings SQL de MySQL** (`"PRIMARY KEY"`, `"AUTO_INCREMENT"`, `"DEFAULT CURRENT_TIMESTAMP()"`),
e `types.Integer.Get()` retorna `"INT"`. Isso cria três problemas:

- Suportar outro banco exigiria mudar a API pública.
- Não dá para inspecionar ou validar a migration antes de gerar o SQL. Por exemplo, para detectar dois `PRIMARY KEY`.
- Não dá para gerar o `Down` automaticamente a partir do `Up`, como o Fluent Migrator faz com `AutoReversingMigration`.

O Fluent Migrator resolve isso com um pipeline em que a API fluente gera **expressões**, e cada **generator** de banco traduz essas expressões em SQL.

## Decisão

Usaremos um pipeline em duas etapas:

1. **`pkg/` produz declarações neutras.** `pkg/constraints` acumula structs com campos tipados
   (`PrimaryKey bool`, `Identity bool`, `Nullable bool`, `Default`), e `pkg/types` descreve tipos (`Integer`, `VarChar{Length}`)
   e valores padrão (`CurrentTimestamp`) **sem nenhum texto SQL**.
2. **`internal/src/builder` gera o script.** O `script_builder` percorre as declarações e monta o DDL incrementalmente.
   Os trechos que variam por banco (nomes de tipo, modificadores, defaults e a forma de cada instrução) vêm da fábrica do provider selecionado ([ADR-0007](0007-providers-abstract-factory.md)).

Opções específicas de um banco (por exemplo, `ENGINE` no MySQL) entram na declaração como um mapa de opções com namespace (`"mysql.engine"`),
e só o provider correspondente as interpreta.

## Alternativas consideradas

- **Manter strings SQL nas declarações:** prende a API pública a um único banco.
- **Builder único com `switch provider` dentro:** cresce mal e acopla tudo.
- **Templates `text/template` por banco:** funciona para casos simples, mas fica difícil de testar e de tratar casos especiais.

## Consequências

- O `Build()` e o `strings.Builder` de `CreateTableConstraint` saem de `pkg/`. Eles existem hoje só para validar a construção da query.
- As declarações fazem parte da API pública e precisam de cuidado com compatibilidade ([ADR-0005](0005-versionamento-e-estabilidade.md)).
- Os testes podem ser feitos em dois níveis: a API gera a declaração esperada, e o builder, com cada provider, gera o SQL esperado (golden files).
