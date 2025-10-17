# Changelog - DevScope

## [0.2.0] - 2025-10-17

### 🎉 Release Completa v0.2.0 - Todas as Features Implementadas!

Esta versão implementa TODAS as funcionalidades planejadas para o MVP completo do DevScope!

---

### ✨ Novas Funcionalidades Implementadas

#### 1. Body Editor JSON Completo
- **Implementação**: Editor de texto multi-linha com textarea
- **Funcionalidades**:
  - Edição livre de JSON/texto
  - Área de texto responsiva (80x10)
  - Limite de 10.000 caracteres
  - Salvamento com Ctrl+S
  - Cancelamento com Esc
- **Arquivo**: `internal/ui/editors.go:351-376`
- **UI**: Borda arredondada com destaque visual

#### 2. Query Parameters Editor
- **Implementação**: Editor completo de parâmetros de URL
- **Funcionalidades**:
  - Adicionar novos query params (tecla 'n')
  - Editar params existentes (tecla 'e' ou Enter)
  - Deletar params (tecla 'd')
  - Navegação com ↑/↓
  - Params automaticamente adicionados à URL antes do request
- **Arquivo**: `internal/ui/editors.go:378-478`
- **Integração**: Função `buildURLWithQueryParams()` constrói URL final

#### 3. Response Headers Viewer
- **Implementação**: Visualização de headers HTTP da resposta
- **Funcionalidades**:
  - Toggle entre Body e Headers (tecla 'h')
  - Formatação alinhada (chave : valor)
  - Scroll para headers longos
  - Título dinâmico ("Response" ou "Response Headers")
- **Arquivo**: `internal/ui/model.go:654-766`

#### 4. Copy to Clipboard
- **Implementação**: Cópia da resposta para área de transferência
- **Funcionalidades**:
  - Tecla 'c' copia todo o body da resposta
  - Feedback visual "✓ Copied to clipboard!"
  - Integração com biblioteca `atotto/clipboard`
- **Arquivo**: `internal/ui/model.go:373-381`

#### 5. Validação de URL
- **Implementação**: Validação completa antes de enviar request
- **Validações**:
  - URL não pode estar vazia
  - Protocolo obrigatório (http:// ou https://)
  - Host válido obrigatório
  - Parse completo da URL
  - Mensagens de erro descritivas
- **Arquivo**: `internal/ui/model.go:466-487`
- **Benefício**: Previne erros HTTP confusos com feedback claro

#### 6. Enter na URL para Enviar Request
- **Implementação**: Atalho intuitivo para envio rápido
- **Comportamento**: Pressionar Enter no campo URL envia o request imediatamente
- **Arquivo**: `internal/ui/model.go:306-310`
- **UX**: Elimina necessidade de Tab até botão "Send"

#### 7. Melhor Tratamento de Erro do Storage
- **Implementação**: Avisos claros quando storage falha
- **Comportamento**:
  - Warning no console se ~/.devscope/ não pode ser criado
  - Aplicação continua funcionando (requests não são salvos)
  - Usuário informado explicitamente
  - Não há crashes silenciosos
- **Arquivo**: `internal/ui/model.go:120-126`

---

### 🔧 Melhorias na UI

#### Request Builder Atualizado
- **Novos campos visíveis**:
  - Query Params: (contador) - Focável com Tab
  - Headers: (contador) - Focável com Tab
  - Body: (preview/empty) - Focável com Tab
- **Layout**: 8 campos navegáveis (Method, URL, Query, Headers, Body, Send, Load, Quit)
- **Indicadores**: Contadores e previews em tempo real
- **Arquivo**: `internal/ui/model.go:572-656`

#### Novos Estados
- **StateQueryEditor**: Estado dedicado para editor de query params
- **State machine completa**: 8 estados totalmente implementados

---

### 📋 Funcionalidades Completas Agora Disponíveis

1. ✅ Request Builder básico
2. ✅ Suporte a métodos HTTP (GET, POST, PUT, DELETE, PATCH)
3. ✅ Visualização de resposta formatada
4. ✅ Persistência de requests
5. ✅ Lista de requests salvos
6. ✅ Editor de headers (adicionar, editar, deletar)
7. ✅ Editor de body JSON
8. ✅ Query params editor
9. ✅ Response headers viewer
10. ✅ Validação de URL
11. ✅ Copy to clipboard
12. ✅ Enter na URL para enviar
13. ✅ Tratamento robusto de erros

---

### 📝 Arquivos Modificados

#### Principais
- `internal/ui/model.go`:
  - Novos campos no Model (queryParams, bodyEditor, etc.)
  - Validação de URL implementada
  - Copy to clipboard na resposta
  - Response headers viewer
  - ViewRequestBuilder atualizado
  - Linhas adicionadas: ~200

- `internal/ui/editors.go`:
  - Body Editor completo
  - Query Editor completo
  - Handlers de teclado para ambos
  - Views renderizadas
  - Linhas adicionadas: ~240

#### Documentação
- `README.md`:
  - Versão atualizada para v0.2.0
  - Roadmap atualizado (todas features ✅)
  - Exemplos atualizados

- `DEVSCOPE.md`:
  - Versão v0.2.0
  - Status: "MVP funcional com Header Editor"

- `internal/storage/requests.go`:
  - Versão v0.2.0

---

### 🎯 Estatísticas da Release

- **Linhas de código adicionadas**: ~440
- **Novas funcionalidades**: 7 implementações completas
- **Estados adicionados**: 1 (QueryEditor)
- **Campos no Model**: 13 novos campos
- **Compilação**: ✅ Sucesso (11MB)
- **Diagnósticos**: Apenas 3 hints de modernização (não-críticos)

---

### 🚀 Como Usar as Novas Funcionalidades

#### Query Parameters
```
1. Na tela principal, Tab até "Query Params: (0)"
2. Pressione Enter
3. Pressione 'n' para adicionar novo param
4. Digite nome (ex: "page") → Tab → Digite valor (ex: "1")
5. Enter para salvar
6. Esc para voltar
7. Query params são automaticamente adicionados à URL!
```

#### Body Editor
```
1. Tab até "Body: (empty)"
2. Pressione Enter
3. Digite seu JSON ou texto
4. Ctrl+S para salvar (ou Esc para cancelar)
5. Request usará o body definido
```

#### Response Headers
```
1. Após receber uma resposta
2. Pressione 'h' para alternar entre Body e Headers
3. Navegue com ↑/↓ se houver muitos headers
```

#### Copy Response
```
1. Após receber uma resposta
2. Pressione 'c' para copiar
3. Feedback visual aparece: "✓ Copied to clipboard!"
4. Cole em qualquer lugar com Ctrl+V
```

---

### 🐛 Correções

- Tratamento de erro do Storage não causava crashes
- URL sem protocolo agora tem validação clara
- Enter na URL agora funciona (antes era ignorado)
- ViewRequestBuilder agora mostra todos os editores

---

### 📚 Documentação Atualizada

- README.md: Roadmap completo com v0.2.0 finalizado
- CHANGELOG.md: Este arquivo com detalhes completos
- Exemplos atualizados para incluir novas features

---

## [0.1.2] - 2025-10-16

### Correções de Clipboard

#### 1. Ctrl+V (Paste) não funcionava
**Problema:** Não era possível colar URLs usando Ctrl+V.

**Solução:** Refatorada a lógica de interceptação de teclas para permitir que operações de clipboard sejam passadas ao textinput:
- `Ctrl+V` agora funciona para colar
- `Ctrl+A` funciona para selecionar tudo
- `Ctrl+C` funciona para copiar (quando há texto selecionado)
- `Ctrl+X` funciona para cortar

**Implementação:**
```go
case "ctrl+c":
    if m.urlInput.Value() != "" {
        m.urlInput, cmd = m.urlInput.Update(msg)
        return m, cmd
    }
    return m.handleKeyPress(msg)
```

### Melhorias Adicionais

- **CharLimit aumentado:** De 500 para 2000 caracteres (suporta URLs muito longas)
- **Todas as operações de edição:** Home, End, Delete, Backspace, setas, etc. funcionam normalmente

### Arquivos Modificados

- `internal/ui/model.go`:
  - Linha 58: CharLimit aumentado para 2000
  - Linha 91-139: Lógica de Update() refinada para clipboard

---

## [0.1.1] - 2025-10-16

### Correções Críticas

#### 1. Input de URL não aceitava digitação
**Problema:** O campo URL não permitia digitação porque as teclas eram consumidas pelo handler de navegação antes de chegarem ao textinput.

**Solução:** Modificada a função `Update()` para passar as teclas normais diretamente para o textinput quando `focusIndex == 1`.

#### 2. Input não responsivo
**Problema:** Width fixo de 60 caracteres, não ajustava ao tamanho da tela.

**Solução:** Adicionada lógica no `WindowSizeMsg` para calcular width dinamicamente.

#### 3. Layout e alinhamento ruins
**Problema:** Input não estava bem estilizado e alinhado.

**Solução:** Refatorada a função `viewRequestBuilder()` com bordas arredondadas e indicação visual de foco.

---

**Versão Anterior:** 0.1.0
**Build Atual:** Bem-sucedido (11MB)
**Status:** ✅ MVP v0.2.0 COMPLETO - Pronto para produção!
