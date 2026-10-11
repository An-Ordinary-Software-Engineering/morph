# ADR-0005: Versionamento e estabilidade da API

- **Status:** Proposto
- **Data:** 2026-10-11

## Contexto

Módulos Go públicos são resolvidos por tags git e seguem SemVer de forma estrita. A partir da `v2`, o import path muda (`/v2`).
Uma mudança incompatível publicada por engano numa versão `v1.x` quebra todos os usuários.

## Decisão

1. Usaremos **SemVer com tags `vX.Y.Z`** na raiz do repositório.
2. O projeto fica em **`v0.x` enquanto as ADRs 0002–0004 não estiverem implementadas e estáveis**. Nessa fase, mudanças incompatíveis sobem o `minor`.
3. A **`v1.0.0`** só sai quando houver pelo menos dois dialetos e `Create`, `Alter` e `Delete` estiverem completos.
4. Mudanças relevantes ficam registradas em `CHANGELOG.md`, no formato [Keep a Changelog](https://keepachangelog.com/pt-BR/).
5. A superfície pública fica mínima: só exportamos o que tem caso de uso. Na dúvida, o código vai para `internal/`.
6. A versão de `go` declarada no `go.mod` é a mais antiga ainda suportada pelo time do Go, ou seja, as duas últimas *major releases*.

## Consequências

- Os usuários podem fixar versões com segurança, e `go get -u` não quebra nada a partir da `v1`.
- Cada release precisa de uma tag e de uma entrada no CHANGELOG.
- O CI deve rodar no mínimo nas duas versões de Go suportadas.
