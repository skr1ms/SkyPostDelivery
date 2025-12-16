.PHONY: run build tests clean proto sqlc swag migrate migrate-create migrate-up migrate-down migrate-force docker-up docker-down dev-up dev-down dev-logs dev-rebuild install-tools tidy

run:
	go run cmd/app/main.go

build:
	go build -o bin/orchestrator cmd/app/main.go

test:
	@echo "Running all tests with coverage..."
	@go test -v -count=1 ./... -coverprofile=coverage.out -covermode=atomic
	@echo "\n====== TOTAL COVERAGE ======"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'
	@rm -f coverage.out

clean:
	rm -rf bin/

proto:
	mkdir -p pkg/pb
	protoc --proto_path=../../proto \
		--go_out=pkg/pb --go_opt=paths=source_relative \
		--go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative \
		../../proto/orchestrator.proto

sqlc:
	sqlc generate

swag:
	swag init -g cmd/app/main.go -o docs --parseDependency --parseInternal

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	migrate -path migrations -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" down

migrate-force:
	migrate -path migrations -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" force $(version)

docker-up:
	docker-compose -f ../deployment/docker/docker-compose.yml up -d

docker-down:
	docker-compose -f ../deployment/docker/docker-compose.yml down

dev-up:
	docker-compose -f ../../deployment/docker/docker-compose.dev.yml up -d

dev-down:
	docker-compose -f ../../deployment/docker/docker-compose.dev.yml down

dev-logs:
	docker-compose -f ../../deployment/docker/docker-compose.dev.yml logs -f

dev-rebuild:
	docker-compose -f ../../deployment/docker/docker-compose.dev.yml up -d --build

install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/swaggo/swag/cmd/swag@latest

lint:
	golangci-lint run

fmt:
	go fmt ./...

tidy:
	go mod tidy
