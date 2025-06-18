FROM golang:1.24.1 AS builder

WORKDIR /app

# Install dependencies
RUN apt-get update && apt-get install -y \
    git \
    build-essential \
    pkg-config \
    librdkafka-dev

# Copy go.mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build app
RUN CGO_ENABLED=1 GOOS=linux go build -o main .


# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/env ./env

# Expose port
EXPOSE 3000

# Run the binary
CMD ["./main"]