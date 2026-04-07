# =============================================================================
# Skoolz Go REST API - Makefile
# =============================================================================
# Quick reference:
#   make help        - Show all available commands
#   make build       - Compile the binary into ./main
#   make migrate     - Run database migrations
#   make run         - Start the REST API server (foreground)
#   make dev         - Hot-reload dev server (requires `air`)
#   make start       - build + migrate + run (one shot)
#   make stop        - Stop any running server
# =============================================================================

# ----- Variables -------------------------------------------------------------
MAIN        := ./
TARGET      := main
SERVER_CMD  := ./$(TARGET) serve-rest
MIGRATE_CMD := ./$(TARGET) migrate

# Default target
.DEFAULT_GOAL := help

.PHONY: help tidy install-deps install-dev-deps \
        build clean migrate run dev start stop \
        test fmt vet

# ----- Help ------------------------------------------------------------------
help:
	@echo ""
	@echo "Skoolz Go REST API - Available targets"
	@echo "======================================"
	@echo "  make build              - Compile binary into ./$(TARGET)"
	@echo "  make migrate            - Run database migrations"
	@echo "  make run                - Run the REST API server"
	@echo "  make start              - build + migrate + run (one shot)"
	@echo "  make dev                - Hot-reload dev server (needs 'air')"
	@echo "  make stop               - Stop any running server"
	@echo "  make clean              - Remove built binary"
	@echo ""
	@echo "  make tidy               - go mod tidy"
	@echo "  make install-deps       - go mod download"
	@echo "  make install-dev-deps   - install air for hot reload"
	@echo "  make fmt                - gofmt all .go files"
	@echo "  make vet                - go vet ./..."
	@echo "  make test               - go test ./..."
	@echo ""

# ----- Dependency management -------------------------------------------------
tidy:
	go mod tidy

install-deps:
	go mod download

install-dev-deps:
	go install github.com/air-verse/air@latest

# ----- Build / Run -----------------------------------------------------------
build: install-deps
	@echo "==> Building $(TARGET)..."
	go build -o $(TARGET) $(MAIN)
	@echo "==> Built ./$(TARGET)"

clean:
	@rm -f $(TARGET)
	@echo "==> Cleaned $(TARGET)"

migrate: build
	@echo "==> Running migrations..."
	$(MIGRATE_CMD)

run: build
	@echo "==> Starting REST server..."
	$(SERVER_CMD)

start: build migrate run

stop:
	-@pkill -f "serve-rest" 2>/dev/null || true
	@echo "==> Stopped any running serve-rest processes"

dev: install-deps
	air serve-rest

# ----- Quality ---------------------------------------------------------------
fmt:
	gofmt -s -w .

vet:
	go vet ./...

test:
	go test ./...
