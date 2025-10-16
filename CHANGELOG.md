# Changelog - DevScope

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
// Tratamento especial para Ctrl+C
case "ctrl+c":
    if m.urlInput.Value() != "" {
        m.urlInput, cmd = m.urlInput.Update(msg)
        return m, cmd
    }
    return m.handleKeyPress(msg) // Quit se campo vazio
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

**Solução:** Modificada a função `Update()` para passar as teclas normais diretamente para o textinput quando `focusIndex == 1`, capturando apenas as teclas de navegação (Tab, Enter, Ctrl+C, etc.) para o handler principal.

```go
// Antes: teclas eram sempre consumidas pelo handleKeyPress
case tea.KeyMsg:
    return m.handleKeyPress(msg)

// Depois: teclas normais vão direto para o textinput
case tea.KeyMsg:
    if m.state == StateRequestBuilder && m.focusIndex == 1 {
        switch msg.String() {
        case "ctrl+c", "ctrl+q", "tab", "shift+tab", "enter", "ctrl+l", "ctrl+?":
            return m.handleKeyPress(msg)
        default:
            m.urlInput, cmd = m.urlInput.Update(msg)
            return m, cmd
        }
    }
```

#### 2. Input não responsivo
**Problema:** Width fixo de 60 caracteres, não ajustava ao tamanho da tela.

**Solução:** Adicionada lógica no `WindowSizeMsg` para calcular width dinamicamente:

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    inputWidth := m.width - 20
    if inputWidth < 40 {
        inputWidth = 40
    }
    if inputWidth > 100 {
        inputWidth = 100
    }
    m.urlInput.Width = inputWidth
```

#### 3. Layout e alinhamento ruins
**Problema:** Input não estava bem estilizado e alinhado.

**Solução:** Refatorada a função `viewRequestBuilder()`:
- URL label em linha separada
- Input com borda arredondada
- Borda laranja quando focado (#FF8C00)
- Borda cinza quando não focado (#2D2D2D)
- Melhor espaçamento entre elementos
- Indicação visual clara de focus

```go
// Nova renderização do input com bordas
if m.focusIndex == 1 {
    inputView := m.urlInput.View()
    styledInput := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color(ColorAccent)).
        Padding(0, 1).
        Width(m.urlInput.Width + 2).
        Render(inputView)
    b.WriteString(styledInput)
}
```

### Melhorias Adicionais

- **Import lipgloss:** Adicionado import explícito para usar estilos inline
- **Feedback visual:** Método HTTP destacado quando focado
- **Headers:** Indicação "[Not implemented yet]" para feature futura
- **Navegação:** Tab funciona corretamente entre todos os campos

### Arquivos Modificados

- `internal/ui/model.go`:
  - Linha 90-132: Refatoração completa da função `Update()`
  - Linha 379-439: Refatoração da função `viewRequestBuilder()`
  - Linha 11: Adicionado import `lipgloss`

### Como Testar

```bash
# Rebuild
go build -o devscope

# Executar
./devscope

# Testar funcionalidades:
# 1. Digite uma URL normalmente (deve funcionar)
# 2. Use Tab para navegar (deve focar corretamente)
# 3. Redimensione o terminal (input deve ajustar)
# 4. Use ←/→ para mudar método HTTP
# 5. Pressione Enter para enviar request
```

### Próximos Passos

- [ ] Implementar editor de Headers
- [ ] Implementar editor de Body JSON
- [ ] Adicionar validação de URL
- [ ] Melhorar feedback de erros
- [ ] Adicionar copy to clipboard na resposta

---

**Versão Anterior:** 0.1.0
**Build:** Bem-sucedido (9.8MB)
**Status:** Pronto para uso
