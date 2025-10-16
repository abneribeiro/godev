# DevScope

Uma aplicação TUI (Text-based User Interface) para inspeção e teste de APIs HTTP diretamente do terminal.

![Version](https://img.shields.io/badge/version-0.1.0-orange)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)
![License](https://img.shields.io/badge/license-MIT-green)

## Sobre

**DevScope** é uma alternativa leve e focada ao Postman/Insomnia para desenvolvedores que preferem o ambiente terminal. Com ele você pode:

- Fazer requests HTTP (GET, POST, PUT, DELETE, PATCH)
- Visualizar respostas JSON formatadas com syntax highlighting
- Salvar requests para reuso
- Navegar rapidamente entre requests salvos
- Tudo sem sair do terminal

## Instalação

### Via Go Install

```bash
go install github.com/abneribeiro/devscope@latest
```

### Build Manual

```bash
git clone https://github.com/abneribeiro/devscope
cd devscope
go build -o devscope
./devscope
```

## Uso Rápido

1. Execute o programa:
```bash
devscope
```

2. Digite a URL da API que deseja testar

3. Use `Tab` para navegar entre os campos

4. Pressione `Enter` para enviar o request

5. Visualize a resposta formatada

6. Pressione `s` para salvar o request

7. Use `Ctrl+L` para carregar requests salvos

## Atalhos de Teclado

### Principais
- `Ctrl+Q` - Sair da aplicação
- `Ctrl+V` - Colar URL
- `Ctrl+C` - Copiar (ou sair se campo vazio)
- `Tab` - Próximo campo
- `Enter` - Enviar request
- `←/→` - Mudar método HTTP
- `Ctrl+L` - Carregar requests salvos

> 📖 **[Ver todos os atalhos](KEYBOARD_SHORTCUTS.md)** - Lista completa com clipboard, edição e navegação

## Exemplos

### Teste Rápido com JSONPlaceholder

1. Abra o DevScope
2. URL: `https://jsonplaceholder.typicode.com/users`
3. Método: `GET` (padrão)
4. Pressione `Enter`
5. Visualize a lista de usuários

### POST com JSON Body

1. URL: `https://httpbin.org/post`
2. Método: `POST` (use `←/→` para mudar)
3. Tab até Headers → (feature em desenvolvimento)
4. Tab até Body → (feature em desenvolvimento)
5. Adicione: `{"name": "John", "email": "john@example.com"}`
6. Pressione `Enter`

### Reutilizar Request Salvo

1. Pressione `Ctrl+L`
2. Navegue com `↑/↓`
3. Pressione `Enter` para carregar
4. Pressione `Enter` novamente para executar

## Configuração

Os requests salvos ficam armazenados em:
```
~/.devscope/config.json
```

Você pode editar este arquivo manualmente se necessário.

## Estrutura do Projeto

```
devscope/
├── main.go                    # Entry point
├── internal/
│   ├── ui/
│   │   ├── model.go           # State machine do Bubbletea
│   │   └── styles.go          # Design system (Lipgloss)
│   ├── http/
│   │   └── client.go          # Cliente HTTP
│   └── storage/
│       └── requests.go        # Persistência de requests
├── go.mod
├── go.sum
├── README.md
└── DEVSCOPE.md               # Especificação completa
```

## Tecnologias

- [Go](https://golang.org/) - Linguagem
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Framework TUI
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Estilização
- [Bubbles](https://github.com/charmbracelet/bubbles) - Componentes TUI

## Roadmap

### v0.1.0 (MVP) - Atual
- [x] Request Builder básico
- [x] Suporte a métodos HTTP (GET, POST, PUT, DELETE, PATCH)
- [x] Visualização de resposta formatada
- [x] Persistência de requests
- [x] Lista de requests salvos
- [ ] Editor de headers
- [ ] Editor de body JSON

### v0.2.0
- [ ] Query params editor
- [ ] Response headers viewer
- [ ] Export request para cURL
- [ ] Import cURL para request
- [ ] Search/filter na lista

### v0.3.0
- [ ] Environment variables
- [ ] Request chaining
- [ ] GraphQL support
- [ ] WebSocket inspector

### v1.0.0
- [ ] Database explorer
- [ ] Metrics dashboard
- [ ] Performance benchmarking
- [ ] Plugin system

## Contribuindo

Contribuições são bem-vindas! Por favor:

1. Fork o repositório
2. Crie uma branch: `git checkout -b feat/nova-feature`
3. Commit: `git commit -m "feat: adiciona nova feature"`
4. Push: `git push origin feat/nova-feature`
5. Abra um Pull Request

### Padrões de Commit

- `feat:` - Nova funcionalidade
- `fix:` - Correção de bug
- `refactor:` - Refatoração de código
- `docs:` - Documentação
- `test:` - Testes
- `style:` - Formatação

## Troubleshooting

### Terminal não renderiza cores

Verifique se seu terminal suporta true color:
```bash
echo $COLORTERM  # deve retornar "truecolor" ou "24bit"
```

### Config não é salvo

Verifique as permissões:
```bash
mkdir -p ~/.devscope
chmod 755 ~/.devscope
```

### Erro ao compilar

Certifique-se de ter Go 1.21+ instalado:
```bash
go version
```

## Licença

MIT License - Veja [LICENSE](LICENSE) para detalhes

## Autor

Abner Ribeiro - [GitHub](https://github.com/abneribeiro)

## Agradecimentos

- [Charm](https://charm.sh/) - Pela incrível stack de ferramentas TUI
- Comunidade Go - Pelo suporte e recursos

---

**Feito com ❤️ e Go**
