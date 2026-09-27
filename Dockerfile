# ---- Stage 1: Build frontend (native, output is arch-independent) ----
FROM --platform=$BUILDPLATFORM node:22-alpine AS frontend
RUN corepack enable && corepack prepare pnpm@9 --activate
WORKDIR /build/frontend
COPY frontend/ .
RUN echo 'packages: []' > pnpm-workspace.yaml
RUN pnpm install --no-frozen-lockfile
RUN pnpm build

# ---- Stage 2: Build Go binary (native toolchain, cross-compiled output) ----
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /build
RUN GOTOOLCHAIN=auto go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /build/server/static ./server/static
RUN cd db && sqlc generate
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-s -w -X goUp/utils.Version=${VERSION}" -o /goUp .


# ---- Stage 3: Minimal runtime ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /goUp /app/goUp
RUN mkdir -p /app/config /app/data
VOLUME ["/app/config"]
VOLUME ["/app/data"]
EXPOSE 8101
ENTRYPOINT ["/app/goUp"]
CMD ["-config", "/app/config/services.yml"]
