# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Copy go mod first for caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source and build
COPY . .
RUN go build ./...
RUN go test ./...

# Final minimal stage (optional, for demonstration)
FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/ .
CMD ["echo", "SystemOne Go library built successfully"]
