package main

import (
	"context"
	"fmt"
	pb "kvstore/pkg/pb/api/proto"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connnect: %v", err)
	}
	defer conn.Close()

	client := pb.NewKVStoreClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  client set <key> <value>")
		fmt.Println("  client get <key>")
		fmt.Println("  client delete <key>")
		fmt.Println("  client list [limit]")
		return
	}

	command := os.Args[1]
	switch command {
	case "set":
		if len(os.Args) != 4 {
			fmt.Println("Usage: client set <key> <value>")
			return
		}

		key := os.Args[2]
		value := os.Args[3]

		resp, err := client.Set(ctx, &pb.SetRequest{Key: key, Value: value})
		if err != nil {
			log.Fatalf("Set failed: %v", err)
		}

		fmt.Printf("Set success: %t\n", resp.Success)
	case "get":
		if len(os.Args) != 3 {
			fmt.Println("Usage: client get <key>")
			return
		}

		key := os.Args[2]
		resp, err := client.Get(ctx, &pb.GetRequest{Key: key})
		if err != nil {
			log.Fatalf("Get failed: %v", err)
		}

		if resp.Found {
			fmt.Printf("Value: %s\n", resp.Value)
		} else {
			fmt.Println("Key not found")
		}
	case "delete":
		if len(os.Args) != 3 {
			fmt.Println("Usage: client delete <key>")
			return
		}
		key := os.Args[2]

		resp, err := client.Delete(ctx, &pb.DeleteRequest{Key: key})
		if err != nil {
			log.Fatalf("Delete failed: %v", err)
		}

		if resp.Existed {
			fmt.Println("Key deleted successfully")
		} else {
			fmt.Println("Key did not exist")
		}
	case "list":
		limit := int32(0)
		if len(os.Args) == 3 {
			// Parse limit if provided
			fmt.Sscanf(os.Args[2], "%d", &limit)
		}

		resp, err := client.List(ctx, &pb.ListRequest{Limit: &limit})
		if err != nil {
			log.Fatalf("List failed: %v", err)
		}

		fmt.Printf("Found %d key-value pairs:\n", len(resp.Pairs))
		for _, pair := range resp.Pairs {
			fmt.Printf("  %s: %s\n", pair.Key, pair.Value)
		}
	default:
		fmt.Println("Unknown command:", command)
		fmt.Println("Available commands: get, set, delete, list")
	}
}
