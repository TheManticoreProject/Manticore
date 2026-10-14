// Package rpc is the transport-neutral IBM i program-call ABI catalogue.
//
// "rpc" here means calling an IBM i program (library/program plus ordered
// parameters) through the Client Access Remote Command service. It is not
// ONC RPC or DCE/RPC, and it owns no sockets, frames, authentication or
// session state. Concrete program contracts live in program-qualified child
// packages (ibmi/rpc/<library>/<program>) and build calls against the neutral
// types of this package; network/ibmi implements the caller side.
//
// Every value is validated by its constructor and is immutable afterwards:
// NewObjectName, Input/EmptyInput/Output/InputOutput and NewCall copy their
// input and fail with an error matching ErrInvalid, so a malformed call never
// reaches a transport. Result.Validate checks that a returned Result aligns
// with its Call. Formatting a Parameter or Call never prints parameter data,
// and IBMMessage.String prints only the message ID and severity.
//
// This package must never import network/ibmi/... or any of its own children.
package rpc
