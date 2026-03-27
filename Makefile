# Go GitHub Readme Stats Makefile

APP_NAME=github-readme-stats
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR=build
ASSETS_DIR=assets

LDFLAGS=-ldflags "-s -w"

PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

.PHONY: all build build-all clean test run dev help

all: clean build

help:
	@echo "Available targets:"
	@echo "  build      - Build for current platform"
	@echo "  build-all  - Build for all platforms"
	@echo "  dev        - Run in debug mode (no cache)"
	@echo "  run        - Run normally"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build directory"

build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) .
	@cp -r $(ASSETS_DIR) $(BUILD_DIR)/ 2>/dev/null || true
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

build-all: clean
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1); \
		GOARCH=$$(echo $$platform | cut -d/ -f2); \
		OUTPUT=$(BUILD_DIR)/$(APP_NAME)-$$GOOS-$$GOARCH; \
		if [ "$$GOOS" = "windows" ]; then OUTPUT="$$OUTPUT.exe"; fi; \
		echo "Building $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build $(LDFLAGS) -o $$OUTPUT .; \
	done
	@cp -r $(ASSETS_DIR) $(BUILD_DIR)/ 2>/dev/null || true
	@echo "All builds complete!"

dev:
	@echo "Running in debug mode (cache disabled)..."
	DEBUG=true go run .

run:
	go run .

test:
	go test -v ./...

clean:
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete!"
