FROM golang:1.24-bullseye AS builder

WORKDIR /build

RUN apt-get update && apt-get install -y \
    librdkafka-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -o main .

FROM debian:bullseye-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    librdkafka1 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /build/main /app/main

RUN chmod +x /app/main

EXPOSE 3000

CMD ["/app/main"]