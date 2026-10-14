// Package rmtcmd is the IBM i Remote Command host-server engine (as-rmtcmd /
// as-rmtcmd-s): one authenticated, long-lived server job that later runs CL
// commands and distributed program calls.
//
// Start authenticates the connection on its own, with fresh seeds:
//
//	0x7001 exchange seeds      -> 0xF001 server seed
//	derive a fresh password substitute (never the Sign-on one)
//	0x7002 start server        -> 0xF002 authenticated server job
//	0x1001 exchange attributes -> 0x8001 job CCSID and data-stream level
//
// The data-stream level is converted once into immutable Capabilities.
//
// Operations:
//
//	Call       0x1003 -> 0x8003 outputs | 0x9003 / nonzero code + messages
//	RunCommand 0x1002 -> 0x8002 return code + messages (privileged raw CL)
//
// Session implements rpc.Caller: it maps neutral rpc.Call parameters to
// envelopes (usage 11, 22, and 13 or 33 by capability), decodes bounded
// uncompressed and RLE outputs, and reports IBM messages. Failures are
// typed: *NativeError for a server-reported return code (the result and its
// messages are returned too), *OutcomeError with the send phase for
// transport or protocol failures. Nothing is retried or replayed. A
// malformed body or a return code with unknown body semantics poisons the
// session.
//
// Close sends the one-way 0x1004 end request and closes the socket.
//
// A Session performs no locking of its own beyond the per-exchange lock of
// the underlying hostserver.Conn; network/ibmi/client owns the lifecycle
// lock and exclusive reservations.
package rmtcmd
