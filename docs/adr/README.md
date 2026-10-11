# Architecture Decision Records

Architectural decisions of morph, in Michael Nygard's format
(Context → Decision → Consequences).

Each ADR exists in two languages with the same number:

- [`en-US/`](en-US/) — English
- [`pt-BR/`](pt-BR/) — Português

## How to use

1. Copy [`template.md`](template.md) to `en-US/NNNN-short-title.md` and to `pt-BR/NNNN-titulo-curto.md`, using the next free number.
2. Open it with status **Proposed** and change it to **Accepted** once there is consensus.
3. An accepted ADR is not edited. To change a decision, write a new ADR that **supersedes** it and mark the old one as **Superseded by ADR-NNNN**.
4. Keep both language versions in sync in the same change.

## Index

| #    | Title (en-US)                                                                         | Título (pt-BR)                                                                   | Status   |
| ---- | ------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- | -------- |
| 0001 | [Record architecture decisions](en-US/0001-record-architecture-decisions.md)          | [Registrar decisões arquiteturais](pt-BR/0001-registrar-decisoes.md)             | Accepted |
| 0002 | [Go library layout](en-US/0002-go-library-layout.md)                                  | [Layout de biblioteca Go](pt-BR/0002-layout-de-biblioteca-go.md)                 | Proposed |
| 0003 | [Separate declarations from script generation](en-US/0003-declarations-and-generation.md) | [Separar declarações da geração do script](pt-BR/0003-declaracoes-e-geracao.md) | Proposed |
| 0004 | [Fluent API and error handling](en-US/0004-fluent-api-and-errors.md)                  | [API fluente e tratamento de erros](pt-BR/0004-api-fluente-e-erros.md)           | Proposed |
| 0005 | [Versioning and API stability](en-US/0005-versioning-and-stability.md)                | [Versionamento e estabilidade da API](pt-BR/0005-versionamento-e-estabilidade.md) | Proposed |
| 0006 | [`mph` CLI with a generated program](en-US/0006-mph-cli-and-generated-program.md)     | [CLI `mph` com programa gerado](pt-BR/0006-cli-mph-e-programa-gerado.md)          | Proposed |
| 0007 | [Providers as an Abstract Factory](en-US/0007-providers-abstract-factory.md)          | [Providers como Abstract Factory](pt-BR/0007-providers-abstract-factory.md)       | Proposed |
