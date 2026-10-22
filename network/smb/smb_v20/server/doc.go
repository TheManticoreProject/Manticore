// Package server implements the server side of SMB 2.0.2 through 3.1.1.
//
// A Server serves disk shares backed by the FileSystem implementations of the
// SMB1 server package (network/smb/smb_v10/server), so both protocols share one
// storage model and one path containment check. IPC$ is always present so
// clients can connect to it, but no named pipe can be opened on it.
//
// Served: the SMB1 to SMB2 negotiate upgrade, every dialect from 2.0.2 to 3.1.1
// with the 3.1.1 negotiate contexts and pre-authentication integrity, NTLM and
// Kerberos session setup over SPNEGO with guest and anonymous policies, tree
// connect, CREATE, CLOSE, READ, WRITE, FLUSH, QUERY_INFO, SET_INFO,
// QUERY_DIRECTORY, CHANGE_NOTIFY (through a FileSystem that implements the SMB1
// package's Watcher), CANCEL, ECHO, compounded requests, credits, signing (HMAC-SHA256,
// AES-CMAC, AES-GMAC), TRANSFORM_HEADER encryption (AES-CCM and AES-GCM) and the
// secure negotiate validation of 3.0 and 3.0.2.
//
// Access rights, share modes and read-only shares are enforced: a handle is
// granted the rights its CREATE asked for and refused an operation outside them,
// and an open whose share mode conflicts with another is refused with
// STATUS_SHARING_VIOLATION.
//
// Oplocks are granted at level II and leases with read caching only, and broken
// with a server-initiated notification when another handle changes the file.
//
// A Handler chain runs before the built-in dispatch, as in the SMB1 server, so a
// caller can observe or answer any request — CaptureHandler uses it to harvest
// NTLM responses — and ServeConn serves an already established transport, for a
// proxy that owns the socket.
//
// Not implemented: LOCK and the security-descriptor info classes are answered
// STATUS_NOT_SUPPORTED; named pipes are not served, so IPC$ accepts a tree
// connect but a CREATE on it is answered STATUS_OBJECT_NAME_NOT_FOUND;
// multichannel session binding is refused with STATUS_REQUEST_NOT_ACCEPTED; and
// the durable-handle create contexts are ignored, so the open is granted as an
// ordinary handle.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/
package server
