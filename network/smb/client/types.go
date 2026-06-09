package client

import (
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb"
)

// NegotiationPolicy controls how the client applies its preference list when
// selecting a dialect during Dial.
type NegotiationPolicy int

const (
	// PolicyStrictOrder tries the preferred versions in order and selects the
	// first one the server accepts. It honors the caller's order exactly — it
	// can select a lower dialect (e.g. SMB1) even when the server also supports
	// a higher one. Out-of-order lists may cost a reconnect per attempt.
	PolicyStrictOrder NegotiationPolicy = iota
	// PolicyHighestInSet performs a single multi-protocol negotiate over the
	// whole preference set and uses the server's highest-supported dialect
	// within it. One round-trip; the list acts as an allow-list rather than a
	// strict order.
	PolicyHighestInSet
)

// Options configures a Dial.
type Options struct {
	// Preferred is the ordered list of protocol versions, highest priority
	// first. When empty, the client offers all supported versions, best first.
	Preferred []smb.SMBProtocolVersion
	// Policy selects how Preferred is applied. The zero value is
	// PolicyStrictOrder.
	Policy NegotiationPolicy
	// Workstation is the NetBIOS workstation name sent during authentication.
	// When empty, a default is used.
	Workstation string
}

// FileHandle is an opaque, version-agnostic handle to an open file. The backend
// that produced it knows how to interpret the underlying value (an SMB1 FID or
// an SMB2 file id); callers must pass it back to that same backend.
type FileHandle struct {
	dialect smb.SMBProtocolVersion
	raw     any
}

// Dialect reports which protocol version's backend owns this handle.
func (h FileHandle) Dialect() smb.SMBProtocolVersion { return h.dialect }

// newFileHandle boxes a backend-specific handle value. Adapters use this; the
// raw value stays unexported so callers cannot tamper with it.
func newFileHandle(dialect smb.SMBProtocolVersion, raw any) FileHandle {
	return FileHandle{dialect: dialect, raw: raw}
}

// OpenOptions carries the file-open parameters shared by every dialect. The
// access, share, disposition, and options fields use the MS-DTYP/MS-SMB2 bit
// definitions, which are identical across SMB1 and SMB2.
type OpenOptions struct {
	DesiredAccess     uint32
	ShareAccess       uint32
	CreateDisposition uint32
	CreateOptions     uint32
	FileAttributes    uint32
}

// FileInfo is a version-agnostic directory entry, normalized from the SMB1
// TRANS2 FIND information levels and the SMB2 QUERY_DIRECTORY information
// classes. It is populated by ListDirectory (Phase 6).
type FileInfo struct {
	Name           string
	FileAttributes uint32
	Size           uint64
	AllocationSize uint64
	CreationTime   time.Time
	LastAccessTime time.Time
	LastWriteTime  time.Time
	ChangeTime     time.Time
}

// IsDir reports whether the entry has the FILE_ATTRIBUTE_DIRECTORY flag set.
func (fi FileInfo) IsDir() bool {
	const fileAttributeDirectory = 0x00000010
	return fi.FileAttributes&fileAttributeDirectory != 0
}
