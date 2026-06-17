# ---- Stage 1: Build frontend ----
FROM node:22-alpine AS frontend
RUN corepack enable && corepack prepare pnpm@9 --activate
WORKDIR /build/frontend
COPY frontend/ .
RUN echo 'packages: []' > pnpm-workspace.yaml
RUN pnpm install --no-frozen-lockfile
RUN pnpm build
# Output lands in /build/server/static via vite outDir

# ---- Stage 2: Build Go binary ----
FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Bring in the built frontend assets
COPY --from=frontend /build/server/static ./server/static
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X goUp/utils.Version=${VERSION}" -o /goUp .

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
