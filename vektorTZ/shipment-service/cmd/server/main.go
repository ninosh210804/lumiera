package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"shipment-service/api/proto/shipmentpb"
	"shipment-service/internal/application"
	shipgrpc "shipment-service/internal/infrastructure/grpc"
	"shipment-service/internal/infrastructure/repository"
)

func main() {
	repo := repository.NewMemoryRepository()
	service := application.NewShipmentService(repo)
	handler := shipgrpc.NewShipmentHandler(service)

	grpcServer := grpc.NewServer()
	shipmentpb.RegisterShipmentServiceServer(grpcServer, handler)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	fmt.Println("gRPC server listening on :50051")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
