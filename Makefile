.PHONY: fmt vet lint test check build run dev-backend dev-frontend clean

BINARY_NAME=app
SERVER_PATH=./cmd/server/main.go

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

test:
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

clean:
	rm -rf bin/ web/dist data/*.db*
