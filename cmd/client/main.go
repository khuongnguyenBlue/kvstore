package main

import (
	"bufio"
	"context"
	"fmt"
	pb "kvstore/pkg/pb/api/proto"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InteractiveClient struct {
	client pb.KVStoreClient
	conn   *grpc.ClientConn
}

func NewInteractiveClient(serverAddr string) (*InteractiveClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	client := pb.NewKVStoreClient(conn)

	return &InteractiveClient{
		client: client,
		conn:   conn,
	}, nil
}

func (ic *InteractiveClient) Close() {
	if ic.conn != nil {
		ic.conn.Close()
	}
}

// createContext creates a new context with timeout for each operation
func (ic *InteractiveClient) createContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func (ic *InteractiveClient) Run() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=== KVStore Interactive Client ===")
	fmt.Println("Type 'help' for available commands")
	fmt.Println("Type 'quit' or 'exit' to exit")
	fmt.Println()

	for {
		fmt.Print("kvstore> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "quit" || input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if input == "help" {
			ic.showHelp()
			continue
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		switch command {
		case "get":
			ic.handleGet(args)
		case "set":
			ic.handleSet(args)
		case "delete":
			ic.handleDelete(args)
		case "list":
			ic.handleList(args)
		case "clear":
			fmt.Print("\033[H\033[2J") // Clear screen
		default:
			fmt.Printf("Unknown command: %s\n", command)
			fmt.Println("Type 'help' for available commands")
		}
	}
}

func (ic *InteractiveClient) showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  get <key>                    - Get value for a key")
	fmt.Println("  set <key> <value> [ttl]      - Set key-value pair with optional TTL")
	fmt.Println("  delete <key>                 - Delete a key")
	fmt.Println("  list [limit]                 - List all key-value pairs")
	fmt.Println("  clear                        - Clear screen")
	fmt.Println("  help                         - Show this help")
	fmt.Println("  quit/exit                    - Exit the client")
	fmt.Println()
}

func (ic *InteractiveClient) handleGet(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: get <key>")
		return
	}

	key := args[0]
	ctx, cancel := ic.createContext()
	defer cancel()

	resp, err := ic.client.Get(ctx, &pb.GetRequest{Key: key})
	if err != nil {
		fmt.Printf("❌ Get failed: %v\n", err)
		return
	}

	if resp.Found {
		fmt.Printf("✅ Key: %s\n", key)
		fmt.Printf("📝 Value: %s\n", resp.Value)
	} else {
		fmt.Printf("❌ Key '%s' not found\n", key)
	}
}

func (ic *InteractiveClient) handleSet(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: set <key> <value> [ttl_seconds]")
		return
	}

	key := args[0]
	value := args[1]

	req := &pb.SetRequest{
		Key:   key,
		Value: value,
	}

	// Handle optional TTL parameter
	if len(args) > 2 {
		ttl, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			fmt.Printf("❌ Invalid TTL value: %v\n", err)
			return
		}
		req.TtlSeconds = &ttl
	}

	ctx, cancel := ic.createContext()
	defer cancel()

	resp, err := ic.client.Set(ctx, req)
	if err != nil {
		fmt.Printf("❌ Set failed: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("✅ Successfully set key '%s' with value '%s'\n", key, value)
		if req.TtlSeconds != nil {
			fmt.Printf("⏰ TTL: %d seconds\n", *req.TtlSeconds)
		}
	} else {
		fmt.Printf("❌ Failed to set key '%s'\n", key)
	}
}

func (ic *InteractiveClient) handleDelete(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: delete <key>")
		return
	}

	key := args[0]
	ctx, cancel := ic.createContext()
	defer cancel()

	resp, err := ic.client.Delete(ctx, &pb.DeleteRequest{Key: key})
	if err != nil {
		fmt.Printf("❌ Delete failed: %v\n", err)
		return
	}

	if resp.Existed {
		fmt.Printf("✅ Key '%s' deleted successfully\n", key)
	} else {
		fmt.Printf("⚠️  Key '%s' did not exist\n", key)
	}
}

func (ic *InteractiveClient) handleList(args []string) {
	var req *pb.ListRequest
	if len(args) == 1 {
		// Parse limit if provided
		parsedLimit, err := strconv.ParseInt(args[0], 10, 32)
		if err != nil {
			fmt.Printf("❌ Invalid limit value: %v\n", err)
			return
		}
		limit := int32(parsedLimit)
		req = &pb.ListRequest{Limit: &limit}
	} else {
		// No limit provided - don't set the field
		req = &pb.ListRequest{}
	}

	ctx, cancel := ic.createContext()
	defer cancel()

	resp, err := ic.client.List(ctx, req)
	if err != nil {
		fmt.Printf("❌ List failed: %v\n", err)
		return
	}

	if len(resp.Pairs) == 0 {
		fmt.Println("📭 No key-value pairs found")
		return
	}

	fmt.Printf("📋 Found %d key-value pairs:\n", len(resp.Pairs))
	fmt.Println("┌─────────────────┬─────────────────┐")
	fmt.Println("│ Key             │ Value           │")
	fmt.Println("├─────────────────┼─────────────────┤")

	for _, pair := range resp.Pairs {
		key := pair.Key
		value := pair.Value

		// Truncate if too long
		if len(key) > 15 {
			key = key[:12] + "..."
		}
		if len(value) > 15 {
			value = value[:12] + "..."
		}

		fmt.Printf("│ %-15s │ %-15s │\n", key, value)
	}
	fmt.Println("└─────────────────┴─────────────────┘")
}

func main() {
	serverAddr := "localhost:9090"
	if len(os.Args) > 1 {
		serverAddr = os.Args[1]
	}

	client, err := NewInteractiveClient(serverAddr)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	client.Run()
}
