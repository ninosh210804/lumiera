.PHONY: proto build run test clean docker-build docker-up docker-down

proto:
	protoc --go_out=. --go-grpc_out=. api/proto/*.proto

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

docker-build:
	docker build -t shipment-service:latest .

docker-run:
	docker run -p 50051:50051 shipment-service:latest

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f shipment-service

help:
	@echo "Available targets:"
	@echo "  proto           - Generate protobuf files"
	@echo "  build           - Build the server binary"
	@echo "  run             - Run the server"
	@echo "  test            - Run all tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  clean           - Clean build artifacts"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  docker-up       - Start services with docker-compose"
	@echo "  docker-down     - Stop services with docker-compose"
	@echo "  docker-logs     - View docker-compose logs"
