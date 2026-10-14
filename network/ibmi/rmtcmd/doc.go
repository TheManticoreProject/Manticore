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
// Close sends the one-way 0x1004 end request and closes the socket.
//
// A Session performs no locking of its own beyond the per-exchange lock of
// the underlying hostserver.Conn; network/ibmi/client owns the lifecycle
// lock and exclusive reservations.
package rmtcmd
