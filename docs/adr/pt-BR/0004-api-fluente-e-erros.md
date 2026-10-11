# ADR-0004: API fluente e tratamento de erros

- **Status:** Proposto
- **Data:** 2026-10-11

## Contexto

O projeto se inspira na semântica do Fluent Migrator, em que as migrations são declaradas por encadeamento de chamadas.
A API atual já segue esse princípio, com um vocabulário próprio:

```go
migrator.CreateTable("user", func(t *table.CreateTableConstraint) *table.CreateTableConstraint {
    return t.
        Column("id", types.Integer{}).PrimaryKey().AutoIncrement().
        Column("name", types.VarChar{Length: 50}).NotNull()
}, table.Options{DropIfExists: true, Engine: "INNODB"})
```

Ela tem quatro problemas:

- **Panic fora de ordem:** chamar um modificador antes de `Column`, por exemplo `t.NotNull()`, acessa `payload[-1]` e causa **panic**.
- **Função de pacote solta:** `CreateTable` é uma função de pacote que devolve um valor. O runner não tem como coletar as operações de um `Up`/`Down` sem estado global.
- **Default preso ao MySQL:** `Default(value string)` recebe SQL pronto (`"CURRENT_TIMESTAMP()"`).
- **Opção presa ao MySQL:** `Options.Engine` é específico do MySQL ([ADR-0003](0003-declaracoes-e-geracao.md)).

## Decisão

1. **Manteremos o estilo `Column(name, type)`.** O tipo é um valor de `pkg/types`, neutro em relação ao banco, informado junto com o nome.
   Os modificadores vêm em seguida, encadeados.

   | Modificador         | Significado                                                       |
   | ------------------- | ----------------------------------------------------------------- |
   | `PrimaryKey()`      | Faz parte da chave primária. Implica `NotNull()`.                 |
   | `AutoIncrement()`   | Valor gerado pelo banco (`AUTO_INCREMENT`, `IDENTITY`, `SERIAL`, conforme o provider). |
   | `NotNull()`         | Não aceita `NULL`. Sem ele, a coluna aceita `NULL`.               |
   | `Unique()`          | Restrição de unicidade.                                           |
   | `Default(v)`        | `v` é `types.CurrentTimestamp` (ou outro valor calculado pelo banco) ou um literal `string`, `bool` ou numérico. |

2. **Tipagem por etapa:** a tabela só oferece `Column`. `Column` retorna um builder de coluna que oferece os modificadores e o próximo `Column`.
   Chamar um modificador antes de declarar uma coluna passa a ser **erro de compilação**, e o panic deixa de existir.
   Por isso, o blueprint não retorna nada: `func(t *table.CreateTableConstraint)`.

3. **Operações a partir de um `Migrator`:** `Up` e `Down` recebem um `*migrator.Migrator` (reexportado como `*morph.Migrator`).
   Ele registra as operações na ordem em que são declaradas (`CreateTable`, `DropTable` e, no futuro, `AlterTable`, `RenameTable`, `Execute`),
   e o runner as lê depois:

   ```go
   type Migration interface {
       Version() int64       // AAAAMMDDHHMMSS
       Up(m *Migrator)
       Down(m *Migrator)
   }
   ```

4. **Opções neutras com extensão por provider:** `table.Options` mantém só os campos neutros (`DropIfExists`).
   Opções de um banco específico vão em `Options.Provider`, um mapa com chaves que usam o provider como namespace ([ADR-0003](0003-declaracoes-e-geracao.md)).

5. **Erros acumulados, nunca panic.** Os builders guardam erros de validação: nome vazio, coluna duplicada, tamanho inválido,
   `AutoIncrement` em coluna não inteira, `Default` de tipo não suportado etc. O runner os retorna (`mph up`) antes de executar qualquer SQL,
   juntos com `errors.Join`, como o `bufio.Scanner` faz com `Err()`. As mensagens incluem a migration, a tabela e a coluna.

### Exemplo (importando pela raiz)

```go
package migrations

import "github.com/An-Ordinary-Software-Engineering/morph"

type CreateUser struct{}

func (CreateUser) Version() int64 { return 20261011120000 }

func (CreateUser) Up(m *morph.Migrator) {
	m.CreateTable("user", func(t *morph.Table) {
		t.Column("id", morph.Integer{}).PrimaryKey().AutoIncrement().
			Column("name", morph.VarChar{Length: 50}).NotNull().
			Column("email", morph.VarChar{Length: 50}).NotNull().Unique().
			Column("created_on", morph.Timestamp{}).NotNull().Default(morph.CurrentTimestamp)
	}, morph.TableOptions{
		DropIfExists: true,
		Provider:     map[string]string{"mysql.engine": "INNODB"},
	})
}

func (CreateUser) Down(m *morph.Migrator) {
	m.DropTable("user")
}
```

## Alternativas consideradas

- **Vocabulário literal do Fluent Migrator** (`WithColumn(name).AsString(50)`): o tipo fica numa etapa separada do nome,
  e cada tipo vira um método. Com `Column(name, type)`, o tipo é um valor de `pkg/types`, e adicionar um tipo novo não muda a API dos builders.
- **Manter `CreateTable` como função de pacote e coletar as operações num registro global:** cria estado global, dificulta testes e não é seguro para concorrência.
- **Retornar `(builder, error)` em cada chamada:** quebra o encadeamento.
- **`panic` + `recover` interno:** esconde o fluxo de erro e é estranho para usuários Go.

## Consequências

- Mudanças incompatíveis em relação à API atual: o blueprint deixa de retornar valor, `CreateTable` vira método de `Migrator`,
  `Default` deixa de receber SQL e `Engine` sai de `Options`. Isso é aceitável enquanto o projeto está em `v0`.
- Na raiz, alguns nomes precisam de prefixo para não colidir nem ficarem ambíguos (`table.Options` → `morph.TableOptions`,
  `table.CreateTableConstraint` → `morph.Table`).
- Os erros só aparecem quando o runner valida, e não na linha exata da chamada. Por isso, as mensagens precisam identificar a migration, a tabela e a coluna.
