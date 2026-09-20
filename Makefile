.PHONY: fmt vet lint test check build run dev-backend dev-frontend sync-test-migrations docker-build docker-up docker-down docker-logs clean

BINARY_NAME=app
SERVER_PATH=./cmd/server/main.go
TEST_MIGRATIONS_DIRS := \
	cmd/server/testdata/migrations \
	internal/core/testdata/migrations \
	internal/domain/finance/testdata/migrations \
	internal/transport/http/testdata/migrations

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

sync-test-migrations:
	@for dir in $(TEST_MIGRATIONS_DIRS); do \
		mkdir -p $$dir && cp -f migrations/*.sql $$dir/; \
	done

test: sync-test-migrations
	go test -v -race -cover ./...

check: fmt vet lint test

build: check
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/$(BINARY_NAME) $(SERVER_PATH)

run: build
	./bin/$(BINARY_NAME)

dev-backend:
	go run $(SERVER_PATH)

dev-frontend:
	cd web && npm run dev

docker-build:
	docker build -t plos-backend:latest .

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down -v

docker-logs:
	docker compose logs -f backend

clean:
	rm -rf bin/ web/dist data/*.db*
