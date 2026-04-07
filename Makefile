# =============================================================================
# Skoolz Go API - Makefile
# =============================================================================
# Quick reference (most common):
#   make help        - Show all available commands
#   make build       - Compile the binary into ./main
#   make migrate     - Run database migrations
#   make run         - Start the REST API server (foreground)
#   make dev         - Hot-reload dev server (requires `air`)
#   make start       - build + migrate + run (one shot)
# =============================================================================

# ----- Variables -------------------------------------------------------------
MAIN          := ./
TARGET        := main
SERVER_CMD    := ./$(TARGET) serve-rest
MIGRATE_CMD   := ./$(TARGET) migrate
GRPC_CMD      := ./$(TARGET) serve-grpc

# Protobuf
PROTOC_DEST   := ./
PROTOC_FLAGS  := --go_out=$(PROTOC_DEST) --go_opt=paths=source_relative \
                 --go-grpc_out=$(PROTOC_DEST) --go-grpc_opt=paths=source_relative
PROTO_FILES   := $(shell find ./proto -name "*.proto" -type f 2>/dev/null)

# Default target shown when running just `make`
.DEFAULT_GOAL := help

# Phony targets (not real files)
.PHONY: help tidy install-deps install-proto-deps install-dev-deps install-mockgen \
        build build-grpc clean clean-proto build-proto proto list-proto \
        migrate run run-grpc dev start stop \
        test fmt vet lint \
        infra-up infra-down infra-logs

# ----- Help ------------------------------------------------------------------
help:
	@echo ""
	@echo "Skoolz Go API - Available targets"
	@echo "================================="
	@echo "  make build              - Compile binary into ./$(TARGET)"
	@echo "  make migrate            - Run database migrations"
	@echo "  make run                - Run the REST API server"
	@echo "  make start              - build + migrate + run (one shot)"
	@echo "  make dev                - Hot-reload dev server (needs 'air')"
	@echo "  make stop               - Stop any running ./$(TARGET) processes"
	@echo ""
	@echo "  make tidy               - go mod tidy"
	@echo "  make install-deps       - go mod download"
	@echo "  make fmt                - gofmt all .go files"
	@echo "  make vet                - go vet ./..."
	@echo "  make test               - go test ./..."
	@echo "  make clean              - Remove built binary"
	@echo ""
	@echo "  make build-grpc         - Build with gRPC support (needs protos generated first)"
	@echo "  make proto              - (Re)generate protobuf files (needs protoc + plugins)"
	@echo "  make install-proto-deps - Install protoc-gen-go + protoc-gen-go-grpc"
	@echo "  make list-proto         - List all .proto files"
	@echo ""
	@echo "  make infra-up           - Start Kafka/NATS/etc via docker-compose"
	@echo "  make infra-down         - Stop infrastructure services"
	@echo "  make infra-logs         - Tail infrastructure logs"
	@echo ""

# ----- Dependency management -------------------------------------------------
tidy:
	go mod tidy

install-deps:
	go mod download

install-proto-deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

install-dev-deps:
	go install github.com/air-verse/air@latest

install-mockgen:
	go install go.uber.org/mock/mockgen@latest

# ----- Build / Run (REST only, no proto required) ---------------------------
build: install-deps
	@echo "==> Building $(TARGET) (REST only)..."
	go build -o $(TARGET) $(MAIN)
	@echo "==> Built ./$(TARGET)"

build-grpc: install-deps
	@echo "==> Building $(TARGET) with gRPC tag..."
	go build -tags grpc -o $(TARGET) $(MAIN)

clean:
	@rm -f $(TARGET)
	@echo "==> Cleaned $(TARGET)"

migrate: build
	@echo "==> Running migrations..."
	$(MIGRATE_CMD)

run: build
	@echo "==> Starting REST server..."
	$(SERVER_CMD)

run-grpc: build-grpc
	$(GRPC_CMD)

start: build migrate run

stop:
	-@pkill -f "serve-rest" 2>/dev/null || true
	-@pkill -f "serve-grpc" 2>/dev/null || true
	@echo "==> Stopped any running serve-rest / serve-grpc processes"

dev: install-deps
	air serve-rest

# ----- Quality ---------------------------------------------------------------
fmt:
	gofmt -s -w .

vet:
	go vet ./...

test:
	go test ./...

# ----- Protobuf (optional, requires protoc) ----------------------------------
build-proto:
	@echo "==> Generating protobuf files..."
	protoc $(PROTOC_FLAGS) $(PROTO_FILES)
	@echo "==> Protobuf generation completed"

clean-proto:
	@echo "==> Cleaning generated protobuf files..."
	find ./proto -name "*.pb.go" -type f -delete
	@echo "==> Clean completed"

proto: clean-proto build-proto

list-proto:
	@echo "Found proto files:"
	@find ./proto -name "*.proto" -type f | sort

# ----- Infrastructure (Docker Compose) ---------------------------------------
infra-up:
	@echo "==> Starting infrastructure services..."
	cd ../../infra && docker-compose up -d

infra-down:
	@echo "==> Stopping infrastructure services..."
	cd ../../infra && docker-compose down

infra-logs:
	@echo "==> Showing infrastructure logs..."
	cd ../../infra && docker-compose logs -f
