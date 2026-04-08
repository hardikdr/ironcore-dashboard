.PHONY: build run dev tidy test

build:
	go build -o bin/ironcore-dashboard ./cmd/server

run:
	go run ./cmd/server --addr :8080

tidy:
	go mod tidy

test:
	go test ./...
