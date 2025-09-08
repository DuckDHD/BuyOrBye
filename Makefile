# GoTTH Stack Makefile - Cross-platform (Unix/Windows)

# Detect the operating system
ifeq ($(OS),Windows_NT)
    detected_OS := Windows
    EXE_EXT := .exe
    RM := del /Q
    RMDIR := rmdir /S /Q
    MKDIR := mkdir
    TAILWIND_URL := https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-windows-x64.exe
    TAILWIND_BIN := tailwindcss.exe
else
    detected_OS := $(shell uname -s)
    EXE_EXT :=
    RM := rm -f
    RMDIR := rm -rf
    MKDIR := mkdir -p
    ifeq ($(detected_OS),Darwin)
        # macOS
        UNAME_M := $(shell uname -m)
        ifeq ($(UNAME_M),arm64)
            TAILWIND_URL := https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
        else
            TAILWIND_URL := https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-x64
        endif
    else
        # Linux
        TAILWIND_URL := https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
    endif
    TAILWIND_BIN := tailwindcss
endif
all: build test

# Install templ if not present
templ-install:
	@if ! command -v templ >/dev/null 2>&1; then \
		echo "Installing templ..."; \
		go install github.com/a-h/templ/cmd/templ@latest; \
		if ! command -v templ >/dev/null 2>&1; then \
			echo "templ installation failed. Make sure $$GOPATH/bin is in your PATH"; \
			exit 1; \
		else \
			echo "templ installed successfully"; \
		fi; \
	else \
		echo "templ already installed"; \
	fi

# Install TailwindCSS standalone executable
tailwind-install:
	@if [ ! -f "$(TAILWIND_BIN)" ]; then \
		echo "Installing TailwindCSS standalone for $(detected_OS)..."; \
		if command -v curl >/dev/null 2>&1; then \
			curl -sLO $(TAILWIND_URL); \
		elif command -v wget >/dev/null 2>&1; then \
			wget -q $(TAILWIND_URL); \
		else \
			echo "Error: curl or wget required to download TailwindCSS"; \
			exit 1; \
		fi; \
		if [ "$(detected_OS)" != "Windows" ]; then \
			chmod +x $(TAILWIND_BIN); \
		fi; \
		echo "TailwindCSS installed successfully"; \
	else \
		echo "TailwindCSS already installed"; \
	fi

# Install Air for live reloading
air-install:
	@if ! command -v air >/dev/null 2>&1; then \
		echo "Installing air..."; \
		go install github.com/air-verse/air@latest; \
		if ! command -v air >/dev/null 2>&1; then \
			echo "air installation failed. Make sure $$GOPATH/bin is in your PATH"; \
			exit 1; \
		else \
			echo "air installed successfully"; \
		fi; \
	else \
		echo "air already installed"; \
	fi

# Build the application
build: tailwind-install templ-install
	@echo "Building..."
	@templ generate
	@./$(TAILWIND_BIN) -i cmd/web/styles/input.css -o cmd/web/assets/css/output.css
	@go build -o main$(EXE_EXT) cmd/api/main.go

# Build for production with optimizations
build-prod: tailwind-install templ-install
	@echo "Building for production..."
	@templ generate
	@./$(TAILWIND_BIN) -i cmd/web/styles/input.css -o cmd/web/assets/css/output.css --minify
	@go build -ldflags="-s -w" -o main$(EXE_EXT) cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Create DB container
docker-run:
	@docker compose up --build

# Shutdown DB container
docker-down:
	@docker compose down

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Integration Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary and generated files
clean:
	@echo "Cleaning..."
	@$(RM) main$(EXE_EXT) 2>/dev/null || true
	@$(RM) cmd/web/assets/css/output.css 2>/dev/null || true
	@find . -name "*_templ.go" -type f -delete 2>/dev/null || true
	@if [ -d "tmp" ]; then $(RMDIR) tmp; fi

# Live Reload with Air
watch: air-install
	@echo "Starting live reload with Air..."
	@air

# Development mode with concurrent processes
dev: tailwind-install templ-install air-install
	@echo "Starting development environment..."
	@echo "This will start Air for live reloading..."
	@air

# Update all tools to latest versions
update-tools:
	@echo "Updating tools to latest versions..."
	@go install github.com/a-h/templ/cmd/templ@latest
	@go install github.com/air-verse/air@latest
	@echo "Tools updated successfully!"

# Check if all required tools are installed
check-tools:
	@echo "Checking required tools..."
	@tools="go templ air"; \
	missing=""; \
	for tool in $$tools; do \
		if command -v $$tool >/dev/null 2>&1; then \
			echo "✓ $$tool is installed"; \
		else \
			missing="$$missing $$tool"; \
		fi; \
	done; \
	if [ -n "$$missing" ]; then \
		echo "❌ Missing tools:$$missing"; \
		echo "Run 'make install-tools' to install missing tools."; \
	else \
		echo "✅ All required tools are installed!"; \
	fi

# Install all required tools
install-tools: templ-install tailwind-install air-install
	@echo "All tools installed successfully!"

# Initialize Air configuration
air-init:
	@if [ ! -f ".air.toml" ]; then \
		echo "Creating .air.toml configuration..."; \
		air init; \
		echo ".air.toml created successfully"; \
	else \
		echo ".air.toml already exists"; \
	fi

# Show system information
info:
	@echo "System Information:"
	@echo "Detected OS: $(detected_OS)"
	@echo "TailwindCSS URL: $(TAILWIND_URL)"
	@echo "TailwindCSS Binary: $(TAILWIND_BIN)"
	@echo "Go version: $(go version)"
	@if [ "$(detected_OS)" = "WSL" ]; then \
		echo "WSL Environment detected - using Linux binaries"; \
	fi

.PHONY: all build build-prod run test clean watch dev itest templ-install tailwind-install air-install docker-run docker-down update-tools check-tools install-tools air-init info