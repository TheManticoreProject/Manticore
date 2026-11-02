// Package rpcserver answers connection-oriented DCE/RPC calls.
//
// It is the server side of what network/dcerpc/v5/client sends: it reads a bind or
// a request PDU, hands the request's stub to a registered Service, and frames the
// reply. It knows nothing about how the bytes arrived, so the same Dispatcher
// serves an interface over an SMB named pipe, over TCP, or over anything else that
// can hand it one PDU and take one back.
//
// # Dispatchers and associations
//
// A Dispatcher is the set of interfaces an endpoint answers for, and holds nothing
// that belongs to a client. Dispatcher.Open starts an Association, which is one
// client's conversation with the endpoint and holds what a bind establishes: which
// interface each presentation context names, and the largest fragment that client
// will accept. Association.Handle is what answers a PDU.
//
// The split is what the transport dictates. A caller knows what one client is —
// one open of a named pipe, one TCP connection — and the Dispatcher does not, so
// the caller opens an association per client and the endpoint's interface table is
// shared between them. A bind and the requests after it therefore land on the same
// association, which is what makes them one conversation; a request naming a
// context its association never negotiated is faulted with
// nca_s_fault_context_mismatch rather than guessed at.
//
// # What a Service is
//
// A Service is one RPC interface: an abstract syntax that identifies it, and a
// Call that runs one opnum against a request stub and returns a response stub. The
// NDR marshalling of those stubs is the Service's business — network/dcerpc/ndr
// does it, and the structures in windows/protocols/... describe the parameters —
// so a Service is the interface's methods and nothing else.
//
// # Authentication
//
// An endpoint refuses authenticated binds unless EnableNTLM configured it. With
// NTLM enabled it accepts RPC_C_AUTHN_WINNT (auth_type 10) at the connect,
// packet, integrity and privacy levels: the bind_ack carries the CHALLENGE, the
// client's AUTH3 carries the AUTHENTICATE, which is verified against
// NTLMAuth.CredentialLookup, and at the packet levels every request fragment is
// verified (and at privacy decrypted) and every response fragment signed (and
// sealed). A request on an association whose AUTH3 failed, or one without a
// valid verifier at a packet level, is faulted with nca_s_fault_access_denied.
// The call level (3) is refused, as is privacy without 128-bit keys. An
// association that negotiated security cannot be rebound without it. SPNEGO
// (auth_type 9) and Kerberos binds are still refused with a bind_nak, and an
// alter_context carrying a verifier is faulted: it does not renegotiate
// security. An authenticated association's PDUs are processed one at a time,
// since its security context advances with every PDU in wire order.
//
// A request arriving in fragments is reassembled before the Service sees it, up
// to MaxRequestStub bytes. The association does not multiplex calls, so one
// request is assembled at a time; a fragment that does not continue it (another
// call_id, context or opnum) is answered with nca_s_proto_error and abandons it.
// A first fragment on a context the association never bound is refused at once
// with nca_s_fault_context_mismatch. A whole request, a bind or alter_context,
// and an orphaned PDU each abandon the request in progress; the orphaned PDU is
// not answered. Replies are fragmented to the size the bind negotiated.
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
