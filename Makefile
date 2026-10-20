BIN     := bin/dws
PKG     := github.com/parmar94/dev-workspace-cli
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS := -s -w -X $(PKG)/internal/cli.version=$(VERSION) -X $(PKG)/internal/cli.commit=$(COMMIT)

.PHONY: build test lint fmt tidy clean docs

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/dws

test:
	go test ./...

lint:
	go vet ./...
	@test -z "$$(gofmt -l .)" || (gofmt -l . && echo "gofmt needed" && exit 1)

fmt:
	gofmt -w .

tidy:
	go mod tidy

docs:
	rm -rf docs/reference
	go run ./cmd/dws gen-docs docs/reference

clean:
	rm -rf bin
