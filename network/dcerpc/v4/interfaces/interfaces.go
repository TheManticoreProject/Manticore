// Package interfaces provides the binding pattern for invoking concrete DCE/RPC
// interfaces over the connectionless (v4) client. A Binding ties an interface
// identity (UUID and version) to a connectionless RPC client, so a per-interface
// package only has to marshal each operation's stub and call Invoke with the right
// opnum.
//
// Endpoint discovery: dynamic interfaces are not on a fixed UDP port. The usual flow
// is to resolve the interface's endpoint through the endpoint mapper first, then
// bind to it:
//
//	eps, _ := epm.New(client.New(udp.New(host, transport.EndpointMapperPort))).Map(iface, maj, min)
//	rpc := client.New(udp.New(host, int(eps[0].Port)))
//	b := interfaces.NewBinding(rpc, iface, maj, min)
//	out, _ := b.Invoke(opnum, stub)
//
// Concrete interface bindings live in subpackages (for example interfaces/mgmt).
package interfaces

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/client"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// Binding associates a connectionless RPC client with a specific interface identity.
type Binding struct {
	rpc   *client.Client
	iface guid.GUID
	major uint16
	minor uint16
}

// NewBinding returns a binding for the given interface UUID and version over rpc.
func NewBinding(rpc *client.Client, iface guid.GUID, major, minor uint16) *Binding {
	return &Binding{rpc: rpc, iface: iface, major: major, minor: minor}
}

// Interface returns the bound interface UUID.
func (b *Binding) Interface() guid.GUID { return b.iface }

// Version returns the bound interface major and minor version.
func (b *Binding) Version() (major, minor uint16) { return b.major, b.minor }

// Invoke calls the operation opnum with the marshalled request stub in and returns
// the response stub.
//
// Calls are issued as idempotent: the connectionless client does not implement the
// conv callback interface needed for at-most-once verification of non-idempotent
// calls, so retransmission must be safe. This suits query operations; an operation
// with side effects should only be invoked when the caller accepts at-least-once
// semantics.
func (b *Binding) Invoke(opnum uint16, in []byte) ([]byte, error) {
	return b.rpc.Call(client.CallRequest{
		Interface:        b.iface,
		InterfaceVersion: uint32(b.major),
		OpNum:            opnum,
		Stub:             in,
		Idempotent:       true,
	})
}
