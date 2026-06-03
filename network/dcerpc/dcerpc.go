// Package dcerpc implements a DCE/RPC (C706) client with the Microsoft MS-RPCE
// extensions, layered on top of a pluggable transport.
//
// The connection-oriented DCE/RPC protocol carries each call as one or more PDUs
// (Bind, Bind_Ack, Request, Response, Fault, ...) over a transport that provides a
// request/response message exchange. The transport abstraction lives in
// network/dcerpc/transport; the SMB named-pipe transport (protocol sequence
// ncacn_np) lives in network/dcerpc/transport/smb and wraps an authenticated
// network/smb/smb_v10/client.Client.
//
// This package currently provides the transport layer (Phase 1). The PDU layer and
// the high-level Bind/Call client are built on top of it in later phases.
//
// References:
//   - [C706] DCE 1.1: Remote Procedure Call
//   - [MS-RPCE] Remote Procedure Call Protocol Extensions
package dcerpc

// Connection-oriented DCE/RPC protocol version, as carried in the common header of
// every PDU ([C706] section 12.6.3.1).
const (
	// MajorVersion is the rpc_vers field value.
	MajorVersion uint8 = 5
	// MinorVersion is the rpc_vers_minor field value.
	MinorVersion uint8 = 0
)
