# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Copy dependency manifests first for better layer caching.
COPY go.mod ./

# Download dependencies (none beyond stdlib in Phase 1).
RUN go mod download

# Copy source and compile a static binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /discovery-engine ./cmd/scanner

# ── Stage 2: runtime ──────────────────────────────────────────────────────────
FROM alpine:3.19

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /discovery-engine ./discovery-engine

USER appuser

# Graph DB and relational DB connection details.
ENV NEO4J_URI=bolt://neo4j:7687 \
    NEO4J_USER=neo4j \
    NEO4J_PASSWORD=password \
    POSTGRES_DSN=postgres://mcp:mcp_dev_password@postgres:5432/mcp \
    PORT=8080

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./discovery-engine"]
