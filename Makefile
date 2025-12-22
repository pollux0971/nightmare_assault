# Makefile for Nightmare Assault
# Cross-platform build automation

# Application info
APP_NAME := nightmare
MAIN_PKG := ./cmd/nightmare

# Version info (extracted at build time)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0-dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%d %H:%M:%S")
GO_VERSION := $(shell go version | awk '{print $$3}')

# Build directories
DIST_DIR := dist

# Go build settings
GO := go
# CGO is required for Linux/Unix audio support (ALSA)
# macOS and Windows use purego (no CGO required)
CGO_ENABLED := 1

# ldflags for version injection and size optimization
LDFLAGS := -s -w \
	-X 'github.com/nightmare-assault/nightmare-assault/internal/version.Version=$(VERSION)' \
	-X 'github.com/nightmare-assault/nightmare-assault/internal/version.Commit=$(COMMIT)' \
	-X 'github.com/nightmare-assault/nightmare-assault/internal/version.BuildTime=$(BUILD_TIME)' \
	-X 'github.com/nightmare-assault/nightmare-assault/internal/version.GoVersion=$(GO_VERSION)'

# Build targets
.PHONY: all build build-all clean test test-epic2 coverage-epic2 stress-test stress-test-quick stress-test-memory stress-test-npc stress-test-persistence help version

# Default target
all: build

# Build for current platform
build:
	@echo "Building for current platform..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME) $(MAIN_PKG)
	@echo "Build complete: $(DIST_DIR)/$(APP_NAME)"
	@ls -lh $(DIST_DIR)/$(APP_NAME)

# Build for all platforms
build-all: clean
	@echo "Building for all platforms..."
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(GO_VERSION)"
	@echo ""
	@mkdir -p $(DIST_DIR)

	@echo "Building Windows amd64 (CGO disabled - uses purego)..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_PKG)

	@echo "Building macOS amd64 (Intel) (CGO disabled - uses purego)..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PKG)

	@echo "Building macOS arm64 (Apple Silicon) (CGO disabled - uses purego)..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_PKG)

	@echo "Building Linux amd64 (CGO enabled - requires ALSA)..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 $(GO) build -ldflags="$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PKG)

	@echo ""
	@echo "Build complete! Binaries in $(DIST_DIR)/:"
	@ls -lh $(DIST_DIR)/

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(DIST_DIR)
	@echo "Clean complete."

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run Epic 2 tests with coverage and race detection
test-epic2:
	@echo "Running Epic 2 (Knowledge System) tests..."
	$(GO) test -v -race -coverprofile=coverage-epic2.out ./internal/npc/knowledge/...
	@echo ""
	@echo "Coverage summary:"
	$(GO) tool cover -func=coverage-epic2.out | grep total

# Generate Epic 2 coverage HTML report
coverage-epic2:
	@echo "Generating Epic 2 coverage report..."
	$(GO) test -coverprofile=coverage-epic2.out ./internal/npc/knowledge/...
	$(GO) tool cover -html=coverage-epic2.out -o coverage-epic2.html
	@echo "✅ Coverage report generated: coverage-epic2.html"

# Run stress tests (Story 8.8)
stress-test:
	@echo "Running stress tests..."
	@echo "⚠️  Note: Some tests may take several minutes"
	$(GO) test -v -timeout 30m ./test/stress/...

# Run quick stress tests (short mode)
stress-test-quick:
	@echo "Running quick stress tests (short mode)..."
	$(GO) test -v -short -timeout 5m ./test/stress/...

# Run memory stability tests
stress-test-memory:
	@echo "Running memory stability tests..."
	$(GO) test -v -run "TestMemory" -timeout 10m ./test/stress/...

# Run NPC dialogue load tests
stress-test-npc:
	@echo "Running NPC dialogue load tests..."
	$(GO) test -v -run "TestNPC" -timeout 10m ./test/stress/...

# Run state persistence tests
stress-test-persistence:
	@echo "Running state persistence tests..."
	$(GO) test -v -run "TestStatePersistence" -timeout 10m ./test/stress/...

# Show version info that will be embedded
version:
	@echo "Version Info:"
	@echo "  Version:    $(VERSION)"
	@echo "  Commit:     $(COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Go Version: $(GO_VERSION)"

# Help
help:
	@echo "Nightmare Assault Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build Targets:"
	@echo "  build                   Build for current platform (default)"
	@echo "  build-all               Build for all supported platforms"
	@echo "  clean                   Remove build artifacts"
	@echo ""
	@echo "Test Targets:"
	@echo "  test                    Run all tests"
	@echo "  test-epic2              Run Epic 2 tests with coverage and race detection"
	@echo "  coverage-epic2          Generate Epic 2 HTML coverage report"
	@echo ""
	@echo "Stress Test Targets (Story 8.8):"
	@echo "  stress-test             Run all stress tests (may take 30+ minutes)"
	@echo "  stress-test-quick       Run quick stress tests (short mode, ~5 minutes)"
	@echo "  stress-test-memory      Run memory stability tests"
	@echo "  stress-test-npc         Run NPC dialogue load tests"
	@echo "  stress-test-persistence Run state persistence tests"
	@echo ""
	@echo "Other Targets:"
	@echo "  version                 Show version info to be embedded"
	@echo "  help                    Show this help message"
	@echo ""
	@echo "Output binaries will be placed in $(DIST_DIR)/"
