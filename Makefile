SWAG   := $(shell go env GOPATH)/bin/swag
BINARY := goUp

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

.PHONY: all docs build fmt lint test clean

all: docs build

docs:
	$(SWAG) init -g main.go -o docs

build:
	cd frontend && $(PM) run build
	go build -o $(BINARY) .

fmt:
	golangci-lint fmt ./...

lint:
	golangci-lint run ./...

test:
	go test ./... -cover

clean:
	rm -f $(BINARY)
	rm -rf docs
