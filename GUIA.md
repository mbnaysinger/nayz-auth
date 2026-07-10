# Guia Rápido de Golang (Para Devs Java)

Bem-vindo ao mundo do Go! Este documento serve como uma "colinha" (cheatsheet) para os principais comandos de terminal e as diferenças sintáticas cruciais em relação a linguagens orientadas a objetos tradicionais, como o Java.

## 1. Comandos Principais de Console

O ecossistema Go já inclui quase tudo o que você precisa nativamente (sem a necessidade de ferramentas externas como Maven ou Gradle).

- **`go mod init <nome-do-modulo>`**: Inicializa um novo projeto criando o arquivo `go.mod` (o equivalente ao `pom.xml` do Java ou `package.json` do Node).
- **`go run <arquivo.go>`**: Compila e executa o código em memória num único passo. Ótimo para uso local e testes rápidos. *(Ex: `go run cmd/nayz-auth/main.go`)*
- **`go build -o <caminho_saida> <arquivo.go>`**: Compila o código gerando um artefato (binário executável final nativo da máquina). *(Ex: `go build -o bin/app cmd/nayz-auth/main.go`)*
- **`go get <url-do-pacote>`**: Baixa uma dependência externa (biblioteca) e a adiciona ao seu `go.mod`. *(Ex: `go get github.com/lib/pq` para banco Postgres)*
- **`go mod tidy`**: O "salva-vidas". Ele limpa o `go.mod`, baixando todas as dependências que estão faltando no seu código e removendo as que você parou de usar. Sempre que importar algo novo, rode esse comando.
- **`go fmt ./...`**: Formata automaticamente todo o código do seu projeto para o padrão oficial do Go.

---

## 2. Sintaxe Básica e Diferenças (Go vs Java)

No Go, a declaração de tipos e funções lê-se da **esquerda para a direita**. O nome vem antes do tipo. A razão dos criadores da linguagem (do Google) para isso é facilitar a leitura em voz alta ("a variável NOME é do tipo STRING").

### 2.1. Declaração de Variáveis
```go
// Estilo Java: String nome = "Maria";
// Em Go, forma explícita:
var nome string = "Maria"

// Forma curta, com inferência de tipo (MUITO utilizada):
// O ":=" cria a variável e infere o tipo na hora. Só funciona dentro de funções.
nome := "Maria" // O Go já sabe que é string
idade := 30     // O Go já sabe que é int
```

### 2.2. Funções e Retornos Múltiplos
A sintaxe da função que você notou inverte a ordem. Além disso, diferente do Java, funções em Go podem retornar múltiplos valores de uma vez (isso é a base do tratamento de erros em Go).

```go
// O nome do parâmetro (a, b) vem ANTES do tipo (int). O tipo de retorno (int) vem no final.
func somar(a int, b int) int {
    return a + b
}

// Retornando DOIS valores (o resultado e um possível erro)
func dividir(a float64, b float64) (float64, error) {
    if b == 0 {
        // Retorna 0 pro resultado, e um objeto de erro
        return 0, fmt.Errorf("não é possível dividir por zero")
    }
    // Retorna o resultado e 'nil' (nulo) para o erro, indicando sucesso
    return a / b, nil
}
```

### 2.3. "Classes" vs Structs
O Go **não possui classes, herança (`extends`) ou a palavra `this`**. Em vez de classes, usamos `structs` (estruturas de dados puras) e associamos métodos a elas usando algo chamado *Receiver* (Receptor).

```go
// 1. Definindo a Estrutura (semelhante aos atributos da classe)
type Usuario struct {
    Nome  string
    Email string
}

// 2. Definindo um "Método" para o Usuario
// O "(u Usuario)" antes do nome da função é o Receiver, agindo como o 'this'.
func (u Usuario) Apresentar() string {
    return "Olá, meu nome é " + u.Nome
}

// 3. Instanciando e usando:
func main() {
    user := Usuario{Nome: "Maike", Email: "maike@exemplo.com"}
    fmt.Println(user.Apresentar())
}
```

### 2.4. Ponteiros (Atenção redobrada)
No Java, todo objeto é passado por referência (você altera num método, altera em tudo). No Go, **tudo é passado como cópia por padrão**. Se você quiser alterar o objeto original num método, você precisa usar um ponteiro (`*`), passando o endereço de memória.

```go
// O asterisco (*) indica que estamos recebendo um PONTEIRO para um Usuario.
func (u *Usuario) AlterarNome(novoNome string) {
    u.Nome = novoNome // Agora sim, altera o objeto original!
}
```
*Sempre que vir um `&` (obter o endereço da variável na memória) ou um `*` (declarar que a variável é um ponteiro), é para garantir que a instância original seja lida ou modificada. Por isso na sua rota `/health` tinha um `*http.Request` — você está lidando com a requisição real que bateu no servidor, não uma cópia.*

### 2.5. Laços de Repetição e Condições
O Go valoriza ter apenas uma maneira de fazer as coisas. Portanto, não existe `while` nem `do-while`. **Tudo é `for`**. E os parênteses `()` foram abolidos nas condições!

```go
// FOR tradicional
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// O equivalente ao "while" do Java
contador := 0
for contador < 5 {
    contador++
}

// IF não usa parênteses
if idade >= 18 {
    fmt.Println("Maior de idade")
}
```

---

## 3. Links Oficiais para Estudo

A documentação do Go é considerada uma das melhores do mundo da programação, extremamente pragmática e clara:

1. **A Tour of Go (Obrigatório!):** [https://go.dev/tour/](https://go.dev/tour/)
   - Um tutorial interativo no seu próprio navegador que te ensina a linguagem passo a passo testando código real. Sugiro fortemente fazer.
2. **Documentação Oficial (Home):** [https://go.dev/doc/](https://go.dev/doc/)
   - Ponto de partida para guias, instalação e melhores práticas.
3. **Effective Go:** [https://go.dev/doc/effective_go](https://go.dev/doc/effective_go)
   - Depois que você aprender o básico, este documento é a "Bíblia" de como escrever código idiomático em Go, em vez de tentar escrever Java usando a sintaxe do Go.
4. **Biblioteca Padrão (Standard Library):** [https://pkg.go.dev/std](https://pkg.go.dev/std)
   - Go tem baterias inclusas! É aqui que você pesquisa como usar pacotes nativos como o `net/http` (para servidores e requisições), `fmt` (para prints), `encoding/json` (para serialização de APIs), etc.
