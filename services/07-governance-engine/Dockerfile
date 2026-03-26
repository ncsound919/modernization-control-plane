# ---- Build stage -------------------------------------------------------
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Copy module files first for layer caching.
COPY go.mod ./
RUN go mod download

# Copy source and build a statically linked binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /governance ./cmd/governance

# ---- Final stage --------------------------------------------------------
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /governance /app/governance

# Copy OPA policy files so they are available at runtime for future
# native OPA integration.
COPY policies/ /app/policies/

# Runtime environment variable defaults (override at deploy time).
ENV PORT=8080 \
    POSTGRES_DSN="postgres://mcp:mcp_dev_password@localhost:5432/mcp" \
    KAFKA_BROKERS="localhost:9092"

EXPOSE 8080

ENTRYPOINT ["/app/governance"]
