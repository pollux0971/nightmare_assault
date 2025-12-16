#!/bin/bash
# Package Nightmare Assault for cross-platform release
# Usage: ./scripts/package.sh [version]
#        If version is not specified, uses git tag

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get project root directory (parent of scripts/)
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo -e "${BLUE}=== Nightmare Assault Release Packager ===${NC}\n"

# Determine version
if [ -n "$1" ]; then
    VERSION="$1"
    echo -e "${YELLOW}Using specified version: ${VERSION}${NC}"
else
    # Try to get version from git tag
    VERSION=$(git describe --tags --exact-match 2>/dev/null || git describe --tags 2>/dev/null || echo "dev-$(git rev-parse --short HEAD)")
    echo -e "${GREEN}Detected version from git: ${VERSION}${NC}"
fi

# Validate version format (should start with 'v')
if [[ ! "$VERSION" =~ ^v ]]; then
    echo -e "${YELLOW}Warning: Version doesn't start with 'v', adding prefix...${NC}"
    VERSION="v${VERSION}"
fi

echo -e "${BLUE}Building version: ${VERSION}${NC}\n"

# Step 1: Clean previous builds
echo -e "${BLUE}[1/5] Cleaning previous builds...${NC}"
make clean
echo -e "${GREEN}✓ Clean complete${NC}\n"

# Step 2: Build all platforms
echo -e "${BLUE}[2/5] Building all platforms...${NC}"
make build-all
echo -e "${GREEN}✓ Build complete${NC}\n"

# Step 3: Create release archives
echo -e "${BLUE}[3/5] Creating release archives...${NC}"
cd dist

# Windows (ZIP)
echo "  - Creating Windows archive..."
zip -q nightmare-${VERSION}-windows-amd64.zip nightmare-windows-amd64.exe
echo -e "    ${GREEN}✓ nightmare-${VERSION}-windows-amd64.zip${NC}"

# macOS Intel (TAR.GZ)
echo "  - Creating macOS Intel archive..."
tar czf nightmare-${VERSION}-darwin-amd64.tar.gz nightmare-darwin-amd64
echo -e "    ${GREEN}✓ nightmare-${VERSION}-darwin-amd64.tar.gz${NC}"

# macOS ARM (TAR.GZ)
echo "  - Creating macOS ARM archive..."
tar czf nightmare-${VERSION}-darwin-arm64.tar.gz nightmare-darwin-arm64
echo -e "    ${GREEN}✓ nightmare-${VERSION}-darwin-arm64.tar.gz${NC}"

# Linux (TAR.GZ)
echo "  - Creating Linux archive..."
tar czf nightmare-${VERSION}-linux-amd64.tar.gz nightmare-linux-amd64
echo -e "    ${GREEN}✓ nightmare-${VERSION}-linux-amd64.tar.gz${NC}"

echo -e "${GREEN}✓ Archives created${NC}\n"

# Step 4: Generate checksums
echo -e "${BLUE}[4/5] Generating SHA256 checksums...${NC}"
sha256sum *.zip *.tar.gz > checksums-${VERSION}.txt
echo -e "${GREEN}✓ Checksums saved to checksums-${VERSION}.txt${NC}\n"

# Step 5: Display summary
echo -e "${BLUE}[5/5] Package Summary${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""
echo -e "${GREEN}Release archives:${NC}"
ls -lh nightmare-${VERSION}-*.zip nightmare-${VERSION}-*.tar.gz 2>/dev/null | awk '{print "  " $9 " (" $5 ")"}'
echo ""
echo -e "${GREEN}Checksums:${NC}"
cat checksums-${VERSION}.txt | sed 's/^/  /'
echo ""
echo -e "${BLUE}======================================${NC}"
echo -e "${GREEN}✓ Packaging complete!${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Test binaries on target platforms"
echo "  2. Create GitHub Release: gh release create ${VERSION}"
echo "  3. Upload archives: gh release upload ${VERSION} dist/nightmare-${VERSION}-* dist/checksums-${VERSION}.txt"
echo "  4. Push tag to remote: git push origin ${VERSION}"
echo ""
echo -e "${BLUE}Distribution files ready in: ${PROJECT_ROOT}/dist/${NC}"
