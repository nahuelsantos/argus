#!/bin/bash
# Update version across all Argus files

set -e

VERSION_FILE="internal/config/VERSION"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

if [ ! -f "$REPO_ROOT/$VERSION_FILE" ]; then
    echo "❌ VERSION file not found at $REPO_ROOT/$VERSION_FILE"
    exit 1
fi

VERSION=$(cat "$REPO_ROOT/$VERSION_FILE" | tr -d '\n\r' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
VERSION_TAG="v$VERSION"

echo -e "${GREEN}🚀 Updating Argus version to $VERSION_TAG${NC}"

cd "$REPO_ROOT"

# Update Dockerfile
if [ -f "Dockerfile" ]; then
    echo -e "${YELLOW}📦 Updating Dockerfile...${NC}"
    sed -i.bak "s/org.opencontainers.image.version=\".*\"/org.opencontainers.image.version=\"$VERSION_TAG\"/" Dockerfile
    rm -f Dockerfile.bak
fi

# Update GitHub Actions
if [ -f ".github/workflows/build-and-publish.yml" ]; then
    echo -e "${YELLOW}🔧 GitHub Actions workflow already uses dynamic versioning${NC}"
fi

# Update README.md
if [ -f "README.md" ]; then
    echo -e "${YELLOW}📚 Updating README.md...${NC}"
    sed -i.bak "s/ghcr.io\/nahuelsantos\/argus:v[0-9]\+\.[0-9]\+\.[0-9]\+/ghcr.io\/nahuelsantos\/argus:$VERSION_TAG/g" README.md
    rm -f README.md.bak
fi

echo -e "${GREEN}✅ Version update complete!${NC}"
echo -e "${GREEN}📝 Summary:${NC}"
echo -e "  ${GREEN}•${NC} VERSION file: $VERSION"
echo -e "  ${GREEN}•${NC} Docker tag: $VERSION_TAG" 
echo -e "  ${GREEN}•${NC} Go build will use: $VERSION_TAG (via ldflags)"
echo -e "  ${GREEN}•${NC} Frontend will get version from API"
echo
echo -e "${YELLOW}💡 Next steps:${NC}"
echo -e "  ${YELLOW}1.${NC} Build: make build"
echo -e "  ${YELLOW}2.${NC} Test: make run"
echo -e "  ${YELLOW}3.${NC} Docker: make docker-build"
echo -e "  ${YELLOW}4.${NC} Release: git tag $VERSION_TAG && git push origin $VERSION_TAG" 