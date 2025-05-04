package llmnr

import (
	"github.com/TheManticoreProject/Manticore/logger"

	"fmt"
	"net"
	"sync"
)

// Server represents an LLMNR server.
//
// The Server struct contains the necessary fields and methods to handle LLMNR (Link-Local Multicast Name Resolution) requests and responses.
// It supports both IPv4 and IPv6 communication over UDP.
//
// Fields:
//   - Handlers: A slice of Handler interfaces that process incoming LLMNR messages.
//   - Network: A string representing the network type. Known networks are "tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only),
//     "udp", "udp4" (IPv4-only), "udp6" (IPv6-only), "ip", "ip4" (IPv4-only), "ip6" (IPv6-only), "unix", "unixgram", and "unixpacket".
//   - Address: A pointer to a net.UDPAddr struct representing the server's address.
//   - Conn: A pointer to a net.UDPConn struct representing the server's UDP connection.
//   - CloseOnce: A sync.Once struct to ensure the server is closed only once.
//   - Closed: A channel that is closed when the server is shut down.
//   - Debug: A boolean flag indicating whether debug mode is enabled.
type Server struct {
	// Handlers is a slice of Handler interfaces that process incoming LLMNR messages.
	Handlers []Handler

	// Known networks are "tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only),
	// "udp", "udp4" (IPv4-only), "udp6" (IPv6-only), "ip", "ip4"
	// (IPv4-only), "ip6" (IPv6-only), "unix", "unixgram" and
	// "unixpacket".
	Network string

	// Address is the address of the server.
	Address *net.UDPAddr

	// Conn is the connection of the server.
	Conn *net.UDPConn

	// CloseOnce is a sync.Once struct to ensure the server is closed only once.
	CloseOnce sync.Once

	// Closed is a channel that is closed when the server is shut down.
	Closed chan struct{}

	// Debug is a boolean flag indicating whether debug mode is enabled.
	Debug bool
}

// NewIPv4Server creates a new LLMNR server for IPv4.
//
// This function initializes a new Server instance configured to use the "udp4" network type for IPv4 communication.
// It does not accept any handlers and initializes the server with an empty list of handlers. The server's internal state,
// including the handlers, closed channel, and debug flag, is initialized.
//
// Returns:
// - A pointer to the newly created Server instance.
// - An error if the server creation fails.
//
// Example usage:
//
//	server, err := llmnr.NewIPv4Server()
//	if err != nil {
//	    log.Fatalf("Failed to create IPv4 server: %v", err)
//	}
//
//	if server.IsIPv4() {
//	    fmt.Println("The server is using an IPv4 address.")
//	} else {
//	    fmt.Println("The server is not using an IPv4 address.")
//	}
func NewIPv4Server() (*Server, error) {
	return NewIPv4ServerWithHandlers([]Handler{})
}

// NewIPv6Server creates a new LLMNR server for IPv6.
//
// This function initializes a new Server instance configured to use the "udp6" network type for IPv6 communication.
// It does not accept any handlers and initializes the server with an empty list of handlers. The server's internal state,
// including the handlers, closed channel, and debug flag, is initialized.
//
// Returns:
// - A pointer to the newly created Server instance.
// - An error if the server creation fails.
//
// Example usage:
//
//	server, err := llmnr.NewIPv6Server()
//	if err != nil {
//	    log.Fatalf("Failed to create IPv6 server: %v", err)
//	}
//
//	if server.IsIPv6() {
//	    fmt.Println("The server is using an IPv6 address.")
//	} else {
//	    fmt.Println("The server is not using an IPv6 address.")
//	}
func NewIPv6Server() (*Server, error) {
	return NewIPv6ServerWithHandlers([]Handler{})
}

// NewIPv4ServerWithHandlers creates a new LLMNR server for IPv4 with the specified handlers.
//
// This function initializes a new Server instance configured to use the "udp4" network type for IPv4 communication.
// It accepts a list of handlers that will be used to process incoming LLMNR requests. The server's internal state,
// including the handlers, closed channel, and debug flag, is initialized.
//
// Parameters:
// - handlers: A slice of Handler instances to handle incoming LLMNR requests.
//
// Returns:
// - A pointer to the newly created Server instance.
// - An error if the server creation fails.
//
// Example usage:
//
//	handlers := []llmnr.Handler{
//	    llmnr.HandlerFunc(myHandlerFunc),
//	}
//	server, err := llmnr.NewIPv4ServerWithHandlers(handlers)
//	if err != nil {
//	    log.Fatalf("Failed to create IPv4 server: %v", err)
//	}
//
//	if server.IsIPv4() {
//	    fmt.Println("The server is using an IPv4 address.")
//	} else {
//	    fmt.Println("The server is not using an IPv4 address.")
//	}
func NewIPv4ServerWithHandlers(handlers []Handler) (*Server, error) {
	server, err := NewServer("udp4", handlers)
	if err != nil {
		return nil, err
	}

	server.Address = &net.UDPAddr{
		IP:   net.ParseIP(IPv4MulticastAddr),
		Port: LLMNRPort,
	}

	return server, nil
}

// NewIPv6ServerWithHandlers creates a new LLMNR server for IPv6 with the specified handlers.
//
// This function initializes a new Server instance configured to use the "udp6" network type for IPv6 communication.
// It accepts a list of handlers that will be used to process incoming LLMNR requests. The server's internal state,
// including the handlers, closed channel, and debug flag, is initialized.
//
// Parameters:
// - handlers: A slice of Handler instances to handle incoming LLMNR requests.
//
// Returns:
// - A pointer to the newly created Server instance.
// - An error if the server creation fails.
//
// Example usage:
//
//	handlers := []llmnr.Handler{
//	    llmnr.HandlerFunc(myHandlerFunc),
//	}
//	server, err := llmnr.NewIPv6ServerWithHandlers(handlers)
//	if err != nil {
//	    log.Fatalf("Failed to create IPv6 server: %v", err)
//	}
//
//	if server.IsIPv6() {
//	    fmt.Println("The server is using an IPv6 address.")
//	} else {
//	    fmt.Println("The server is not using an IPv6 address.")
//	}
func NewIPv6ServerWithHandlers(handlers []Handler) (*Server, error) {
	server, err := NewServer("udp6", handlers)
	if err != nil {
		return nil, err
	}

	server.Address = &net.UDPAddr{
		IP:   net.ParseIP(IPv6MulticastAddr),
		Port: LLMNRPort,
	}

	return server, nil
}

// NewServer creates a new LLMNR server with the specified network and handlers.
//
// This function initializes a new Server instance with the provided network type and a list of handlers.
// The server is configured to use the specified network (e.g., "udp4" for IPv4, "udp6" for IPv6) and
// initializes its internal state, including the handlers, closed channel, and debug flag.
//
// Parameters:
// - network: A string specifying the network type (e.g., "udp4", "udp6").
// - handlers: A slice of Handler instances to handle incoming LLMNR requests.
//
// Returns:
// - A pointer to the newly created Server instance.
// - An error if the server creation fails.
//
// Example usage:
//
//	handlers := []llmnr.Handler{
//	    llmnr.HandlerFunc(myHandlerFunc),
//	}
//	server, err := llmnr.NewServer("udp4", handlers)
//	if err != nil {
//	    log.Fatalf("Failed to create server: %v", err)
//	}
//
//	if server.IsIPv4() {
//	    fmt.Println("The server is using an IPv4 address.")
//	} else {
//	    fmt.Println("The server is not using an IPv4 address.")
//	}
func NewServer(network string, handlers []Handler) (*Server, error) {
	return &Server{
		Handlers: handlers,
		Closed:   make(chan struct{}),
		Debug:    false,
		Network:  network,
	}, nil
}

// IsIPv4 checks if the server's address is an IPv4 address.
//
// Returns:
// - A boolean value: true if the server's address is an IPv4 address, false otherwise.
//
// The function uses the net.IP.To4 method to determine if the IP address is an IPv4 address.
// If the IP address is not an IPv4 address, the method will return nil, and the function will return false.
//
// Example usage:
// server, err := llmnr.NewIPv4Server()
//
//	if err != nil {
//	    log.Fatalf("Failed to create server: %v", err)
//	}
//
//	if server.IsIPv4() {
//	    fmt.Println("The server is using an IPv4 address.")
//	} else {
//	    fmt.Println("The server is not using an IPv4 address.")
//	}
func (s *Server) IsIPv4() bool {
	return s.Address.IP.To4() != nil
}

// IsIPv6 checks if the server's address is an IPv6 address.
//
// Returns:
// - A boolean value: true if the server's address is an IPv6 address, false otherwise.
//
// The function uses the net.IP.To16 method to determine if the IP address is an IPv6 address.
// If the IP address is not an IPv6 address, the method will return nil, and the function will return false.
//
// Example usage:
// server, err := llmnr.NewIPv6Server()
//
//	if err != nil {
//	    log.Fatalf("Failed to create server: %v", err)
//	}
//
//	if server.IsIPv6() {
//	    fmt.Println("The server is using an IPv6 address.")
//	} else {
//	    fmt.Println("The server is not using an IPv6 address.")
//	}
func (s *Server) IsIPv6() bool {
	return s.Address.IP.To16() != nil
}

// SetDebug enables or disables debug mode for the LLMNR server.
//
// Parameters:
// - debug: A boolean value indicating whether to enable (true) or disable (false) debug mode.
//
// When debug mode is enabled, the server will print detailed information about incoming packets,
// decoded messages, and any errors encountered during processing. This can be useful for troubleshooting
// and understanding the server's behavior.
//
// Example usage:
// server, err := llmnr.NewServer(handler)
//
//	if err != nil {
//	    log.Fatalf("Failed to create server: %v", err)
//	}
//
// server.SetDebug(true)
//
// err = server.ListenAndServe()
//
//	if err != nil {
//	    log.Fatalf("Server encountered an error: %v", err)
//	}
func (s *Server) SetDebug(debug bool) {
	s.Debug = debug
}

// ListenAndServe starts the LLMNR server and begins listening for incoming UDP packets on the IPv4 multicast address.
// It creates a UDP connection and assigns it to the server's connection field. The function then enters a loop to
// continuously read from the UDP connection, decode incoming messages, and pass them to the server's handler.
//
// Returns:
// - An error if the server fails to start listening on the UDP connection or encounters an error during execution.
//
// Example usage:
// server, err := llmnr.NewServer(handler)
//
//	if err != nil {
//	    log.Fatalf("Failed to create server: %v", err)
//	}
//
// err = server.ListenAndServe()
//
//	if err != nil {
//	    log.Fatalf("Server encountered an error: %v", err)
//	}
func (s *Server) ListenAndServe() error {
	if len(s.Handlers) == 0 {
		return fmt.Errorf("no handlers registered")
	}

	if s.Network != "tcp" && s.Network != "tcp4" && s.Network != "tcp6" &&
		s.Network != "udp" && s.Network != "udp4" && s.Network != "udp6" &&
		s.Network != "ip" && s.Network != "ip4" && s.Network != "ip6" &&
		s.Network != "unix" && s.Network != "unixgram" && s.Network != "unixpacket" {
		return fmt.Errorf("invalid network: %s", s.Network)
	}

	conn, err := net.ListenMulticastUDP(s.Network, nil, s.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.Conn = conn

	return s.Serve()
}

// Close gracefully shuts down the LLMNR server by closing the UDP connection and signaling the server to stop.
// It ensures that the server is closed only once, even if Close is called multiple times.
//
// The function uses a sync.Once to guarantee that the server's resources are released only once. It closes the
// server's closed channel to signal any goroutines to stop, and if the server has an active UDP connection, it
// closes the connection.
//
// Returns:
// - An error if the server fails to close the UDP connection or encounters an error during the shutdown process.
//
// Example usage:
// server, err := llmnr.NewServer(handler)
//
//	if err != nil {
//	    log.Fatalf("Failed to create server: %v", err)
//	}
//
// err = server.ListenAndServe()
//
//	if err != nil {
//	    log.Fatalf("Server encountered an error: %v", err)
//	}
//
// // When you want to stop the server
// err = server.Close()
//
//	if err != nil {
//	    log.Fatalf("Failed to close server: %v", err)
//	}
func (s *Server) Close() error {
	s.CloseOnce.Do(
		func() {
			close(s.Closed)
			if s.Conn != nil {
				s.Conn.Close()
			}
		},
	)
	return nil
}

// Serve handles incoming LLMNR requests and processes them in a loop until the server is closed.
//
// The function reads packets from the UDP connection, decodes the messages, and handles LLMNR queries.
// It uses a buffer to read the incoming packets and processes each packet in a separate goroutine.
//
// The function also supports debugging, printing detailed information about the received packets and errors.
//
// Returns:
// - An error if the server encounters an issue while reading from the UDP connection.
//
// Example usage:
// server, err := llmnr.NewServer(handler)
//
//	if err != nil {
//	    log.Fatalf("Failed to create server: %v", err)
//	}
//
// err = server.ListenAndServe()
//
//	if err != nil {
//	    log.Fatalf("Server encountered an error: %v", err)
//	}
//
// // When you want to stop the server
// err = server.Close()
//
//	if err != nil {
//	    log.Fatalf("Failed to close server: %v", err)
//	}
func (s *Server) Serve() error {
	buffer := make([]byte, MaxPacketSize)
	for {
		select {
		case <-s.Closed:
			return nil
		default:
			n, remoteAddr, err := s.Conn.ReadFromUDP(buffer)
			if err != nil {
				if s.Debug {
					logger.Debug(fmt.Sprintf("Error reading from UDP: %s\n", err.Error()))
				}
				continue
			}
			if s.Debug {
				logger.Debug(fmt.Sprintf("Received packet from %s\n", remoteAddr.String()))
			}

			msg, err := DecodeMessage(buffer[:n])
			if err != nil {
				if s.Debug {
					logger.Debug(fmt.Sprintf("Error decoding message: %s\n", err.Error()))
				}
				continue
			}
			if s.Debug {
				logger.Debug(fmt.Sprintf("Decoded message: %+v\n", msg))
			}

			if !msg.IsQuery() {
				if s.Debug {
					logger.Debug(fmt.Sprintf("Received non-query message from %s\n", remoteAddr.String()))
				}
				continue
			}
			if s.Debug {
				logger.Debug(fmt.Sprintf("Received query message from %s\n", remoteAddr.String()))
			}

			// Create response writer
			writer := NewResponseWriter(s, remoteAddr)

			// Handle the query in a separate goroutine
			go s.processHandlers(s, remoteAddr, writer, msg)
		}
	}
}

// processHandlers processes the incoming LLMNR message by invoking all registered handlers.
//
// Parameters:
// - w: The ResponseWriter used to send responses back to the client.
// - r: The Message received from the client.
//
// The function iterates over all handlers registered in the Server instance and calls their ServeLLMNR method,
// passing the ResponseWriter and the received Message as arguments. Each handler is responsible for processing
// the message and generating an appropriate response.
//
// Example usage:
//
//	server := &Server{
//	    Handlers: []Handler{handler1, handler2},
//	}
//
// server.processHandlers(responseWriter, message)
//
// This function is typically called internally by the Server when a new LLMNR query is received.
func (s *Server) processHandlers(server *Server, remoteAddr net.Addr, writer ResponseWriter, message *Message) {
	for _, handler := range s.Handlers {
		continueProcessing := handler.Run(server, remoteAddr, writer, message)
		if !continueProcessing {
			break
		}
	}
}
