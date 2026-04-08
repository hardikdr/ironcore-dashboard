.PHONY: build build-frontend run dev-backend dev-frontend tidy test

build-frontend:
	cd frontend && npm install && npm run build

build: build-frontend
	go build -o bin/ironcore-dashboard ./cmd/server

run:
	go run ./cmd/server --addr :8080 --kubeconfig $(HOME)/.kube/config

dev-backend:
	go run ./cmd/server --addr :8080 --kubeconfig $(HOME)/.kube/config

dev-frontend:
	cd frontend && npm run dev

test:
	go test ./...

tidy:
	go mod tidy
