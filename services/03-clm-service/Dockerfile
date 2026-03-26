# --- builder stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Copy module files first to leverage Docker layer caching.
COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /clm-service ./cmd/clm

# --- final stage ---
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /clm-service /app/clm-service

# Default environment — override at runtime via docker run -e or Kubernetes env vars.
ENV POSTGRES_DSN="postgres://mcp:mcp_dev_password@localhost:5432/mcp" \
    VAULT_ADDR="http://localhost:8200" \
    VAULT_TOKEN="token" \
    PORT="8080"

EXPOSE 8080

USER nobody

ENTRYPOINT ["/app/clm-service"]
