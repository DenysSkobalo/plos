# Stage 1: Build Svelte 5 Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build CGO-free Go Binary with Embedded Assets
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/plos ./cmd/server/main.go

# Stage 3: Minimal Scratch/Alpine Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /app/plos /app/plos

RUN mkdir -p /app/data
VOLUME ["/app/data"]

EXPOSE 8080
ENV DB_PATH=/app/data/finance.db
ENTRYPOINT ["/app/plos"]
