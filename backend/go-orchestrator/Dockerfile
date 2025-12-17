FROM golang:1.25.3-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bin/orchestrator cmd/app/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/orchestrator .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080 50053

CMD ["./orchestrator"]