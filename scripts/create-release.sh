#!/bin/bash

# Script para criar release do GoDev
# Uso: ./scripts/create-release.sh v0.2.0 "Release message"

set -e

VERSION=$1
MESSAGE=$2

if [ -z "$VERSION" ] || [ -z "$MESSAGE" ]; then
    echo "Uso: $0 <version> <message>"
    echo "Exemplo: $0 v0.2.0 'Release v0.2.0 - Full-featured API testing tool'"
    exit 1
fi

echo "🏷️  Criando release $VERSION..."

# Verificar se há mudanças não commitadas
if [ -n "$(git status --porcelain)" ]; then
    echo "❌ Erro: Há mudanças não commitadas. Commit primeiro!"
    git status --short
    exit 1
fi

# Verificar se já existe a tag
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo "❌ Erro: Tag $VERSION já existe!"
    exit 1
fi

# Push para main
echo "📤 Fazendo push para origin/main..."
git push origin main

# Criar tag anotada
echo "🏷️  Criando tag $VERSION..."
git tag -a "$VERSION" -m "$MESSAGE"

# Push da tag (dispara GitHub Actions)
echo "🚀 Fazendo push da tag $VERSION..."
git push origin "$VERSION"

echo ""
echo "✅ Release $VERSION criada com sucesso!"
echo ""
echo "📋 Próximos passos:"
echo "   1. Acompanhe o build em: https://github.com/abneribeiro/godev/actions"
echo "   2. Aguarde 3-5 minutos para os binários serem gerados"
echo "   3. Edite a release em: https://github.com/abneribeiro/godev/releases/tag/$VERSION"
echo "   4. Adicione as release notes de: .github/RELEASE_NOTES_v0.2.0.md"
echo ""
