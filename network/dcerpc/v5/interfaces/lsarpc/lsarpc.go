// Package lsarpc implements a minimal client for the Local Security Authority
// (Domain Policy) Remote Protocol interface (lsarpc, [MS-LSAD] / [MS-LSAT]).
//
// It is the first concrete DCE/RPC interface in the tree and exists to validate the
// whole stack end to end: SMB session -> named pipe -> DCE/RPC bind -> call. Only the
// two calls needed for that proof are implemented, LsarOpenPolicy2 (the mandatory
// first call that returns a policy handle) and LsarClose. Their NDR is hand-marshalled
// rather than produced by a generic NDR engine.
//
// References:
//   - [MS-LSAD] Local Security Authority (Domain Policy) Remote Protocol:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/1b5471ef-4c33-4a91-b079-dfcbb82f05cc
//   - [MS-LSAD] LsarOpenPolicy2 (Opnum 44):
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/9456a963-7c21-4710-af77-d0a2f5a72d6b
//   - [MS-LSAD] LsarClose (Opnum 0):
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/a0ad7d8b-bb3b-4d3d-8c2e-e0c5b3a2a8b4
//   - [MS-LSAD] 2.2.2.1 LSAPR_HANDLE (RPC context handle):
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/0d093105-e8c8-45f7-a79d-182aafd60c6e
//   - [MS-LSAD] 2.2.1.1.2 ACCESS_MASK for policy objects:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/b61b7268-987a-420b-84f9-6c75f8dc8558
package lsarpc

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the named pipe to open for the lsarpc interface. When the SMB tree is
// connected to IPC$, the pipe is opened by its name relative to that share (the
// "\PIPE" namespace is implicit), so the NT_CREATE_ANDX FileName is "\lsarpc" rather
// than the MS-RPCE well-known endpoint "\pipe\lsarpc".
const PipeName = `\lsarpc`

// Opnums for the implemented methods ([MS-LSAD] section 3.1.4).
const (
	OpnumClose       uint16 = 0
	OpnumOpenPolicy2 uint16 = 44
)

// Policy object access rights ([MS-LSAD] section 2.2.1.1.2), plus the generic
// MAXIMUM_ALLOWED bit ([MS-DTYP] ACCESS_MASK).
const (
	PolicyViewLocalInformation  uint32 = 0x00000001
	PolicyViewAuditInformation  uint32 = 0x00000002
	PolicyGetPrivateInformation uint32 = 0x00000004
	PolicyTrustAdmin            uint32 = 0x00000008
	PolicyCreateAccount         uint32 = 0x00000010
	PolicyLookupNames           uint32 = 0x00000800
	MaximumAllowed              uint32 = 0x02000000
)

// Common NTSTATUS codes returned by these methods ([MS-LSAD] return values).
const (
	StatusSuccess          uint32 = 0x00000000
	StatusAccessDenied     uint32 = 0xC0000022
	StatusInvalidParameter uint32 = 0xC000000D
)

// SyntaxID returns the lsarpc abstract syntax identifier:
// 12345778-1234-abcd-ef00-0123456789ab, version 0.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x12345778, B: 0x1234, C: 0xabcd, D: 0xef00, E: 0x0123456789ab},
		MajorVersion: 0,
		MinorVersion: 0,
	}
}

// PolicyHandle is an RPC context handle (LSAPR_HANDLE): a 4-byte attributes field
// followed by a 16-byte GUID, 20 bytes total ([MS-RPCE] 2.3.2.2, [MS-LSAD] 2.2.2.1).
type PolicyHandle [20]byte

// IsZero reports whether the handle is all zeros (for example, after a successful
// Close).
func (h PolicyHandle) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}

// objectAttributes models LSAPR_OBJECT_ATTRIBUTES ([MS-LSAD] 2.2.2.3). All fields are
// ignored by the server except RootDirectory, which must be NULL, so each is modeled
// as a 4-octet zero field (the four pointer members as NULL referents).
type objectAttributes struct {
	Length                   ndr.DWORD
	RootDirectory            ndr.DWORD // [unique] NULL
	ObjectName               ndr.DWORD // [unique] NULL
	Attributes               ndr.DWORD
	SecurityDescriptor       ndr.DWORD // [unique] NULL
	SecurityQualityOfService ndr.DWORD // [unique] NULL
}

// openPolicy2Request is the [in] parameter set of LsarOpenPolicy2: a NULL unique
// SystemName pointer, an inline ObjectAttributes (a top-level [ref] struct), and the
// desired access mask.
type openPolicy2Request struct {
	SystemName    *ndr.WSTR `ndr:"unique"`
	Attributes    objectAttributes
	DesiredAccess ndr.DWORD
}

func (*openPolicy2Request) Opnum() uint16 { return OpnumOpenPolicy2 }

// closeRequest is the [in,out] parameter of LsarClose: the context handle.
type closeRequest struct {
	Handle PolicyHandle
}

func (*closeRequest) Opnum() uint16 { return OpnumClose }

// handleResponse is the common reply shape: a 20-byte context handle followed by the
// NTSTATUS return value.
type handleResponse struct {
	Handle PolicyHandle
	Status ndr.DWORD
}

// OpenPolicy2 calls LsarOpenPolicy2 (opnum 44) and returns a policy handle.
func OpenPolicy2(rpc *client.Client, desiredAccess uint32) (PolicyHandle, error) {
	req := &openPolicy2Request{DesiredAccess: ndr.DWORD(desiredAccess)}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return PolicyHandle{}, fmt.Errorf("LsarOpenPolicy2: %w", err)
	}
	if uint32(resp.Status) != StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarOpenPolicy2 failed: %s", StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}

// Close calls LsarClose (opnum 0) on a handle. On success the server returns a zeroed
// handle, which is returned to the caller.
func Close(rpc *client.Client, handle PolicyHandle) (PolicyHandle, error) {
	req := &closeRequest{Handle: handle}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return PolicyHandle{}, fmt.Errorf("LsarClose: %w", err)
	}
	if uint32(resp.Status) != StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarClose failed: %s", StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}

// StatusString returns a mnemonic for the documented status codes, otherwise the hex
// value.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "STATUS_SUCCESS"
	case StatusAccessDenied:
		return "STATUS_ACCESS_DENIED"
	case StatusInvalidParameter:
		return "STATUS_INVALID_PARAMETER"
	default:
		return fmt.Sprintf("0x%08x", status)
	}
}
