package llmnr

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/TheManticoreProject/Manticore/network/llmnr/class"
	"github.com/TheManticoreProject/Manticore/network/llmnr/constants"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
)

// Client represents an LLMNR client that can send queries and receive responses.
//
// The Client struct provides methods to create a new client, send queries, and close the client connection.
// It manages a UDP connection and uses a sync.Map to keep track of ongoing queries.
//
// Fields:
//   - conn: A pointer to the UDP connection used for sending and receiving LLMNR messages.
//   - timeout: The duration to wait for a response before timing out.
//   - queries: A sync.Map that maps query IDs to channels for receiving responses.
//   - closeOnce: Ensures the client is closed only once.
//   - closed: A channel that is closed when the client is closed.
//   - dest: The destination address queries are sent to. It defaults to the LLMNR
//     IPv4 multicast group (224.0.0.252:5355) so that normal usage is unchanged;
//     it is overridable to allow queries to be directed at a specific responder.
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

	// dest is the destination address queries are written to. It defaults to
	// the LLMNR IPv4 multicast group and is overridable so queries can be
	// directed at a specific responder (e.g. for testing or unicast probing).
	dest *net.UDPAddr
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
		dest: &net.UDPAddr{
			IP:   net.ParseIP(constants.IPv4MulticastAddr),
			Port: constants.ListenPort,
		},
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
func (c *Client) Query(ctx context.Context, name string, qtype llmnr_type.Type) (*message.Message, error) {
	msg := message.NewMessage()
	msg.SetQuery()
	if err := msg.AddQuestion(name, llmnr_type.Type(qtype), class.ClassIN); err != nil {
		return nil, fmt.Errorf("failed to add question: %w", err)
	}

	// Create response channel
	responseChan := make(chan *message.Message, 1)
	c.Queries.Store(msg.Header.Identifier, responseChan)
	defer c.Queries.Delete(msg.Header.Identifier)

	// Send query to the configured destination (the LLMNR multicast group by
	// default, or an overridden responder address).
	encoded, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to encode message: %w", err)
	}

	if _, err := c.Conn.WriteToUDP(encoded, c.dest); err != nil {
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
	buffer := make([]byte, constants.MaxPacketSize)
	for {
		select {
		case <-c.Closed:
			return
		default:
			n, _, err := c.Conn.ReadFromUDP(buffer)
			if err != nil {
				continue
			}

			msg := message.Message{}
			_, err = msg.Unmarshal(buffer[:n])
			if err != nil {
				continue
			}

			if !msg.IsResponse() {
				continue
			}

			// Find the matching query
			if ch, ok := c.Queries.Load(msg.Header.Identifier); ok {
				responseChan := ch.(chan *message.Message)
				select {
				case responseChan <- &msg:
				default:
				}
			}
		}
	}
}
