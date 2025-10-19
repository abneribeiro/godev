# Release Guide - GoDev

Guia completo para criar releases e publicar packages do GoDev.

## 📦 Parte 1: Criar Release no GitHub

### Método Automático via GitHub Actions ⭐ RECOMENDADO

Nosso workflow já está configurado em `.github/workflows/release.yml`!

#### Passo a passo:

```bash
# 1. Certifique-se que tudo está commitado
git status

# 2. Push para o GitHub
git push origin main

# 3. Use o script helper
chmod +x scripts/create-release.sh
./scripts/create-release.sh v0.2.0 "Release v0.2.0 - Full-featured API testing tool"

# OU manualmente:
git tag -a v0.2.0 -m "Release v0.2.0 - Full-featured API testing tool"
git push origin v0.2.0
```

#### O que acontece automaticamente:

1. ✅ GitHub Actions detecta tag `v*`
2. ✅ Compila 5 binários:
   - `godev-linux-amd64`
   - `godev-linux-arm64`
   - `godev-darwin-amd64` (macOS Intel)
   - `godev-darwin-arm64` (macOS Apple Silicon)
   - `godev-windows-amd64.exe`
3. ✅ Gera `checksums.txt` (SHA256)
4. ✅ Cria release automaticamente
5. ✅ Upload de todos os binários

#### Acompanhe:

- URL: `https://github.com/abneribeiro/godev/actions`
- Tempo: ~3-5 minutos

---

## 🚀 Parte 2: Publicar Package Go

**Go packages são automaticamente indexados!** Não precisa fazer nada manual.

### Como funciona:

1. Você cria a tag:
   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```

2. Go Module Proxy detecta automaticamente em segundos

3. Usuários podem instalar:
   ```bash
   go install github.com/abneribeiro/godev@latest
   go install github.com/abneribeiro/godev@v0.2.0
   ```

### Verificar se está publicado:

```bash
# Listar versões
curl https://proxy.golang.org/github.com/abneribeiro/godev/@v/list

# Ver info de versão
curl https://proxy.golang.org/github.com/abneribeiro/godev/@v/v0.2.0.info

# Forçar atualização (se necessário)
GOPROXY=https://proxy.golang.org GO111MODULE=on \
  go list -m github.com/abneribeiro/godev@v0.2.0
```

---

## 🛠️ Scripts Auxiliares

### Build Local (testar antes de release):

```bash
chmod +x scripts/build-all.sh
./scripts/build-all.sh v0.2.0
```

Gera binários em `dist/` para todas as plataformas.

### Criar Release Completo:

```bash
./scripts/create-release.sh v0.2.0 "Release message"
```

Faz: push + tag + push tag automaticamente.

---

## 📋 Checklist de Release

**Antes de criar:**

- [ ] Todos commits em main
- [ ] CHANGELOG.md atualizado
- [ ] README.md versão correta
- [ ] Build local funciona: `go build`
- [ ] `.github/RELEASE_NOTES_vX.X.X.md` criada

**Após criar tag:**

- [ ] GitHub Actions completou ✅
- [ ] 5 binários + checksums gerados
- [ ] Release marcada como "latest"
- [ ] pkg.go.dev indexou
- [ ] `go install` funciona

---

## 🐛 Troubleshooting

### GitHub Actions falhou

1. Veja: `https://github.com/abneribeiro/godev/actions`
2. Corrija o erro
3. Crie nova tag: `v0.2.1`

### Tag errada

```bash
# Deletar localmente
git tag -d v0.2.0

# Deletar no GitHub
git push origin :refs/tags/v0.2.0

# Recriar
git tag -a v0.2.0 -m "Mensagem correta"
git push origin v0.2.0
```

### Go proxy não encontra

```bash
GOPROXY=proxy.golang.org GO111MODULE=on \
  go get github.com/abneribeiro/godev@v0.2.0
```

---

## 📚 Recursos

- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Go Module Proxy](https://proxy.golang.org/)
