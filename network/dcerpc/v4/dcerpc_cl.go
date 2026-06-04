// Package dcerpccl implements the connectionless (datagram) DCE/RPC protocol
// (C706 protocol version 4, rpc_vers = 4), as used over datagram transports such
// as ncadg_ip_udp (UDP, well-known endpoint-mapper port 135). It is the v4
// counterpart to the connection-oriented protocol implemented under
// network/dcerpc/v5.
//
// The connectionless protocol differs fundamentally from the connection-oriented
// one:
//
//   - No Bind handshake. Each PDU is self-describing: the interface (if_id,
//     if_vers), activity (act_id), and server instance (server_boot) are carried
//     in every PDU header, so a call is established implicitly by the first
//     request rather than by a prior bind.
//   - The common header is a fixed 80 octets (vs. 16 for connection-oriented) and
//     carries a single version field (rpc_vers, encoded in the 4 least
//     significant bits) — there is no separate minor-version field as there is in
//     the connection-oriented header.
//   - Reliability lives above the transport: calls are carried as datagrams with
//     RPC-layer fragmentation and acknowledgement PDUs (request/ping/response/
//     fault/working/nocall/reject/ack/cl_cancel/fack/cancel_ack).
//   - The NDR transfer syntax and data-representation label (drep) are shared with
//     the connection-oriented protocol; only the first 3 octets of drep appear in
//     connectionless headers. The shared NDR codec lives in network/dcerpc/ndr and
//     the transfer-syntax identifiers in network/dcerpc/syntax.
//
// Connectionless PDU types and their ptype values ([C706] chapter 12; note the
// numbering differs from the connection-oriented ptype values):
//
//	request 0   ping 1   response 2   fault 3   working 4   nocall 5
//	reject 6    ack 7    cl_cancel 8  fack 9    cancel_ack 10
//
// STATUS: The connectionless PDU codec (network/dcerpc/v4/pdu, Phase 1), the
// ncadg_ip_udp datagram transport (network/dcerpc/v4/transport, Phase 2), and the
// client-side protocol machine (network/dcerpc/v4/client, Phase 3) are implemented.
// The endpoint mapper over UDP and concrete interface bindings are not yet
// implemented. v4 is effectively dead against modern Windows, which removed
// ncadg_ip_udp server support starting with Windows Vista / Server 2008; it remains
// relevant only for legacy and non-Windows DCE targets.
//
// References:
//   - [C706] DCE 1.1: Remote Procedure Call (The Open Group, 1997):
//     https://pubs.opengroup.org/onlinepubs/9629399/toc.htm
//   - [C706] chapter 12, RPC PDU Encodings (connectionless PDU formats):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [C706] chapter 10, Connectionless RPC Protocol Machines (state machine,
//     fragmentation, acknowledgement): https://pubs.opengroup.org/onlinepubs/9629399/chap10.htm
//   - [MS-RPCE] Remote Procedure Call Protocol Extensions:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/
package dcerpccl

// ProtocolVersion is the connectionless DCE/RPC protocol version, carried in the 4
// least significant bits of the rpc_vers field of every connectionless PDU header
// ([C706] section 12.6.3.1). Unlike the connection-oriented header, the
// connectionless header has no separate minor-version field.
const ProtocolVersion uint8 = 4
