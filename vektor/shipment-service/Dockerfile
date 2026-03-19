# Multi-stage build for minimal image
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache protobuf-compiler

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate protobuf files
RUN protoc --go_out=. --go-grpc_out=. api/proto/*.proto

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/bin/server .

# Expose gRPC port
EXPOSE 50051

# Run the server
CMD ["./server"]
