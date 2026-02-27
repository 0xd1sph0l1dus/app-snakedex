# ── Stage 1 : build ───────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Download dependencies first (layer cached unless go.mod/go.sum change).
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# modernc.org/sqlite is pure Go — CGO_ENABLED=0 produces a static binary.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /snakedex ./cmd/server

# ── Stage 2 : runtime ─────────────────────────────────────────────
FROM alpine:3.21

RUN addgroup -S snakedex && adduser -S -G snakedex snakedex

WORKDIR /app

# Copy binary and assets.
COPY --from=builder /snakedex ./
COPY templates/ ./templates/
COPY static/    ./static/

# Persist the SQLite database in a named volume.
RUN mkdir -p /data && chown snakedex:snakedex /data
VOLUME ["/data"]

USER snakedex

ENV PORT=8080
ENV DB_PATH=/data/snakedex.db

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1

CMD ["./snakedex"]
