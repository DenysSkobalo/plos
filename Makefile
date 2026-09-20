.PHONY: build run dev-backend dev-frontend test clean

BINARY_NAME=app
SERVER_PATH=./cmd/server/main.go

build:
	@echo "Building frontend..."
	cd web && npm run build
	@echo "Building backend..."
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/$(BINARY_NAME) $(SERVER_PATH)

run: build
	./bin/$(BINARY_NAME)

dev-backend:
	go run $(SERVER_PATH)

dev-frontend:
	cd web && npm run dev

test:
	go test -v -race ./...

clean:
	rm -rf bin/ web/dist data/*.db*
