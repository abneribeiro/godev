# DevScope TUI - Especificação Completa

**Versão:** 0.1.0
**Data:** 2025-10-16
**Status:** MVP em desenvolvimento

---

## 📋 Visão Geral

**DevScope** é uma aplicação TUI (Text-based User Interface) para inspeção e teste de APIs HTTP diretamente do terminal. Oferece uma alternativa leve e focada ao Postman/Insomnia para desenvolvedores que preferem o ambiente terminal.

### Problema que Resolve

- Evitar abrir navegador ou aplicações pesadas para testes rápidos de API
- Ter histórico local de requests para reuso
- Visualizar respostas JSON formatadas diretamente no terminal
- Workflow rápido e integrado ao terminal durante desenvolvimento

---

## 🎯 Objetivo do MVP

Criar um **HTTP Inspector** funcional que permita:

1. Fazer requests HTTP (GET, POST, PUT, DELETE, PATCH)
2. Visualizar respostas formatadas com syntax highlighting
3. Salvar requests para reuso
4. Adicionar headers e body customizados

---

## 🏗️ Arquitetura

```
devscope/
├── main.go                    # Entry point + setup Bubbletea
├── internal/
│   ├── ui/
│   │   ├── model.go           # Model principal (state machine)
│   │   ├── views.go           # Funções View() para cada tela
│   │   ├── update.go          # Lógica Update() por tela
│   │   ├── styles.go          # Estilos Lipgloss centralizados
│   │   └── components.go      # Componentes reutilizáveis
│   ├── http/
│   │   └── client.go          # Cliente HTTP e estruturas
│   └── storage/
│       ├── requests.go        # Gerenciamento de requests salvos
│       └── config.go          # Configuração da aplicação
├── go.mod
├── go.sum
├── README.md
└── DEVSCOPE.md               # Este arquivo
```

---

## 📊 State Machine

### Estados da Aplicação

```go
type AppState int

const (
    StateRequestBuilder AppState = iota  // Tela de criação/edição de request
    StateLoading                         // Executando request HTTP
    StateViewResponse                    // Visualizando resposta
    StateRequestList                     // Lista de requests salvos
    StateHeaderEditor                    // Editando headers
    StateBodyEditor                      // Editando body JSON
    StateHelp                            // Tela de ajuda
)
```

### Fluxo de Estados

```
┌─────────────────┐
│ Request Builder │ ←──────────────────┐
└────────┬────────┘                    │
         │ (Enter: Send)               │
         ↓                             │
    ┌─────────┐                        │
    │ Loading │                        │
    └────┬────┘                        │
         │                             │
         ↓                             │
┌─────────────────┐                    │
│ View Response   │ ──(Esc: Back)──────┘
└────────┬────────┘
         │ (s: Save)
         ↓
┌─────────────────┐
│ Request List    │ ──(Enter: Load)──→ Request Builder
└─────────────────┘
```

---

## 🎨 Design System

### Paleta de Cores

```go
// Base
ColorBg      = "#0D0D0D"  // Fundo principal
ColorPanel   = "#1A1A1A"  // Fundo de painéis/cards
ColorBorder  = "#2D2D2D"  // Bordas

// Texto
ColorText    = "#E4E4E4"  // Texto principal
ColorMuted   = "#888888"  // Texto secundário
ColorDim     = "#555555"  // Texto desabilitado

// Acento
ColorAccent  = "#FF8C00"  // Laranja (ações principais)
ColorSuccess = "#00C853"  // Verde (sucesso)
ColorError   = "#D32F2F"  // Vermelho (erros)
ColorWarning = "#FFA726"  // Amarelo (avisos)

// Códigos HTTP
Color2xx     = "#00C853"  // 2xx Success
Color3xx     = "#FFA726"  // 3xx Redirect
Color4xx     = "#FF5722"  // 4xx Client error
Color5xx     = "#D32F2F"  // 5xx Server error
```

### Tipografia

| Tipo | Estilo | Uso |
|------|--------|-----|
| Título | Bold + ColorAccent | Títulos de seções |
| Texto Principal | Regular + ColorText | Conteúdo geral |
| Texto Secundário | Regular + ColorMuted | Dicas e metadados |
| Código/JSON | Monospace + ColorText | Respostas e bodies |

### Componentes Base

#### Painel
```
┌─────────────────────────────────────┐
│ Título                              │
├─────────────────────────────────────┤
│ Conteúdo                            │
│                                     │
└─────────────────────────────────────┘
```

#### Botão Ativo
```
[ Texto do Botão ]
```
- Fundo: ColorAccent (#FF8C00)
- Texto: ColorBg (#0D0D0D)
- Padding: 0 horizontal, 2 vertical

#### Botão Inativo
```
Texto do Botão
```
- Fundo: Transparente
- Texto: ColorText (#E4E4E4)

#### Input Field (Focused)
```
┌─────────────────────────────────────┐
│ valor do input_                     │
└─────────────────────────────────────┘
```
- Borda: ColorAccent (#FF8C00)

#### Input Field (Unfocused)
```
┌─────────────────────────────────────┐
│ valor do input                      │
└─────────────────────────────────────┘
```
- Borda: ColorBorder (#2D2D2D)

---

## 🖥️ Telas da Aplicação

### 1. Request Builder (Estado Inicial)

```
┌─────────────────────────────────────────────────────┐
│ DevScope v0.1.0                                     │
├─────────────────────────────────────────────────────┤
│                                                     │
│  Method: [GET ▾]                                     │
│  URL:    [____________________________________]      │
│                                                     │
│  Headers: (0)  [Edit]                                │
│  Body:    (empty) [Edit]                             │
│                                                     │
│                                                     │
│  [ Send Request ]  [ Load Saved ]  [ Quit ]          │
│                                                     │
└─────────────────────────────────────────────────────┘
  Tab: next • Enter: action • Ctrl+Q: quit • Ctrl+?: help
```

**Campos:**
- **Method**: Seletor dropdown (GET, POST, PUT, DELETE, PATCH)
- **URL**: Text input editável
- **Headers**: Botão que abre HeaderEditor (mostra quantidade)
- **Body**: Botão que abre BodyEditor (mostra status)

**Navegação:**
- `Tab` / `Shift+Tab`: Navegar entre campos
- `Enter`: Ativar ação (send, edit, load)
- `Ctrl+Q`: Sair da aplicação
- `Ctrl+?`: Abrir ajuda

---

### 2. Loading (Durante Request)

```
┌─────────────────────────────────────────────────────┐
│ DevScope v0.1.0                                     │
├─────────────────────────────────────────────────────┤
│                                                     │
│                                                     │
│              Sending request...                      │
│              GET /api/users                          │
│                                                     │
│                  ⣾ Loading                          │
│                                                     │
│                                                     │
└─────────────────────────────────────────────────────┘
```

**Comportamento:**
- Spinner animado
- Mostra método e path do endpoint
- Não permite interação (exceto Ctrl+C para cancelar)

---

### 3. View Response (Após Request)

```
┌─────────────────────────────────────────────────────┐
│ Response                                            │
├─────────────────────────────────────────────────────┤
│ Status: 200 OK • 142ms • 1.3KB                       │
├─────────────────────────────────────────────────────┤
│                                                     │
│ {                                                   │
│   "users": [                                        │
│     {                                               │
│       "id": 1,                                      │
│       "name": "Alice",                              │
│       "email": "alice@example.com"                  │
│     },                                              │
│     {                                               │
│       "id": 2,                                      │
│       "name": "Bob",                                │
│       "email": "bob@example.com"                    │
│     }                                               │
│   ]                                                 │
│ }                                                   │
│                                                     │
├─────────────────────────────────────────────────────┤
│ [ Back ]  [ Save Request ]  [ Copy to Clipboard ]   │
└─────────────────────────────────────────────────────┘
  Esc: back • s: save • c: copy • ↑↓: scroll
```

**Elementos:**
- **Status bar**: Status code (colorido por tipo), tempo de resposta, tamanho
- **Body**: JSON com syntax highlighting
- **Scroll**: Suporte para respostas longas

**Cores do Status:**
- 2xx: Verde (#00C853)
- 3xx: Amarelo (#FFA726)
- 4xx: Laranja (#FF5722)
- 5xx: Vermelho (#D32F2F)

**Navegação:**
- `Esc`: Voltar para Request Builder
- `s`: Salvar request
- `c`: Copiar resposta para clipboard
- `↑/↓`: Scroll (se conteúdo maior que tela)

---

### 4. Request List (Requests Salvos)

```
┌─────────────────────────────────────────────────────┐
│ Saved Requests                                      │
├─────────────────────────────────────────────────────┤
│                                                     │
│ > Get All Users              GET                     │
│   Create User                POST                    │
│   Update User Profile        PUT                     │
│   Delete User                DELETE                  │
│                                                     │
│                                                     │
│ [ New Request ]  [ Delete ]  [ Back ]                │
│                                                     │
└─────────────────────────────────────────────────────┘
  ↑↓: navigate • Enter: load • d: delete • n: new • Esc: back
```

**Estrutura:**
- Lista de requests com nome e método
- Item selecionado destacado com `>`

**Navegação:**
- `↑/↓`: Navegar na lista
- `Enter`: Carregar request selecionado
- `d`: Deletar request selecionado
- `n`: Criar novo request (Request Builder vazio)
- `Esc`: Voltar

---

### 5. Header Editor

```
┌─────────────────────────────────────────────────────┐
│ Headers Editor                                      │
├─────────────────────────────────────────────────────┤
│                                                     │
│ > Content-Type: application/json                     │
│   Authorization: Bearer token123                     │
│   X-Custom-Header: value                             │
│                                                     │
│   [ Add Header ]                                     │
│                                                     │
│ [ Done ]  [ Cancel ]                                 │
│                                                     │
└─────────────────────────────────────────────────────┘
  ↑↓: navigate • Enter: edit • Ctrl+N: add • Ctrl+D: delete
```

**Funcionalidades:**
- Lista de headers existentes (key: value)
- Adicionar novo header
- Editar header existente
- Deletar header

**Navegação:**
- `↑/↓`: Navegar entre headers
- `Enter`: Editar header selecionado
- `Ctrl+N`: Adicionar novo header
- `Ctrl+D`: Deletar header selecionado
- `Esc`: Salvar e voltar

---

### 6. Body Editor

```
┌─────────────────────────────────────────────────────┐
│ Body Editor (JSON)                                  │
├─────────────────────────────────────────────────────┤
│                                                     │
│ {                                                   │
│   "name": "John Doe",                               │
│   "email": "john@example.com",                      │
│   "age": 30_                                        │
│ }                                                   │
│                                                     │
│                                                     │
│ [ Done ]  [ Validate ]  [ Cancel ]                   │
│                                                     │
└─────────────────────────────────────────────────────┘
  Ctrl+S: save • Ctrl+V: validate • Esc: cancel
```

**Funcionalidades:**
- Textarea multi-linha para editar JSON
- Validação de JSON antes de salvar
- Feedback visual se JSON inválido

**Navegação:**
- Editor de texto livre
- `Ctrl+S`: Salvar e voltar
- `Ctrl+V`: Validar JSON (mostra erro se inválido)
- `Esc`: Cancelar e voltar

---

### 7. Help Screen

```
┌─────────────────────────────────────────────────────┐
│ DevScope - Help                                     │
├─────────────────────────────────────────────────────┤
│                                                     │
│ Global Shortcuts:                                    │
│   Ctrl+Q        Quit application                     │
│   Ctrl+?        Show this help                       │
│   Esc           Back/Cancel                          │
│   Tab           Next field                           │
│   Shift+Tab     Previous field                       │
│                                                     │
│ Request Builder:                                     │
│   Enter         Send request                         │
│   Ctrl+L        Load saved requests                  │
│                                                     │
│ Response View:                                       │
│   s             Save request                         │
│   c             Copy response                        │
│   ↑/↓           Scroll                               │
│                                                     │
│ Request List:                                        │
│   Enter         Load request                         │
│   d             Delete request                       │
│   n             New request                          │
│                                                     │
└─────────────────────────────────────────────────────┘
  Press any key to close
```

---

## 💾 Estrutura de Dados

### Config File: `~/.devscope/config.json`

```json
{
  "version": "0.1.0",
  "requests": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Get All Users",
      "method": "GET",
      "url": "https://jsonplaceholder.typicode.com/users",
      "headers": {
        "Content-Type": "application/json"
      },
      "body": "",
      "created_at": "2025-10-16T14:30:00Z",
      "last_used": "2025-10-16T15:45:00Z"
    },
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "name": "Create User",
      "method": "POST",
      "url": "https://jsonplaceholder.typicode.com/users",
      "headers": {
        "Content-Type": "application/json",
        "Authorization": "Bearer test-token"
      },
      "body": "{\n  \"name\": \"John Doe\",\n  \"email\": \"john@example.com\"\n}",
      "created_at": "2025-10-16T14:35:00Z",
      "last_used": "2025-10-16T14:35:00Z"
    }
  ]
}
```

### Go Structs

```go
// SavedRequest representa um request salvo
type SavedRequest struct {
    ID        string            `json:"id"`
    Name      string            `json:"name"`
    Method    string            `json:"method"`
    URL       string            `json:"url"`
    Headers   map[string]string `json:"headers"`
    Body      string            `json:"body"`
    CreatedAt time.Time         `json:"created_at"`
    LastUsed  time.Time         `json:"last_used"`
}

// Config representa a configuração da aplicação
type Config struct {
    Version  string          `json:"version"`
    Requests []SavedRequest  `json:"requests"`
}

// Response representa a resposta HTTP
type Response struct {
    StatusCode   int
    Status       string
    Body         string
    Headers      map[string][]string
    ResponseTime time.Duration
    Size         int64
}
```

---

## 🔧 Tecnologias e Dependências

### Core
- **Go 1.21+**: Linguagem base
- **bubbletea**: Framework TUI (Elm architecture)
- **lipgloss**: Estilização e layout
- **bubbles**: Componentes prontos (textinput, textarea, list)

### Auxiliares
- **chroma**: Syntax highlighting para JSON
- **uuid**: Geração de IDs únicos
- **net/http**: Cliente HTTP padrão

### go.mod
```go
module github.com/abner/devscope

go 1.21

require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/charmbracelet/bubbles v0.18.0
    github.com/alecthomas/chroma v0.10.0
    github.com/google/uuid v1.5.0
)
```

---

## 🎯 Casos de Uso

### Caso 1: Testar GET em API pública
```
1. Abrir DevScope
2. Inserir URL: https://jsonplaceholder.typicode.com/users
3. Pressionar Enter (send request)
4. Ver lista de usuários formatada
5. Pressionar 's' para salvar como "Get JSONPlaceholder Users"
6. Pressionar Esc para voltar
```

### Caso 2: POST com body JSON
```
1. Abrir DevScope
2. Mudar method para POST (setas ou Tab)
3. URL: https://httpbin.org/post
4. Tab para "Edit Body"
5. Escrever: {"test": "data"}
6. Esc para voltar
7. Tab para "Edit Headers"
8. Adicionar: Content-Type: application/json
9. Esc para voltar
10. Enter para Send Request
11. Ver echo do body na resposta
```

### Caso 3: Reusar request salvo
```
1. Abrir DevScope
2. Pressionar 'Load Saved' ou Ctrl+L
3. Navegar com setas até "Get JSONPlaceholder Users"
4. Pressionar Enter
5. Request é carregado automaticamente
6. Pressionar Enter novamente para re-executar
```

### Caso 4: Request com autenticação
```
1. Abrir DevScope
2. URL: https://api.github.com/user
3. Tab para "Edit Headers"
4. Adicionar: Authorization: token ghp_xxxxxxxxxxxx
5. Esc para voltar
6. Enter para Send Request
7. Ver dados do usuário autenticado
8. Salvar como "GitHub Get User"
```

---

## 🚀 Roadmap

### MVP (v0.2.0) - Completo ✅
- [x] Setup projeto e estrutura
- [x] Request Builder básico (GET)
- [x] HTTP client
- [x] View response com JSON formatado
- [x] Persistência de requests
- [x] Lista de requests salvos
- [x] Suporte a todos métodos HTTP (GET, POST, PUT, DELETE, PATCH)
- [x] Tela de ajuda
- [x] Navegação com Tab entre campos
- [x] Input responsivo
- [x] Header editor (IMPLEMENTADO v0.2.0)
  - Adicionar, editar e deletar headers
  - UI consistente com painéis
  - Validação de campos
- [ ] Body editor JSON (próxima versão v0.3.0)

### v0.2.0 - Features Adicionais
- [ ] Themes customizáveis
- [ ] Query params editor
- [ ] Response headers viewer
- [ ] Export request para cURL
- [ ] Import cURL para request
- [ ] Search/filter na lista de requests
- [ ] Folders/collections para organizar requests

### v0.3.0 - Advanced
- [ ] Environment variables
- [ ] Request chaining (usar response em outro request)
- [ ] GraphQL support
- [ ] WebSocket inspector
- [ ] Request diff (comparar respostas)

### v1.0.0 - Production Ready
- [ ] Database explorer (SQL/NoSQL)
- [ ] Metrics dashboard
- [ ] Performance benchmarking
- [ ] Plugin system
- [ ] Cloud sync (opcional)

---

## 🧪 Testes

### Testes Unitários
- `internal/http`: Testes de HTTP client com mock server
- `internal/storage`: Testes de leitura/escrita de config
- `internal/ui`: Testes de state transitions

### Testes de Integração
- Fluxo completo: criar request → enviar → salvar → carregar
- Teste com APIs públicas reais (JSONPlaceholder, httpbin)

### Testes Manuais
- Verificar UI em diferentes tamanhos de terminal
- Testar com respostas grandes (scroll)
- Testar com JSON inválido
- Testar timeout e erros de rede

---

## 📝 Contribuindo

### Padrões de Código
- Usar `gofmt` para formatação
- Seguir [Effective Go](https://golang.org/doc/effective_go)
- Comentar funções públicas
- Manter funções pequenas e focadas

### Estrutura de Commits
```
<tipo>: <descrição curta>

<descrição detalhada opcional>
```

Tipos:
- `feat`: Nova funcionalidade
- `fix`: Correção de bug
- `refactor`: Refatoração de código
- `docs`: Documentação
- `test`: Testes
- `style`: Formatação, lint

### Workflow
1. Fork do repositório
2. Criar branch: `git checkout -b feat/nova-feature`
3. Commit: `git commit -m "feat: adiciona export para cURL"`
4. Push: `git push origin feat/nova-feature`
5. Abrir Pull Request

---

## 🐛 Troubleshooting

### Problema: Terminal não renderiza cores
**Solução**: Verificar se terminal suporta true color
```bash
echo $COLORTERM  # deve retornar "truecolor" ou "24bit"
```

### Problema: Config não é salvo
**Solução**: Verificar permissões do diretório
```bash
mkdir -p ~/.devscope
chmod 755 ~/.devscope
```

### Problema: JSON não é formatado
**Solução**: Verificar se chroma está instalado corretamente
```bash
go get github.com/alecthomas/chroma
```

---

## 📚 Referências

- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Lipgloss Examples](https://github.com/charmbracelet/lipgloss/tree/master/examples)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Go HTTP Client](https://pkg.go.dev/net/http)
- [Chroma Syntax Highlighting](https://github.com/alecthomas/chroma)

---

## 📄 Licença

MIT License - Veja LICENSE para detalhes

---

## 👥 Autores

- Abner - Desenvolvedor principal

---

**Última atualização:** 2025-10-16
