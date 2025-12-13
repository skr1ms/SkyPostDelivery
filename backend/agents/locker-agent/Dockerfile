FROM golang:1.25.3 AS builder

RUN apt-get update && apt-get install -y \
    libopencv-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o locker-agent cmd/app/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    libopencv-core4.6 \
    libopencv-imgcodecs4.6 \
    libopencv-objdetect4.6 \
    libopencv-videoio4.6 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /build/locker-agent .

RUN chmod +x locker-agent

EXPOSE 8000

CMD ["./locker-agent"]
