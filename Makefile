.PHONY: build build-frontend run dev-backend dev-frontend tidy test demo smoke-test

build-frontend:
	cd frontend && npm install && npm run build

build: build-frontend
	go build -o bin/ironcore-dashboard ./cmd/server

run:
	go run ./cmd/server --addr :8080 --kubeconfig $(HOME)/.kube/config

dev-backend:
	go run ./cmd/server --addr :8080 --kubeconfig $(HOME)/.kube/config

demo:
	go run ./cmd/server --demo --addr :8080

dev-frontend:
	cd frontend && npm run dev

test:
	go test ./...

# smoke-test: build, start demo server, run Playwright tests, stop server.
smoke-test: build
	@echo "Starting demo server on :8181..."
	./bin/ironcore-dashboard --demo --addr :8181 & echo $$! > /tmp/ironcore-demo.pid
	@sleep 2
	@echo "Running Playwright smoke tests..."
	cd frontend && BASE_URL=http://localhost:8181 npm run test:smoke; \
	  STATUS=$$?; \
	  kill $$(cat /tmp/ironcore-demo.pid) 2>/dev/null; \
	  rm -f /tmp/ironcore-demo.pid; \
	  exit $$STATUS

tidy:
	go mod tidy
