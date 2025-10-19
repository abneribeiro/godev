# GoDev - Documentação Técnica

## Visão Geral

GoDev é uma ferramenta de linha de comando moderna para testar APIs HTTP e interagir com bancos de dados PostgreSQL. Construída com Go e o framework Bubbletea, oferece uma interface TUI (Terminal User Interface) elegante e eficiente.

## Arquitetura

### Estrutura de Diretórios

```
godev/
├── main.go                       # Ponto de entrada da aplicação
├── internal/
│   ├── ui/                       # Camada de interface do usuário
│   │   ├── model.go              # State machine e lógica de negócio
│   │   ├── editors.go            # Editores (Headers, Body, Query Params)
│   │   └── styles.go             # Sistema de estilos visuais
│   ├── http/                     # Camada HTTP
│   │   └── client.go             # Cliente HTTP e utilitários
│   ├── storage/                  # Camada de persistência HTTP
│   │   └── requests.go           # Persistência de requests e histórico
│   └── database/                 # Camada de banco de dados
│       ├── postgres.go           # Cliente PostgreSQL
│       └── storage.go            # Persistência de queries e conexões
└── go.mod
```

### Componentes Principais

#### 1. UI Layer (internal/ui/)

**model.go**: Implementa o padrão Elm Architecture com:
- Estados da aplicação (StateRequestBuilder, StateHistory, StateDatabase, etc.)
- Gerenciamento de inputs e editores
- Lógica de navegação entre estados
- Integração com HTTP client e Database client

**editors.go**: Implementa editores especializados:
- Header Editor: Adicionar/editar/deletar headers HTTP
- Body Editor: Editor de JSON com validação
- Query Params Editor: Gerenciar parâmetros de URL

**styles.go**: Sistema de design com:
- Paleta de cores consistente
- Estilos para diferentes componentes
- Utilitários de renderização

#### 2. HTTP Layer (internal/http/)

**client.go**: Cliente HTTP com suporte a:
- Métodos: GET, POST, PUT, DELETE, PATCH
- Formatação automática de JSON
- Métricas de performance (tempo de resposta, tamanho)
- Export para comandos cURL

Funcionalidades:
```go
func (c *Client) Send(req Request) Response
func RequestToCurl(req Request) string
func FormatSize(bytes int64) string
func FormatDuration(d time.Duration) string
```

#### 3. Storage Layer (internal/storage/)

**requests.go**: Gerenciamento de persistência HTTP:

Estruturas de dados:
```go
type SavedRequest struct {
    ID          string
    Name        string
    Method      string
    URL         string
    Headers     map[string]string
    Body        string
    QueryParams map[string]string
    CreatedAt   time.Time
    LastUsed    time.Time
}

type RequestExecution struct {
    ID           string
    Timestamp    time.Time
    Method       string
    URL          string
    Headers      map[string]string
    Body         string
    QueryParams  map[string]string
    StatusCode   int
    Status       string
    ResponseBody string
    ResponseTime int64
    Error        string
}
```

Funcionalidades:
- Salvar/carregar requests
- Histórico de execuções (últimas 100)
- Busca e filtro de requests
- Migração automática de configs antigas

#### 4. Database Layer (internal/database/)

**postgres.go**: Cliente PostgreSQL com:
- Conexão e gerenciamento de sessão
- Execução de queries (SELECT, INSERT, UPDATE, DELETE)
- Consulta de schema (tabelas, colunas, tipos)
- Métricas de execução

**storage.go**: Persistência de queries e conexões:

Estruturas:
```go
type SavedQuery struct {
    ID        string
    Name      string
    Query     string
    CreatedAt time.Time
    LastUsed  time.Time
}

type QueryExecution struct {
    ID             string
    Timestamp      time.Time
    Query          string
    RowsAffected   int64
    ExecutionTime  int64
    Error          string
    ConnectionInfo string
}

type ConnectionConfig struct {
    Host     string
    Port     int
    Database string
    User     string
    Password string
    SSLMode  string
}
```

## Funcionalidades Implementadas

### 1. Request History Tracking

**Localização**: `internal/storage/requests.go`, `internal/ui/model.go`

**Descrição**: Rastreamento completo do histórico de execuções de requests HTTP.

**Características**:
- Armazena últimas 100 execuções
- Inclui request completo (método, URL, headers, body, query params)
- Registra response (status, body, tempo de execução)
- Registra erros quando ocorrem
- Navegação e visualização em interface dedicada
- Carregar request do histórico
- Deletar execuções individuais
- Limpar todo o histórico

**Atalhos**:
- `Ctrl+R`: Acessar histórico
- `↑↓`: Navegar
- `Enter`: Carregar request
- `d`: Deletar execução
- `c`: Limpar todo histórico (com confirmação)

**Arquivo de dados**: `~/.godev/config.json` (campo `history`)

### 2. Search and Filter Saved Requests

**Localização**: `internal/storage/requests.go`, `internal/ui/model.go`

**Descrição**: Sistema de busca e filtro para requests salvos.

**Características**:
- Busca em tempo real
- Filtro por nome, método e URL
- Case-insensitive
- Feedback visual de resultados
- Preserva funcionalidades (load, delete) na lista filtrada

**Atalhos**:
- `/`: Ativar campo de busca
- `Esc`: Limpar busca
- `Enter`: Confirmar busca

**Implementação**:
```go
func (s *Storage) FilterRequests(query string) []SavedRequest {
    if query == "" {
        return s.config.Requests
    }
    query = strings.ToLower(query)
    filtered := []SavedRequest{}
    for _, req := range s.config.Requests {
        if strings.Contains(strings.ToLower(req.Name), query) ||
           strings.Contains(strings.ToLower(req.Method), query) ||
           strings.Contains(strings.ToLower(req.URL), query) {
            filtered = append(filtered, req)
        }
    }
    return filtered
}
```

### 3. Export to cURL Commands

**Localização**: `internal/http/client.go`, `internal/ui/model.go`

**Descrição**: Exportação de requests para comandos cURL compatíveis.

**Características**:
- Formato cURL multi-linha (legível)
- Suporte a todos os métodos HTTP
- Inclusão de headers
- Inclusão de body (JSON)
- Cópia automática para clipboard
- Feedback visual de sucesso

**Atalhos**:
- `x`: Copiar request atual como cURL

**Formato gerado**:
```bash
curl 'https://api.example.com/endpoint' \
  -X POST \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer token' \
  -d '{"key": "value"}'
```

**Implementação**:
```go
func RequestToCurl(req Request) string {
    var parts []string
    parts = append(parts, "curl")
    parts = append(parts, fmt.Sprintf("'%s'", req.URL))
    if req.Method != "GET" {
        parts = append(parts, "-X", req.Method)
    }
    for key, value := range req.Headers {
        parts = append(parts, "-H", fmt.Sprintf("'%s: %s'", key, value))
    }
    if req.Body != "" {
        parts = append(parts, "-d", fmt.Sprintf("'%s'", req.Body))
    }
    return joinCurlParts(parts)
}
```

### 4. PostgreSQL Database Support

**Localização**: `internal/database/`, `internal/ui/model.go`

**Descrição**: Suporte completo para interação com bancos PostgreSQL.

**Características**:

#### Conexão:
- Gerenciamento de múltiplas conexões
- Salvar configurações de conexão
- SSL Mode configurável
- Validação de conexão (ping)

#### Execução de Queries:
- SELECT (retorna resultados tabulares)
- INSERT, UPDATE, DELETE (retorna rows affected)
- SHOW commands
- Validação de queries vazias
- Métricas de execução

#### Schema Browser:
- Listar todas as tabelas
- Ver detalhes de colunas (nome, tipo, nullable)
- Ordenação alfabética

#### Storage:
- Salvar queries favoritas
- Histórico de execuções (últimas 100)
- Busca e filtro de queries
- Salvar múltiplas conexões

**Atalhos**:
- `Ctrl+D`: Acessar modo Database

**Estrutura de conexão**:
```go
type ConnectionConfig struct {
    Host     string  // ex: localhost
    Port     int     // ex: 5432
    Database string  // nome do database
    User     string  // usuário
    Password string  // senha
    SSLMode  string  // disable, require, verify-ca, verify-full
}
```

**Arquivo de dados**: `~/.godev/database.json`

## Estados da Aplicação

```go
const (
    StateHome                     // Tela inicial com seleção de modo (NOVO v0.4.0)
    StateRequestBuilder           // Tela principal de construção de requests
    StateLoading                  // Carregando response
    StateViewResponse             // Visualizando response
    StateRequestList              // Lista de requests salvos
    StateHeaderEditor             // Editor de headers
    StateBodyEditor               // Editor de body JSON
    StateQueryEditor              // Editor de query params
    StateHelp                     // Tela de ajuda
    StateHistory                  // Histórico de execuções
    StateDatabase                 // Menu principal database
    StateDatabaseConnect          // Formulário de conexão
    StateDatabaseQueryEditor      // Editor de SQL queries
    StateDatabaseResult           // Visualização de resultados
    StateDatabaseQueryList        // Lista de queries salvas
    StateDatabaseSchema           // Visualizador de schema
    StateDatabaseQueryHistory     // Histórico de queries
    StateDatabaseExport           // Exportação de resultados
    StateEnvironments             // Lista de environments (NOVO v0.4.0)
    StateEnvironmentEditor        // Editor de variáveis (NOVO v0.4.0)
)
```

## Atalhos Globais (v0.4.0 - Padronizados)

### F-Keys (Navegação Rápida)
- `F1` / `?`: Mostrar ajuda
- `F2`: Enviar request
- `F3`: Carregar requests salvos
- `F4`: Ver histórico de requests
- `F5`: Acessar modo Database / Executar query SQL
- `F6`: Environment Variables

### Navegação
- `Tab`: Próximo campo
- `↑↓`: Navegar em listas
- `Esc`: Voltar/Cancelar
- `Ctrl+Q` / `Ctrl+C`: Sair da aplicação

### Ações Letter-Keys (API Mode)
- `h`: Edit headers
- `b`: Edit body
- `q`: Edit query parameters
- `s`: Save current request
- `x`: Copy request as cURL
- `c`: Copy response
- `/`: Search (in lists)
- `Ctrl+Q`: Sair da aplicação
- `Ctrl+?`: Mostrar ajuda

### Request Builder
- `←→`: Mudar método HTTP
- `s`: Salvar request
- `x`: Copiar como cURL
- `c`: Copiar response

### Editores
- `n` ou `a`: Adicionar novo item
- `e`: Editar item selecionado
- `d`: Deletar item
- `Ctrl+S`: Salvar e validar (body editor)

### Lists
- `/`: Ativar busca
- `c`: Limpar histórico (com confirmação)
- `y`: Confirmar ação destrutiva

## Configuração e Storage

### Arquivos de Configuração

Todos os arquivos são armazenados em `~/.godev/`:

1. **config.json**: Configuração principal HTTP
```json
{
  "version": "0.2.0",
  "requests": [ /* SavedRequest[] */ ],
  "history": [ /* RequestExecution[] */ ]
}
```

2. **database.json**: Configuração Database
```json
{
  "version": "0.3.0",
  "saved_queries": [ /* SavedQuery[] */ ],
  "query_history": [ /* QueryExecution[] */ ],
  "saved_connections": [ /* ConnectionConfig[] */ ]
}
```

### Migração Automática

A aplicação migra automaticamente configurações antigas de `~/.devscope` para `~/.godev` na primeira execução.

## Performance e Limites

### Limites de Histórico
- Request History: 100 execuções
- Query History: 100 execuções
- Ordenação: Mais recente primeiro

### Timeouts
- HTTP Client: 30 segundos
- Database: Configurável por driver

### Validações
- URL: Protocolo obrigatório (http/https)
- JSON Body: Validação automática
- SQL Query: Não pode ser vazia

## Dependências

### Core
- `github.com/charmbracelet/bubbletea`: Framework TUI
- `github.com/charmbracelet/lipgloss`: Estilos TUI
- `github.com/charmbracelet/bubbles`: Componentes TUI

### Funcionalidades
- `github.com/google/uuid`: Geração de IDs únicos
- `github.com/atotto/clipboard`: Integração com clipboard
- `github.com/lib/pq`: Driver PostgreSQL

## Notas de Implementação

### Reutilização de Código

A implementação seguiu o princípio de máxima reutilização:

1. **Editores**: Header, Body e Query Param editors compartilham mesma estrutura
2. **Storage**: Padrões similares para HTTP requests e Database queries
3. **Histórico**: Mesma lógica para Request History e Query History
4. **Busca/Filtro**: Mesma implementação para requests e queries
5. **Estilos**: Sistema centralizado de estilos compartilhado

### Estados e Transições

O modelo segue Elm Architecture:
- Estado é imutável
- Atualizações através de mensagens
- Renderização reativa
- Separação clara de concerns

### Error Handling

- Erros HTTP são capturados e exibidos
- Erros de validação são mostrados em tempo real
- Erros de conexão DB são tratados graciosamente
- Fallbacks para operações críticas

## Próximas Melhorias Sugeridas

1. **Environment Variables**: Suporte a variáveis de ambiente
2. **Request Chaining**: Usar response em próximo request
3. **GraphQL Support**: Adicionar suporte a GraphQL
4. **WebSocket Inspector**: Inspecionar conexões WebSocket
5. **Collections**: Organizar requests em coleções
6. **Import cURL**: Importar comandos cURL como requests
7. **Custom Themes**: Temas de cores customizáveis
8. **DB Schema Diff**: Comparar schemas entre conexões
9. **Query Builder Visual**: Constructor visual de queries SQL
10. **Export/Import**: Exportar e importar configurações completas

## Versão

Versão atual: **0.4.0**

Mudanças principais desta versão:
- ✅ **Home Screen** - Tela inicial com seleção de modo (API/Database) (COMPLETO)
- ✅ **Atalhos Padronizados** - F-keys para navegação rápida (COMPLETO)
  - F1: Help
  - F2: Send request
  - F3: Load requests
  - F4: History
  - F5: Database/Execute query
  - F6: Environment variables
- ✅ **Environment Variables** - Sistema completo (COMPLETO - 100% funcional)
  - Múltiplos ambientes (dev, staging, prod)
  - CRUD de variáveis
  - Sintaxe de template {{VARIABLE}}
  - Substituição automática em URL, headers e body
  - Indicador visual de ambiente ativo no title
  - Arquivo environments.json
- ✅ **F5 para Queries** - Executar queries SQL com F5 (COMPLETO)
- ✅ Request History Tracking (COMPLETO)
- ✅ Search and Filter para saved requests (COMPLETO)
- ✅ Export to cURL commands (COMPLETO)
- ✅ PostgreSQL Database Support (COMPLETO)

## Fluxo Completo Environment Variables

### Acesso ao Gerenciador
1. Usuário pressiona `F6` do Request Builder → entra em StateEnvironments
2. Vê lista de environments existentes
3. Indicador ★ mostra qual environment está ativo

### Criar Novo Environment
1. Pressiona `n` → abre StateEnvironmentEditor
2. Input aparece para nome do environment
3. Digita nome (ex: "dev", "staging", "prod")
4. Pressiona `Ctrl+S` → environment criado
5. Volta para lista automaticamente

### Adicionar Variáveis
1. Seleciona environment e pressiona `Enter` → abre editor
2. Pressiona `n` para adicionar variável
3. Dialog aparece com dois campos:
   - Key: nome da variável (ex: API_URL)
   - Value: valor da variável (ex: https://api.dev.example.com)
4. `Tab` para navegar entre campos
5. `Enter` para salvar → variável adicionada

### Editar Variável
1. No editor de environment, seleciona variável
2. Pressiona `e` → dialog aparece com valores atuais
3. Modifica key ou value
4. `Enter` para salvar → variável atualizada

### Deletar Variável
1. Seleciona variável
2. Pressiona `d` → confirmação aparece
3. Pressiona `y` → variável deletada

### Definir Environment Ativo
1. Na lista de environments, seleciona um
2. Pressiona `s` → environment definido como ativo
3. Indicador ★ aparece ao lado do nome
4. Mensagem de sucesso: "✓ Environment saved/activated successfully!"

### Usar Variáveis em Requests
1. Volta ao Request Builder (Esc)
2. Title mostra: "GoDev v0.4.0 [ENV: dev]"
3. Usa sintaxe de template em:
   - **URL**: `{{API_URL}}/users`
   - **Headers**: `Authorization: Bearer {{API_TOKEN}}`
   - **Body**: `{"endpoint": "{{API_URL}}"}`
4. Ao enviar request (F2 ou Enter):
   - Função `sendRequest()` chama `storage.GetActiveEnvironmentVariables()`
   - `storage.ReplaceVariables()` substitui {{VAR}} pelos valores
   - Request enviado com valores substituídos

### Exemplo Completo
```
Environment: dev
Variables:
  API_URL = https://api.dev.example.com
  API_TOKEN = dev_token_123

Request Builder:
  URL: {{API_URL}}/users
  Headers: Authorization: Bearer {{API_TOKEN}}

Request enviado:
  URL: https://api.dev.example.com/users
  Headers: Authorization: Bearer dev_token_123
```

### Storage
**Arquivo**: `~/.godev/environments.json`

**Estrutura**:
```go
type Variable struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

type Environment struct {
    Name      string     `json:"name"`
    Variables []Variable `json:"variables"`
}

type EnvironmentConfig struct {
    Version           string        `json:"version"`
    Environments      []Environment `json:"environments"`
    ActiveEnvironment string        `json:"active_environment"`
}
```

**Função de Substituição**:
```go
func ReplaceVariables(text string, variables []Variable) string {
    re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
    result := re.ReplaceAllStringFunc(text, func(match string) string {
        varName := strings.TrimSpace(match[2 : len(match)-2])
        for _, v := range variables {
            if v.Key == varName {
                return v.Value
            }
        }
        return match // Mantém {{VAR}} se não encontrar
    })
    return result
}
```

### Atalhos de Teclado
**Na lista de Environments (StateEnvironments)**:
- `↑↓` / `j k`: Navegar
- `n` / `a`: Novo environment
- `Enter`: Editar environment selecionado
- `s`: Definir como ativo
- `d`: Deletar (com confirmação `y`)
- `Esc`: Voltar ao Request Builder

**No editor de Environment (StateEnvironmentEditor)**:
- `↑↓` / `j k`: Navegar variáveis
- `n` / `a`: Nova variável
- `e`: Editar variável selecionada
- `d`: Deletar variável (confirmação `y`)
- `Ctrl+S`: Salvar environment (apenas ao criar novo)
- `Esc`: Voltar à lista

**No dialog de variável**:
- `Tab`: Próximo campo (Key → Value)
- `Enter`: Salvar variável
- `Esc`: Cancelar

## Fluxo Completo PostgreSQL

### Conexão ao Banco
1. Usuário pressiona `Ctrl+D` → entra em Database Mode
2. Vê menu com opção "Connect to Database"
3. Pressiona `c` → abre formulário de conexão (StateDatabaseConnect)
4. Preenche campos:
   - Host (default: localhost)
   - Port (default: 5432)
   - Database
   - User
   - Password (masked)
5. Pressiona `Enter` → tenta conectar
6. Se sucesso: volta ao menu conectado
7. Se erro: mostra mensagem de erro

### Executando Queries
1. Do menu database, pressiona `q` → abre editor (StateDatabaseQueryEditor)
2. Escreve query SQL no textarea
3. Pressiona `F5` → executa query
4. Transição para StateDatabaseResult
5. Visualiza:
   - Se SELECT: tabela com colunas e linhas
   - Se INSERT/UPDATE/DELETE: rows affected
   - Tempo de execução em ms
6. Pode pressionar `s` para salvar query
7. `Esc` volta ao editor

### Gerenciando Queries Salvas
1. Do menu database, pressiona `l` → abre lista (StateDatabaseQueryList)
2. Navega com `↑↓`
3. Vê preview da query selecionada
4. Pressiona `Enter` → carrega no editor
5. Pressiona `d` → deleta query
6. `Esc` volta ao menu

### Desconectando
1. Do menu database, pressiona `d`
2. Fecha conexão
3. Volta ao estado "não conectado"

---

**Desenvolvido com Go** | Framework: Bubbletea | Licença: MIT
