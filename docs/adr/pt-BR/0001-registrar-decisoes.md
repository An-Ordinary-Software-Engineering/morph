# ADR-0001: Registrar decisões arquiteturais

- **Status:** Aceito
- **Data:** 2026-10-11

## Contexto

O morph vai ser publicado como pacote reutilizável. Quem chegar depois, sejam contribuidores ou o próprio autor
daqui a alguns meses, precisa entender por que a estrutura e a API são como são, e não só o que elas são.

## Decisão

Registraremos decisões arquiteturais relevantes como ADRs em `docs/adr/`, no formato de Michael Nygard, em português (`docs/adr/pt-BR/`) e inglês (`docs/adr/en-US/`).
O índice e o template ficam na raiz de `docs/adr/`, somente em inglês.
A visão consolidada e atual da arquitetura fica em `docs/ARCHITECTURE.pt_BR.md` e `docs/ARCHITECTURE.en_US.md`.

## Consequências

- Uma decisão que afeta a estrutura, a API pública ou dependências exige uma ADR.
- ADRs aceitas são imutáveis. Mudanças entram como novas ADRs.
- As duas versões de idioma (ADR e `ARCHITECTURE`) são atualizadas juntas.
