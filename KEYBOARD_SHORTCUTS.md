# Atalhos de Teclado - DevScope

## 🎯 Navegação Global

| Atalho | Ação |
|--------|------|
| `Tab` | Próximo campo |
| `Shift+Tab` | Campo anterior |
| `Ctrl+Q` | Sair da aplicação |
| `Ctrl+?` | Mostrar ajuda |
| `Esc` | Voltar/Cancelar |

---

## ✏️ Edição no Campo URL (quando focado)

### Clipboard
| Atalho | Ação |
|--------|------|
| `Ctrl+V` | Colar |
| `Ctrl+C` | Copiar texto selecionado |
| `Ctrl+X` | Cortar texto selecionado |
| `Ctrl+A` | Selecionar tudo |

### Navegação no Texto
| Atalho | Ação |
|--------|------|
| `←` / `→` | Mover cursor |
| `Home` / `Ctrl+A` | Ir para início |
| `End` / `Ctrl+E` | Ir para final |
| `Ctrl+←` | Palavra anterior |
| `Ctrl+→` | Próxima palavra |

### Edição
| Atalho | Ação |
|--------|------|
| `Backspace` | Deletar caractere anterior |
| `Delete` | Deletar caractere atual |
| `Ctrl+U` | Deletar do cursor até início |
| `Ctrl+K` | Deletar do cursor até final |
| `Ctrl+W` | Deletar palavra anterior |

---

## 🔧 Seletor de Método HTTP

| Atalho | Ação |
|--------|------|
| `←` / `h` | Método anterior (GET ← POST ← PUT...) |
| `→` / `l` | Próximo método (GET → POST → PUT...) |

**Métodos disponíveis:** GET → POST → PUT → DELETE → PATCH

---

## 📤 Request Builder

| Atalho | Ação |
|--------|------|
| `Enter` | Enviar request (quando no botão Send) |
| `Ctrl+L` | Abrir lista de requests salvos |

---

## 📥 Visualização de Resposta

| Atalho | Ação |
|--------|------|
| `Esc` | Voltar para Request Builder |
| `s` | Salvar request atual |
| `c` | Copiar resposta (futuro) |
| `↑` / `k` | Scroll para cima |
| `↓` / `j` | Scroll para baixo |

---

## 📋 Lista de Requests Salvos

| Atalho | Ação |
|--------|------|
| `↑` / `k` | Item anterior |
| `↓` / `j` | Próximo item |
| `Enter` | Carregar request selecionado |
| `d` | Deletar request selecionado |
| `n` | Criar novo request |
| `Esc` | Voltar |

---

## 💡 Dicas de Uso

### Workflow Rápido
1. **Colar URL:** `Ctrl+V` no campo URL
2. **Mudar método:** `→` ou `←` para navegar
3. **Enviar:** `Tab` até "Send Request" + `Enter`
4. **Salvar:** Pressione `s` na tela de resposta
5. **Reusar:** `Ctrl+L` → selecione → `Enter`

### Edição Eficiente
- **Limpar campo:** `Ctrl+A` (selecionar) + `Delete`
- **Corrigir parte da URL:** `Ctrl+←/→` para navegar por palavras
- **Deletar domínio:** `Home` → `Ctrl+K`

### Navegação Rápida
- **Vim users:** Use `h`, `j`, `k`, `l` onde aplicável
- **Entre campos:** Sempre `Tab` (não precisa usar mouse)
- **Voltar rapidamente:** `Esc` funciona em qualquer tela

---

## 🚨 Comportamento Especial do Ctrl+C

O `Ctrl+C` tem comportamento inteligente:

- **No campo URL com texto:** Copia o texto selecionado
- **No campo URL vazio:** Sai da aplicação
- **Fora do campo URL:** Sai da aplicação

Isso permite usar `Ctrl+C` para copiar URLs sem sair acidentalmente do programa!

---

## 📝 Limitações Atuais

- **Sem Ctrl+Z (Undo):** O textinput não suporta nativamente
- **Sem multi-linha:** Campo URL é single-line
- **Sem mouse:** TUI só funciona com teclado

---

**Dica Pro:** Imprima este arquivo ou mantenha-o aberto em outro terminal enquanto aprende os atalhos! 🚀
