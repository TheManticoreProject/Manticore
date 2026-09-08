// Package rpcserver answers connection-oriented DCE/RPC calls.
//
// It is the server side of what network/dcerpc/v5/client sends: it reads a bind or
// a request PDU, hands the request's stub to a registered Service, and frames the
// reply. It knows nothing about how the bytes arrived, so the same Dispatcher
// serves an interface over an SMB named pipe, over TCP, or over anything else that
// can hand it one PDU and take one back.
//
// # What a Service is
//
// A Service is one RPC interface: an abstract syntax that identifies it, and a
// Call that runs one opnum against a request stub and returns a response stub. The
// NDR marshalling of those stubs is the Service's business — network/dcerpc/ndr
// does it, and the structures in windows/protocols/... describe the parameters —
// so a Service is the interface's methods and nothing else.
//
// # What this does not do
//
// RPC-level authentication is refused rather than half-implemented. A bind
// carrying an authentication verifier is answered with a bind_nak, because
// completing one means an NTLM or Kerberos exchange over auth3 PDUs and a
// per-context security context to verify and seal every later request against.
// Over an SMB named pipe the pipe is already authenticated by the SMB session, so
// the common case needs none; a caller that needs RPC-level auth is better served
// by that being absent than by a bind that succeeds and then cannot verify
// anything.
//
// A request arriving in fragments is refused with a fault rather than
// reassembled. Reassembly needs per-call state keyed by call id, and the request
// stubs of the interfaces served here — an information level and a couple of
// counts — do not approach a fragment's worth of bytes. Replies are fragmented,
// which is the direction that matters: a share enumeration easily exceeds one
// fragment.
//
// A Dispatcher serves one endpoint, and a request is dispatched to the single
// interface that endpoint serves rather than by the presentation context the
// request names. A transport that cannot tell two opens of an endpoint apart —
// an SMB named pipe reached through a PipeHandler is one — cannot hold the
// per-bind context table that dispatching by context id would need, so an
// endpoint registered with more than one interface faults a request instead of
// guessing. One interface per endpoint is how \srvsvc and \wkssvc are used.
//
// Only the NDR (little-endian, version 2) transfer syntax is accepted. NDR64 is
// declined at bind time rather than accepted and then misencoded, which is the
// only useful answer when the encoder in use speaks one of the two.
//
// References:
//   - [C706] DCE 1.1 RPC, chapter 12 (connection-oriented protocol):
//     https://pubs.opengroup.org/onlinepubs/9629399/
//   - [MS-RPCE] Remote Procedure Call Protocol Extensions:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/290c38b1-92fe-4229-91e6-4fc376ea09c4
package rpcserver
