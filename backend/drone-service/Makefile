.PHONY: run build test tests clean swag proto docker-build docker-run lint fmt tidy

run:
	go run cmd/app/main.go

build:
	go build -o bin/drone-service cmd/app/main.go

test:
	go test -v ./...

tests:
	@echo "Running all tests with coverage..."
	@go test -v -count=1 ./... -coverprofile=coverage.out -covermode=atomic
	@echo "\n====== TOTAL COVERAGE ======"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'
	@rm -f coverage.out

swag:
	swag init -g cmd/app/main.go -o docs --parseDependency --parseInternal

proto:
	mkdir -p pkg/pb
	protoc --proto_path=../../proto \
		--go_out=pkg/pb --go_opt=paths=source_relative \
		--go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative \
		../../proto/orchestrator.proto

clean:
	rm -rf bin/ coverage.out

docker-build:
	docker build -t go-drone-service:latest .

docker-run:
	docker run -p 8001:8001 --env-file .env go-drone-service:latest

lint:
	golangci-lint run

fmt:
	go fmt ./...

tidy:
	go mod tidy
