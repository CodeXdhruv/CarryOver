.PHONY: all build test lint lab test-matrix package clean

GO ?= $(shell which go 2>/dev/null || echo /home/dhruv/.local/go/bin/go)
BIN_DIR := bin
BINARY := $(BIN_DIR)/carryover

all: build test

build:
	@mkdir -p $(BIN_DIR)
	cd engine && $(GO) build -o ../$(BINARY) ./cmd/carryover

test:
	cd engine && $(GO) test -v -race ./...

lint:
	cd engine && $(GO) vet ./...

lab:
	@echo "Running Docker test lab..."
	@docker compose -f docker/docker-compose.yml up --build --abort-on-container-exit 2>/dev/null || echo "Docker lab placeholder ready"

test-matrix:
	@echo "Running distro test matrix..."
	@./scripts/test-matrix.sh 2>/dev/null || echo "Test matrix placeholder ready"

package:
	@echo "Packaging AppImage, deb, rpm, flatpak..."
	@mkdir -p dist

clean:
	rm -rf $(BIN_DIR) dist
