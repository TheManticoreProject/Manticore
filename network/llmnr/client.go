package llmnr

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// Client represents an LLMNR client that can send queries and receive responses.
//
// The Client struct provides methods to create a new client, send queries, and close the client connection.
// It manages a UDP connection and uses a sync.Map to keep track of ongoing queries.
//
// Fields:
// - conn: A pointer to the UDP connection used for sending and receiving LLMNR messages.
// - timeout: The duration to wait for a response before timing out.
// - queries: A sync.Map that maps query IDs to channels for receiving responses.
// - closeOnce: Ensures the client is closed only once.
// - closed: A channel that is closed when the client is closed.
//
// Usage example:
//
//	client, err := NewClient()
//	if err != nil {
//	    log.Fatalf("Failed to create client: %v", err)
//	}
//	defer client.Close()
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	resp, err := client.Query(ctx, "example.local", TypeA)
//	if err != nil {
//	    log.Fatalf("Query failed: %v", err)
//	}
//	fmt.Printf("Received response: %v\n", resp)
type Client struct {
	Conn      *net.UDPConn
	Timeout   time.Duration
	Queries   sync.Map
	CloseOnce sync.Once
	Closed    chan struct{}
}

// NewClient creates a new LLMNR client with a UDP connection.
//
// The function initializes a UDP connection for the client to use for sending and receiving LLMNR messages.
// It sets a default timeout duration for queries and starts a read loop to handle incoming responses.
//
// Returns:
//   - A pointer to the newly created Client.
//   - An error if the UDP connection could not be created.
//
// Usage example:
//
//	client, err := NewClient()
//	if err != nil {
//	    log.Fatalf("Failed to create client: %v", err)
//	}
//	defer client.Close()
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	resp, err := client.Query(ctx, "example.local", TypeA)
//	if err != nil {
//	    log.Fatalf("Query failed: %v", err)
//	}
//	fmt.Printf("Received response: %v\n", resp)
func NewClient() (*Client, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{})
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %w", err)
	}

	c := &Client{
		Conn:    conn,
		Timeout: 2 * time.Second,
		Closed:  make(chan struct{}),
	}

	go c.readLoop()

	return c, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	c.CloseOnce.Do(func() {
		close(c.Closed)
		c.Conn.Close()
	})
	return nil
}

// Query sends an LLMNR query and waits for a response
func (c *Client) Query(ctx context.Context, name string, qtype uint16) (*Message, error) {
	msg := NewMessage()
	msg.SetQuery()
	if err := msg.AddQuestion(name, qtype, ClassIN); err != nil {
		return nil, fmt.Errorf("failed to add question: %w", err)
	}

	// Create response channel
	responseChan := make(chan *Message, 1)
	c.Queries.Store(msg.ID, responseChan)
	defer c.Queries.Delete(msg.ID)

	// Send query
	addr := &net.UDPAddr{
		IP:   net.ParseIP(IPv4MulticastAddr),
		Port: LLMNRPort,
	}

	encoded, err := msg.Encode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode message: %w", err)
	}

	if _, err := c.Conn.WriteToUDP(encoded, addr); err != nil {
		return nil, fmt.Errorf("failed to send query: %w", err)
	}

	// Wait for response or timeout
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(c.Timeout):
		return nil, fmt.Errorf("query timeout")
	case resp := <-responseChan:
		return resp, nil
	}
}

func (c *Client) readLoop() {
	buffer := make([]byte, MaxPacketSize)
	for {
		select {
		case <-c.Closed:
			return
		default:
			n, _, err := c.Conn.ReadFromUDP(buffer)
			if err != nil {
				continue
			}

			msg, err := DecodeMessage(buffer[:n])
			if err != nil {
				continue
			}

			if !msg.IsResponse() {
				continue
			}

			// Find the matching query
			if ch, ok := c.Queries.Load(msg.ID); ok {
				responseChan := ch.(chan *Message)
				select {
				case responseChan <- msg:
				default:
				}
			}
		}
	}
}
