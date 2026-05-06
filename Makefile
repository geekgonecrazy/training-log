.PHONY: all proto web build dev test clean tidy seed

GO ?= go
BIN_DIR ?= bin
SERVER_BIN := $(BIN_DIR)/server
SEED_BIN := $(BIN_DIR)/seed-user

all: build

proto:
	buf generate

tidy:
	$(GO) mod tidy

web:
	cd web && npm install && npm run build
	rm -rf webfs/dist
	mkdir -p webfs/dist
	cp -R web/build/. webfs/dist/

build: proto web
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(SERVER_BIN) ./cmd/server
	$(GO) build -o $(SEED_BIN) ./cmd/seed-user

server: proto
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(SERVER_BIN) ./cmd/server

seed: proto
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(SEED_BIN) ./cmd/seed-user

# Run backend + frontend dev servers together. Backend on :8080, vite on :5173 with /v1 proxy.
dev:
	@echo "Run in two terminals:"
	@echo "  1) $(GO) run ./cmd/server"
	@echo "  2) cd web && npm run dev"

test:
	$(GO) test ./...

clean:
	rm -rf $(BIN_DIR) web/build web/.svelte-kit
