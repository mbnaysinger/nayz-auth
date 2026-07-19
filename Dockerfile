# ---------- Build ----------
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /nayz-auth ./cmd/nayz-auth

# ---------- Runtime ----------
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /nayz-auth ./nayz-auth
# Migrations aplicadas automaticamente no start (golang-migrate, file://db/migrations)
COPY --from=builder /app/db/migrations ./db/migrations

USER app

# Configuração via ambiente:
# DATABASE_URL (obrigatória) | JWT_SECRET (obrigatória)
# REDIS_URL (default localhost:6379) | SMTP_HOST/SMTP_PORT | PORT (default 8080) | GRPC_PORT (default 50051)
EXPOSE 8080 50051

ENTRYPOINT ["./nayz-auth"]
