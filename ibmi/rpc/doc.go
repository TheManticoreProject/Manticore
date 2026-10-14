// Package rpc is the transport-neutral IBM i program-call ABI catalogue.
//
// "rpc" here means calling an IBM i program (library/program plus ordered
// parameters) through the Client Access Remote Command service. It is not
// ONC RPC or DCE/RPC, and it owns no sockets, frames, authentication or
// session state. Concrete program contracts live in program-qualified child
// packages (ibmi/rpc/<library>/<program>) and build calls against the neutral
// types of this package; network/ibmi implements the caller side.
//
// This package must never import network/ibmi/... or any of its own children.
package rpc
