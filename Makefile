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
CGO_ENABLED := 0

# ldflags for version injection and size optimization
LDFLAGS := -s -w \
	-X 'github.com/nightmare-assault/nightmare-assault/internal/version.Version=$(VERSION)' \
	-X 'github.com/nightmare-assault/nightmare-assault/internal/version.Commit=$(COMMIT)' \
	-X 'github.com/nightmare-assault/nightmare-assault/internal/version.BuildTime=$(BUILD_TIME)' \
	-X 'github.com/nightmare-assault/nightmare-assault/internal/version.GoVersion=$(GO_VERSION)'

# Build targets
.PHONY: all build build-all clean test help version

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

	@echo "Building Windows amd64..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GO) build -ldflags="$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_PKG)

	@echo "Building macOS amd64 (Intel)..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GO) build -ldflags="$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PKG)

	@echo "Building macOS arm64 (Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) $(GO) build -ldflags="$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_PKG)

	@echo "Building Linux amd64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GO) build -ldflags="$(LDFLAGS)" \
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
	@echo "Targets:"
	@echo "  build      Build for current platform (default)"
	@echo "  build-all  Build for all supported platforms"
	@echo "  clean      Remove build artifacts"
	@echo "  test       Run all tests"
	@echo "  version    Show version info to be embedded"
	@echo "  help       Show this help message"
	@echo ""
	@echo "Output binaries will be placed in $(DIST_DIR)/"
