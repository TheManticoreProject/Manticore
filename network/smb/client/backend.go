package client

import (
	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
	"github.com/TheManticoreProject/Manticore/network/smb"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// Backend is the version-agnostic contract a concrete SMB engine must satisfy to
// be driven by the generic Client. Each engine (smb_v10, smb_v20, …) is wrapped
// by an adapter in this package; the engines themselves do not depend on it.
//
// The surface is flat and stateful, mirroring the underlying engines: a backend
// holds the current session and tree, and the methods act against them. A
// FileHandle returned by OpenFile is opaque to the caller and must be passed
// back to the same backend that produced it.
type Backend interface {
	// Dialect reports the negotiated protocol version this backend speaks.
	Dialect() smb.SMBProtocolVersion

	// ConnectionInfo reports the connection capabilities negotiated with the server
	// (signing requirement, maximum read/write sizes), normalized across dialects.
	ConnectionInfo() ConnectionInfo

	// Login authenticates a session with the server.
	Login(creds *credentials.Credentials) error

	// TreeConnect connects to a share by name (e.g. "C$"), making it the current
	// tree for subsequent file operations. The backend forms the full UNC path.
	TreeConnect(share string) error

	// OpenFile opens or creates a file on the current tree and returns a handle.
	OpenFile(path string, opts OpenOptions) (FileHandle, error)

	// ReadFile reads up to n bytes from the open file starting at off.
	ReadFile(h FileHandle, off uint64, n uint32) ([]byte, error)

	// WriteFile writes data to the open file starting at off and returns the
	// number of bytes written.
	WriteFile(h FileHandle, off uint64, data []byte) (uint32, error)

	// CloseFile closes a handle returned by OpenFile.
	CloseFile(h FileHandle) error

	// ListDirectory enumerates entries of a directory on the current tree that
	// match pattern (e.g. "*").
	ListDirectory(path, pattern string) ([]FileInfo, error)

	// DeleteFile deletes a file on the current tree.
	DeleteFile(path string) error

	// CreateDirectory creates a directory on the current tree.
	CreateDirectory(path string) error

	// DeleteDirectory removes an empty directory on the current tree.
	DeleteDirectory(path string) error

	// RenameFile renames (or moves) a file or directory on the current tree. It
	// fails if newPath already exists.
	RenameFile(oldPath, newPath string) error

	// CheckDirectory returns nil if path names an existing directory on the
	// current tree, and an error otherwise.
	CheckDirectory(path string) error

	// RPCTransport opens a named pipe on the current tree (IPC$) and returns it as
	// a DCE/RPC PDU transport, framed for the negotiated dialect. It is the bridge
	// from the version-agnostic SMB client to the DCE/RPC stack, so callers can run
	// MS-RPC interfaces (srvsvc, lsarpc, …) without knowing whether SMB1 or SMB2 was
	// negotiated. The current tree must be IPC$.
	RPCTransport(pipeName string) (dcerpctransport.Transport, error)

	// TreeDisconnect disconnects the current tree.
	TreeDisconnect() error

	// Logoff tears down the authenticated session.
	Logoff() error

	// Disconnect closes the underlying transport connection.
	Disconnect() error
}
