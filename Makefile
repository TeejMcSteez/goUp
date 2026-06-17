SWAG   := $(shell go env GOPATH)/bin/swag
BINARY := goUp
VERSION := $(shell git describe --tags --always --dirty)

# Detect package manager: prefer the one matching the existing lock file,
# then fall back to whatever is installed, then plain npm.
PM := $(shell \
  if [ -f frontend/pnpm-lock.yaml ] && command -v pnpm >/dev/null 2>&1; then echo pnpm; \
  elif [ -f frontend/bun.lockb ]     && command -v bun  >/dev/null 2>&1; then echo bun;  \
  elif [ -f frontend/yarn.lock ]     && command -v yarn >/dev/null 2>&1; then echo yarn; \
  elif command -v pnpm >/dev/null 2>&1; then echo pnpm; \
  elif command -v yarn >/dev/null 2>&1; then echo yarn; \
  elif command -v bun  >/dev/null 2>&1; then echo bun;  \
  else echo npm; fi)

.PHONY: all docs build fmt lint test clean prof dev

all: docs lint fmt test build

dev:
	cd frontend && $(PM) run build
	go run .

docs:
	$(SWAG) init -g main.go -o docs

build:
	cd frontend && $(PM) run build
	go build -ldflags "-X goUp/utils.Version=$(VERSION)" -o $(BINARY) .

fmt:
	golangci-lint fmt ./...

lint:
	cd frontend && $(PM) lint && cd ../
	golangci-lint run ./...

test:
	go test ./... -cover

prof:
	mkdir -p perf
	go test -benchmem -cpuprofile=perf/server.cpu.out -memprofile=perf/server.mem.out ./server
	go test -benchmem -cpuprofile=perf/utils.cpu.out  -memprofile=perf/utils.mem.out  ./utils
	go test -benchmem -cpuprofile=perf/workers.cpu.out -memprofile=perf/workers.mem.out ./workers

clean:
	rm -f $(BINARY)
	rm -rf perf
	rm -f server.test utils.test workers.test
