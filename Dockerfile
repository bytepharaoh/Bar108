FROM golang:1.26.1-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/server \
    ./cmd/main.go

# Stage 2 — Runner
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S bar108 && adduser -S bar108 -G bar108

WORKDIR /app

COPY --from=builder --chown=bar108:bar108 /app/server .

# Copy migrations so the app can run them at startup
COPY --chown=bar108:bar108 db/migrations ./db/migrations

USER bar108
EXPOSE 8080
CMD ["./server"]
