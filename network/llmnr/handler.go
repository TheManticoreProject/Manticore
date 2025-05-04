package llmnr

import "net"

// Handler defines the interface for processing LLMNR queries
type Handler interface {
	Run(server *Server, remoteAddr net.Addr, writer ResponseWriter, message *Message) bool
}

// HandlerFunc is an adapter to allow regular functions to serve as LLMNR handlers
type HandlerFunc func(*Server, net.Addr, ResponseWriter, *Message) bool

// Run executes the handler function to process an LLMNR query.
//
// Parameters:
// - server: The Server instance that received the query.
// - remoteAddr: The address of the client that sent the query.
// - writer: The ResponseWriter used to send responses back to the client.
// - message: The Message received from the client.
//
// The function calls the handler function with the provided ResponseWriter and Message.
// It allows regular functions to be used as LLMNR handlers by adapting them to the Handler interface.
//
// Example usage:
//
//	handlerFunc := func(w llmnr.ResponseWriter, r *llmnr.Message) {
//	    // Process the LLMNR query and write a response
//	}
//	handler := llmnr.HandlerFunc(handlerFunc)
//	server.RegisterHandler(handler)
//
// This function is typically called internally by the Server when a new LLMNR query is received.
func (f HandlerFunc) Run(server *Server, remoteAddr net.Addr, writer ResponseWriter, message *Message) bool {
	return f(server, remoteAddr, writer, message)
}

// RegisterHandler registers a new handler to the LLMNR server.
//
// Parameters:
// - handler: The Handler to be registered. It must implement the Handler interface.
//
// The function appends the provided handler to the server's list of handlers. These handlers
// will be invoked to process incoming LLMNR queries. Handlers are executed in the order they
// are registered.
//
// Example usage:
//
//	handler := &MyHandler{}
//	server.RegisterHandler(handler)
//
// This function is typically called before starting the server to ensure that all necessary
// handlers are in place to process incoming queries.
func (s *Server) RegisterHandler(handler Handler) {
	s.Handlers = append(s.Handlers, handler)
}
