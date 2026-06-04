// Package dcerpc implements the connection-oriented DCE/RPC protocol (C706
// protocol version 5, rpc_vers = 5) with the Microsoft MS-RPCE extensions,
// layered on top of a pluggable transport. It lives under network/dcerpc/v5 to
// distinguish it from the connectionless/datagram protocol (rpc_vers = 4), whose
// implementation lives under network/dcerpc/v4.
//
// The connection-oriented DCE/RPC protocol carries each call as one or more PDUs
// (Bind, Bind_Ack, Request, Response, Fault, ...) over a transport that provides a
// request/response message exchange. The transport abstraction lives in
// network/dcerpc/v5/transport; the SMB named-pipe transport (protocol sequence
// ncacn_np) lives in network/dcerpc/v5/transport/smb and wraps an authenticated
// network/smb/smb_v10/client.Client.
//
// References:
//   - [C706] DCE 1.1: Remote Procedure Call (The Open Group, 1997):
//     https://pubs.opengroup.org/onlinepubs/9629399/toc.htm
//   - [C706] chapter 12, RPC PDU Encodings (connection-oriented PDU formats):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [MS-RPCE] Remote Procedure Call Protocol Extensions:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/
package dcerpc

// Connection-oriented DCE/RPC protocol version, as carried in the common header of
// every PDU ([C706] section 12.6.3.1).
const (
	// MajorVersion is the rpc_vers field value.
	MajorVersion uint8 = 5
	// MinorVersion is the rpc_vers_minor field value.
	MinorVersion uint8 = 0
)
