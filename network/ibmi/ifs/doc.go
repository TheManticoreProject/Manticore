// Package ifs is a read-only client for the IBM i File host server
// (as-file / as-file-s, server ID 0xE002): open a stream file of the
// integrated file system by path, read it, and close it.
//
// A Session is one authenticated server job, started by Start on a
// connection dialed to the File endpoint with the same credential and
// password family as Sign-on (see network/ibmi/client.OpenIFS). Start
// requests data-stream level 8, as the V5R3-V5R4 client does, and learns
// the pathname CCSID. Every level below 16 shares the 32-bit layouts this
// package implements (V5R4 answers 12); a server answering 16 or more is
// refused rather than spoken to in the wrong layout. Paths
// are sent exactly as given, in the negotiated pathname CCSID (UTF-16BE for
// CCSIDs 1200 and 13488), with no case change or normalization; file bytes
// are returned unconverted.
//
// Operations on a Session are serialized, so it is safe for concurrent
// use. A read may be answered by a chain of replies; the whole chain is
// consumed, bounded by StartOptions.MaxChain packets and by the requested
// length, and bytes received before an end-of-file indication are always
// returned with it. File handles are scoped to the server job: after a
// transport failure the connection is poisoned and every open File becomes
// unusable; a handle is never replayed on another connection.
//
// Errors report the file server's numeric return code and operation, never
// paths or file content.
//
// Layouts follow iSeriesNavigatorReverse docs/08-file-server-ifs.md and
// docs/10 (the JTOpen 4.8 generation that the V5R4 baseline targets).
// Writing, directory listing and attribute queries are not implemented.
package ifs
