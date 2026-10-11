# ADR-0006: CLI `mph` com programa gerado

- **Status:** Proposto
- **Data:** 2026-10-11

## Contexto

O cliente escreve migrations (`Up`/`Down`) em Go, num diretório dedicado do seu projeto, e quer aplicá-las com `mph up`.
Go é compilado, então um binário `mph` já instalado **não consegue carregar código Go do cliente em tempo de execução**.
Além disso, o código do cliente fica em outro módulo e não pode importar `internal/` do morph.

## Decisão

A CLI `mph` (`cmd/mph`) vai **gerar um programa temporário e executá-lo dentro do módulo do cliente**. É o mesmo mecanismo do `go test`,
que gera um `main` que chama `testing.MainStart`.

### Convenções do projeto cliente

```
projeto-do-cliente/
├── go.mod                    require github.com/An-Ordinary-Software-Engineering/morph vX.Y.Z
├── morph.yaml                provider, diretório das migrations e DSN
└── migrations/               pacote Go dedicado (package migrations)
    ├── 20261011120000_create_user.go
    └── 20261012090000_add_role.go
```

```yaml
# morph.yaml
provider: mysql
dir: ./migrations
dsn: ${DATABASE_URL}
```

- Uma migration é um tipo exportado que implementa `morph.Migration`. A versão vem de `Version()` e segue o formato `AAAAMMDDHHMMSS`.
- `mph new <nome>` cria o arquivo já com a versão e o esqueleto de `Up`/`Down`.

### Fluxo de `mph up`

1. Lê o `morph.yaml`.
2. Analisa o pacote de `dir` com `go/packages` e encontra os tipos que implementam `Migration`. Não é preciso `init()` nem registro manual.
3. Gera `.morph/main.go` no módulo do cliente:

   ```go
   package main

   import (
       "os"

       "github.com/An-Ordinary-Software-Engineering/morph"
       _ "github.com/go-sql-driver/mysql" // só o driver do provider configurado
       "projeto-do-cliente/migrations"
   )

   func main() {
       morph.Main(os.Args[1:], migrations.CreateUser{}, migrations.AddRole{})
   }
   ```

4. Executa `go run ./.morph up` no diretório do cliente.
5. `morph.Main` (alias de `pkg/runner.Main`) delega a `internal/src/runner`. O runner valida as declarações, ordena as migrations por versão,
   consulta a tabela de versões, gera o script com `internal/src/builder` e o provider escolhido e o executa numa transação, quando o banco suportar DDL transacional.

### Comandos

| Comando             | Efeito |
| ------------------- | ------ |
| `mph new <nome>`    | Cria uma migration com a versão atual. |
| `mph up`            | Aplica as migrations pendentes. |
| `mph down [n]`      | Reverte as últimas `n` migrations (padrão: 1). |
| `mph status`        | Lista as migrations aplicadas e as pendentes. |
| `mph script`        | Gera o SQL sem executar (*dry-run*). |
| `mph build -o bin`  | Gera um binário autônomo com as migrations, para ambientes sem o toolchain Go. |

### Ponto de entrada público

`pkg/runner.Main` é o **único** ponto de `pkg/` que importa `internal/`. Ele é fino, documentado como "uso por código gerado"
e reexportado na raiz como `morph.Main`. O builder, os providers e o runner continuam internos.

## Alternativas consideradas

- **Runner público com `main` escrito pelo cliente:** é mais simples, mas exige boilerplate em todo projeto e não atende ao `mph up` buscando num diretório.
- **Pacote `plugin` do Go:** só funciona em Linux/macOS, exige o mesmo compilador e as mesmas versões de dependências, e é frágil.
- **Interpretador embutido (ex.: `yaegi`):** suporte incompleto à linguagem, execução mais lenta e comportamento diferente do código compilado.
- **Migrations em SQL ou YAML em vez de Go:** perde a API fluente, que é o propósito do projeto.

## Consequências

- **Versão coerente:** o programa gerado compila com a versão do morph declarada no `go.mod` do cliente, então a CLI e a biblioteca nunca ficam com versões diferentes.
  Ainda assim, a CLI deve verificar se o protocolo de `Main` é compatível e avisar quando não for.
- `mph up`, `down`, `status` e `script` exigem o toolchain Go. Para produção sem Go, o caminho é usar `mph build`.
- O driver do banco precisa estar no `go.mod` do cliente. A CLI deve detectar quando ele falta e sugerir o `go get` correspondente.
- `.morph/` é um diretório gerado, que o cliente deve ignorar no controle de versão. O `mph new` pode adicioná-lo ao `.gitignore`.
- A primeira execução compila o programa. As seguintes reaproveitam o cache de build do Go.
