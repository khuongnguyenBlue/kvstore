package main

import (
	"kvstore/internal/server"
	"kvstore/internal/storage"
	pb "kvstore/pkg/pb/api/proto"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	store := storage.NewMemoryStore()
	kvServer := server.New(store)

	grpcServer := grpc.NewServer()

	pb.RegisterKVStoreServer(grpcServer, kvServer)

	listen, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Println("gRPC server starting on port 9090...")

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to server: %v", err)
	}
}
