# Stage 1: Build binary
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static linking for CGO (SQLite) on Alpine
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /app/bin/server ./cmd/server

# Stage 2: Runtime image
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/server /app/server

# Local database persistence layer
RUN mkdir -p /app/data

ENV PORT=8080
ENV DB_PATH=/app/data/plos.db

EXPOSE 8080

ENTRYPOINT ["/app/server"]
