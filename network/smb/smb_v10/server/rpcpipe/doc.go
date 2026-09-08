// Package rpcpipe makes an SMB1 IPC$ share answer RPC calls.
//
// It supplies the PipeHandler that network/smb/smb_v10/server defines but does
// not implement: register an IPC$ share, give it a Handler, and a client that
// opens \srvsvc or \wkssvc on it reaches a working RPC endpoint. Without it a
// caller has to write DCE/RPC framing and NDR by hand before a server is
// browsable.
//
// Each open of a pipe gets its own Session, so what one client's bind negotiates
// belongs to that client.
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
// # Pipes, opens and interfaces
//
// OpenPipe returns a Session, and the session is where a client's RPC state
// lives: its bind establishes the presentation contexts and the reply fragment
// size that its later requests are read against. Two clients of \srvsvc get two
// sessions and negotiate separately, so neither can see or disturb what the other
// agreed.
//
// \srvsvc carries srvsvc and \wkssvc carries wkssvc, which is how Windows serves
// them, and a bind naming an interface its pipe does not carry is refused with a
// bind_nak. That is a property of these two pipes rather than a limit: a pipe may
// carry several interfaces, and Options.Services adds interfaces to a built-in
// pipe or adds a pipe of its own, so a caller can serve its own interface over
// IPC$ without writing a PipeHandler. A request is dispatched by the presentation
// context its bind negotiated, so which interface answers is what the client
// asked for and never a guess.
package rpcpipe
