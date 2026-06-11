package client

import (
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
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
	// DialTimeout bounds each connect/negotiate attempt during Dial, so that a
	// server which silently drops a probe (e.g. an SMB1-only host receiving an
	// SMB2 negotiate) fails the attempt instead of blocking forever, letting
	// PolicyStrictOrder move on to the next preferred version. The bound is
	// lifted once negotiation completes, so it does not constrain long-blocking
	// operations such as CHANGE_NOTIFY. Zero applies DefaultDialTimeout; a
	// negative value disables the bound.
	DialTimeout time.Duration
}

// DefaultDialTimeout is the per-attempt connect/negotiate bound applied when
// Options.DialTimeout is zero.
const DefaultDialTimeout = 10 * time.Second

// dialTimeout resolves the effective per-attempt bound from the options: zero
// means DefaultDialTimeout, negative means unbounded.
func (o Options) dialTimeout() time.Duration {
	switch {
	case o.DialTimeout < 0:
		return 0
	case o.DialTimeout == 0:
		return DefaultDialTimeout
	}
	return o.DialTimeout
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

// OpenOptions carries the file-open parameters shared by every dialect. Combine
// the bit values from windows/fileflags with bitwise OR, e.g.
// DesiredAccess: fileflags.GENERIC_READ | fileflags.FILE_READ_ATTRIBUTES.
type OpenOptions struct {
	DesiredAccess     uint32 // fileflags GENERIC_* / FILE_* access rights
	ShareAccess       uint32 // fileflags FILE_SHARE_*
	CreateDisposition uint32 // fileflags FILE_OPEN / FILE_CREATE / ...
	CreateOptions     uint32 // fileflags FILE_DIRECTORY_FILE / FILE_NON_DIRECTORY_FILE / ...
	FileAttributes    uint32 // fileflags FILE_ATTRIBUTE_*
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
	return fi.FileAttributes&fileflags.FILE_ATTRIBUTE_DIRECTORY != 0
}

// ConnectionInfo holds the connection capabilities negotiated with the server,
// normalized across dialects. SMB1 negotiates a single MaxBufferSize rather than
// separate read/write maxima, so on SMB1 both MaxReadSize and MaxWriteSize report
// that value.
type ConnectionInfo struct {
	// SigningRequired reports whether the server requires SMB message signing.
	SigningRequired bool
	// MaxReadSize is the largest read the server will accept, in bytes.
	MaxReadSize uint32
	// MaxWriteSize is the largest write the server will accept, in bytes.
	MaxWriteSize uint32
	// SupportsNTLMv2 reports whether the server advertised NTLM2 / extended session
	// security (NTLMv2) in its NTLM CHALLENGE. Meaningful only after Login.
	SupportsNTLMv2 bool
}

// ServerIdentity is the server identity learned during NTLM authentication: the
// NetBIOS and DNS computer/domain names advertised in the CHALLENGE TargetInfo,
// and the OS version from the CHALLENGE Version field. Fields are empty/zero when
// the server did not advertise them or when authentication has not occurred.
type ServerIdentity struct {
	NetBIOSComputerName string
	NetBIOSDomainName   string
	DNSComputerName     string
	DNSDomainName       string
	OSVersionMajor      uint8
	OSVersionMinor      uint8
	OSVersionBuild      uint16
	// OSName is the server's native OS string from the SMB1 SESSION_SETUP response.
	// It is populated for SMB1 only; SMB2/SMB3 do not transmit an OS name on the wire
	// (use the OSVersion* fields there).
	OSName string
}

// SecurityInformation selects which parts of a SECURITY_DESCRIPTOR a
// QuerySecurityDescriptor call retrieves. The values are the SECURITY_INFORMATION
// bits from [MS-DTYP] 2.4.7; combine them with bitwise OR.
type SecurityInformation uint32

const (
	// OwnerSecurityInformation requests the owner SID.
	OwnerSecurityInformation SecurityInformation = 0x00000001
	// GroupSecurityInformation requests the primary-group SID.
	GroupSecurityInformation SecurityInformation = 0x00000002
	// DaclSecurityInformation requests the discretionary ACL.
	DaclSecurityInformation SecurityInformation = 0x00000004
	// SaclSecurityInformation requests the system ACL (requires extra privilege).
	SaclSecurityInformation SecurityInformation = 0x00000008
)
