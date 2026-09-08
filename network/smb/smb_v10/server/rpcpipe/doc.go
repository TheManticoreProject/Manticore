// Package rpcpipe makes an SMB1 IPC$ share answer RPC calls.
//
// It supplies the PipeHandler that network/smb/smb_v10/server defines but does
// not implement: register an IPC$ share, give it a Handler, and a client that
// opens \srvsvc or \wkssvc on it reaches a working RPC endpoint. Without it a
// caller has to write DCE/RPC framing and NDR by hand before a server is
// browsable.
//
//	srv, err := server.NewServer(server.Config{})
//	srv.AddShare(&server.Share{Name: "PUBLIC", FS: fs, Comment: "Public files"})
//	srv.AddShare(&server.Share{
//	    Name:  "IPC$",
//	    Type:  server.ShareTypeNamedPipe,
//	    Pipes: rpcpipe.ForServer(srv, rpcpipe.Options{ServerName: "MANTICORE"}),
//	})
//
// It is a package of its own rather than part of the server. The server has no
// reason to know about DCE/RPC, and a caller who only serves files should not
// have an RPC stack linked in; the dependency runs from here to the server and
// never back.
//
// # What is served
//
// srvsvc NetrShareEnum at information levels 0 and 1, which is the call a client
// makes to list shares — "net view", "smbclient -L" and Explorer's network
// browser all arrive through it — and wkssvc NetrWkstaGetInfo at levels 100 and
// 101, which is what a client asks to identify the host.
//
// The shares reported are the server's own when the Handler was built with
// ForServer, read on each call, so a share added with AddShare afterwards appears
// in the next enumeration. A caller that wants to report something else supplies
// Options.Shares instead.
//
// Any other opnum faults with nca_s_op_rng_error. Any other information level
// answers ERROR_INVALID_LEVEL, which is a successful call reporting a refusal
// rather than a fault, and is what makes a client fall back to a level that is
// served.
//
// # One interface per pipe
//
// A pipe here carries exactly one interface: \srvsvc is srvsvc and \wkssvc is
// wkssvc. That is how both are used, and it is also the only arrangement this can
// serve — the PipeHandler contract names a pipe by name and not by open handle,
// so a handler cannot tell two clients of one pipe apart and cannot hold the
// per-open bind state that a pipe carrying several interfaces would need. A bind
// is still checked against the interface the pipe carries, and one naming a
// different interface is refused with a bind_nak.
package rpcpipe
