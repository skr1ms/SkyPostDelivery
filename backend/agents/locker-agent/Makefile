.PHONY: run build test clean fmt tidy deps install-gocv

run:
	go run cmd/app/main.go

build:
	go build -o bin/locker-agent cmd/app/main.go

test:
	go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -count=1 ./... -coverprofile=coverage.out -covermode=atomic
	@echo "\n====== TOTAL COVERAGE ======"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'
	@rm -f coverage.out

clean:
	rm -rf bin/ coverage.out

fmt:
	go fmt ./...

tidy:
	go mod tidy

deps:
	go mod download

install-gocv:
	@echo "Installing GoCV dependencies..."
	@echo "Please refer to: https://gocv.io/getting-started/"
	@echo "For Ubuntu/Debian: sudo apt-get install libopencv-dev"
	@echo "For macOS: brew install opencv"

docker-build:
	docker build -t locker-agent:latest .

docker-run:
	docker run -p 8000:8000 --privileged --env-file .env locker-agent:latest

lint:
	golangci-lint run
