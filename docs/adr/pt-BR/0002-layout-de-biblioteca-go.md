# ADR-0002: Layout de biblioteca Go

- **Status:** Proposto
- **Data:** 2026-10-11

## Contexto

O morph tem duas responsabilidades com públicos diferentes:

- **Declarar** migrations (`Up`/`Down`). O código do cliente faz isso, então essa parte precisa ser pública e estável.
- **Gerar e executar** o script DDL a partir dessas declarações. Isso é um detalhe interno, acionado pela CLI `mph` ([ADR-0006](0006-cli-mph-e-programa-gerado.md)).

A referência oficial ([go.dev/doc/modules/layout](https://go.dev/doc/modules/layout)) sugere manter o código na raiz do módulo.
Mesmo assim, queremos que a fronteira entre o que o cliente usa e o que é interno apareça na própria estrutura de pastas,
e que o compilador ajude a garantir essa fronteira.

## Decisão

1. **`pkg/` = API declarativa pública.** É o que o cliente usa para escrever migrations:
   - `pkg/migrator`: o `Migrator` (`CreateTable`, `DropTable`, ...) e a interface `Migration`;
   - `pkg/constraints`: declarações acumuladas incrementalmente (`table`, `columns`);
   - `pkg/types`: tipos de coluna e valores padrão **neutros em relação ao banco**;
   - `pkg/runner`: ponto de entrada fino (`Main`) usado pelo programa gerado pela CLI.

   Nada em `pkg/` produz SQL.
2. **`internal/src/` = geração e execução.** O compilador impede o import de fora do módulo:
   - `internal/src/builder`: o `script_builder` percorre as declarações e monta o script DDL;
   - `internal/src/provider`: fábricas por banco ([ADR-0007](0007-providers-abstract-factory.md));
   - `internal/src/runner`: ordena as migrations, controla a tabela de versões e executa o script.
3. **Raiz do módulo = pacote `morph`, um alias único.** Ele só reexporta o que está em `pkg/` (aliases de tipo, wrappers de função e constantes),
   para que o cliente importe um único pacote.
4. **`cmd/mph/` = CLI.** Fica no mesmo módulo e, por isso, pode usar `internal/`.
5. Os nomes de pacote são curtos, no singular e sem `util`, `common` ou `helpers`.

### Regras de dependência

- `internal/src/*` pode importar `pkg/*`. É assim que o builder lê as declarações.
- **`pkg/*` nunca expõe tipos de `internal/` em assinaturas exportadas.** Sem isso, o cliente recebe um tipo que não consegue nomear.
  Só o `pkg/runner` importa `internal/`, e apenas para delegar a execução.
- A raiz importa apenas `pkg/*`.
- Nenhum pacote de `pkg/` importa a raiz. Isso evita ciclos de import.

## Alternativas consideradas

- **Tudo na raiz (layout oficial mínimo):** a fronteira entre declarar e gerar ficaria só no uso de maiúsculas e minúsculas, sem ajuda do compilador.
- **Núcleo em `internal/src` com fachada na raiz e dialetos públicos em `pkg/`:** esta foi a versão anterior desta ADR. Nela, cada símbolo exposto
  precisaria ser declarado duas vezes, e a API declarativa, que é o produto em si, ficaria escondida atrás de aliases. Foi descartada.
- **Builder público:** permitiria que terceiros criassem dialetos próprios. Foi descartado porque o conjunto de providers é fechado ([ADR-0007](0007-providers-abstract-factory.md)).

## Consequências

- Import do cliente: `import "github.com/An-Ordinary-Software-Engineering/morph"`. Os subpacotes de `pkg/` também podem ser importados diretamente.
- A geração do script pode mudar livremente, já que só `pkg/` e a raiz têm compromisso de compatibilidade.
- A raiz precisa ser mantida em sincronia com `pkg/`. Um teste deve verificar se todo símbolo exportado em `pkg/` tem um alias na raiz.
- O código atual ainda gera SQL dentro de `pkg/constraints` (`Builder strings.Builder`, `Build()`). Isso é provisório, serve só para validar
  a construção, e deve sair de `pkg/` ([ADR-0003](0003-declaracoes-e-geracao.md)).
