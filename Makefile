APP       := inkwell
MODULE    := github.com/eltaline/inkwell
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
  -X $(MODULE)/internal/version.Version=$(VERSION) \
  -X $(MODULE)/internal/version.Commit=$(COMMIT) \
  -X $(MODULE)/internal/version.BuildDate=$(BUILD_DATE)

GO       := go
GOFLAGS  :=
BIN_DIR  := bin

.PHONY: all build test lint clean

all: test build

build:
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BIN_DIR)/$(APP) ./cmd/inkwell

test:
	$(GO) test ./... -race -count=1

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BIN_DIR)
