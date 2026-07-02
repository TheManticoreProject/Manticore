// Package rpcinterface_123456781234abcdef0001234567cffb_1_0 is the descriptor for the
// Netlogon RPC interface, abstract syntax 12345678-1234-abcd-ef00-01234567cffb version 1.0
// ([MS-NRPC]).
//
// Hand-written (not idlgen-generated): only the slice of MS-NRPC needed to establish and
// use a Netlogon secure channel is modelled — NetrServerReqChallenge (4),
// NetrServerAuthenticate2 (15) and NetrServerPasswordSet2 (30) — so the opnum table below
// is intentionally partial and documents just the implemented methods.
package rpcinterface_123456781234abcdef0001234567cffb_1_0

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the IPC$-relative named pipe for the Netlogon interface ([MS-NRPC] 2.1: the
// Netlogon server listens on the \PIPE\netlogon endpoint over ncacn_np). Netlogon is also
// reachable over ncacn_ip_tcp at a dynamic port resolved through the endpoint mapper.
const PipeName = `\netlogon`

// Opnums for the implemented on-the-wire methods ([MS-NRPC] 3.5.4.4).
const (
	OpnumNetrServerReqChallenge  uint16 = 4
	OpnumNetrServerAuthenticate2 uint16 = 15
	OpnumNetrServerPasswordSet2  uint16 = 30
)

// OpnumToName maps each implemented opnum to its method name; the single source of truth
// for the (partial) wire numbering of this interface.
var OpnumToName = map[uint16]string{
	OpnumNetrServerReqChallenge:  "NetrServerReqChallenge",
	OpnumNetrServerAuthenticate2: "NetrServerAuthenticate2",
	OpnumNetrServerPasswordSet2:  "NetrServerPasswordSet2",
}

// NameToOpnum is the inverse of OpnumToName, derived at init.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()

// NTSTATUS values returned by Netlogon methods ([MS-ERREF] 2.3). Netlogon returns NTSTATUS
// (not the Win32 error_status_t that MS-RRP uses). StatusSuccess is the canonical success
// value; StatusAccessDenied is what the server returns for a rejected secure-channel
// credential.
const (
	StatusSuccess            uint32 = 0x00000000 // STATUS_SUCCESS
	StatusInvalidParameter   uint32 = 0xC000000D // STATUS_INVALID_PARAMETER
	StatusAccessDenied       uint32 = 0xC0000022 // STATUS_ACCESS_DENIED
	StatusNoTrustSAMAccount  uint32 = 0xC000018B // STATUS_NO_TRUST_SAM_ACCOUNT
	StatusNoSuchUser         uint32 = 0xC0000064 // STATUS_NO_SUCH_USER
	StatusNotSupported       uint32 = 0xC00000BB // STATUS_NOT_SUPPORTED
	StatusDowngradeDetected  uint32 = 0xC0000388 // STATUS_DOWNGRADE_DETECTED
)

// StatusString returns a mnemonic for the documented status codes, otherwise the hex value.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "STATUS_SUCCESS"
	case StatusInvalidParameter:
		return "STATUS_INVALID_PARAMETER"
	case StatusAccessDenied:
		return "STATUS_ACCESS_DENIED"
	case StatusNoTrustSAMAccount:
		return "STATUS_NO_TRUST_SAM_ACCOUNT"
	case StatusNoSuchUser:
		return "STATUS_NO_SUCH_USER"
	case StatusNotSupported:
		return "STATUS_NOT_SUPPORTED"
	case StatusDowngradeDetected:
		return "STATUS_DOWNGRADE_DETECTED"
	default:
		return fmt.Sprintf("0x%08x", status)
	}
}

// NETLOGON_SECURE_CHANNEL_TYPE ([MS-NRPC] 2.2.1.3.13) identifies the kind of secure channel
// being established. It is an NDR enum (2 octets under NDR20); use the `ndr:"enum"` tag on
// fields of this type.
type NETLOGON_SECURE_CHANNEL_TYPE uint16

const (
	NullSecureChannel             NETLOGON_SECURE_CHANNEL_TYPE = 0
	MsvApSecureChannel            NETLOGON_SECURE_CHANNEL_TYPE = 1
	WorkstationSecureChannel      NETLOGON_SECURE_CHANNEL_TYPE = 2
	TrustedDnsDomainSecureChannel NETLOGON_SECURE_CHANNEL_TYPE = 3
	TrustedDomainSecureChannel    NETLOGON_SECURE_CHANNEL_TYPE = 4
	UasServerSecureChannel        NETLOGON_SECURE_CHANNEL_TYPE = 5
	ServerSecureChannel           NETLOGON_SECURE_CHANNEL_TYPE = 6
	CdcServerSecureChannel        NETLOGON_SECURE_CHANNEL_TYPE = 7
)

// Netlogon negotiate option bits ([MS-NRPC] 3.1.4.2), one per capability letter (A–Z).
const (
	NegotiateAccountLockout             uint32 = 0x00000001 // A
	NegotiateNT35BDCContinuousUpdate    uint32 = 0x00000002 // B
	NegotiateRC4                        uint32 = 0x00000004 // C
	NegotiatePromotionCount             uint32 = 0x00000008 // D
	NegotiateBDCHandlingChangelogs      uint32 = 0x00000010 // E
	NegotiateRestartFullDCSync          uint32 = 0x00000020 // F
	NegotiateNoValidationLevel2         uint32 = 0x00000040 // G
	NegotiateDatabaseRedo               uint32 = 0x00000080 // H
	NegotiateRefusePasswordChange       uint32 = 0x00000100 // I
	NegotiateSendToSAM                  uint32 = 0x00000200 // J
	NegotiateGenericPassThrough         uint32 = 0x00000400 // K
	NegotiateConcurrentRPC              uint32 = 0x00000800 // L
	NegotiateAvoidUserAccountDBRepl     uint32 = 0x00001000 // M
	NegotiateAvoidSecurityAuthorityRepl uint32 = 0x00002000 // N
	NegotiateStrongKeys                 uint32 = 0x00004000 // O
	NegotiateTransitiveTrusts           uint32 = 0x00008000 // P
	NegotiateDNSTrusts                  uint32 = 0x00010000 // Q
	NegotiatePasswordSet2               uint32 = 0x00020000 // R
	NegotiateGetDomainInfo              uint32 = 0x00040000 // S
	NegotiateCrossForestTrusts          uint32 = 0x00080000 // T
	NegotiateNoNT4Emulation             uint32 = 0x00100000 // U
	NegotiateRODCPassThrough            uint32 = 0x00200000 // V
	NegotiateAES                        uint32 = 0x01000000 // W (AES-128 CFB8 + SHA2)
	NegotiateAuthenticatedRPCViaLSASS   uint32 = 0x20000000 // X
	NegotiateSecureRPC                  uint32 = 0x40000000 // Y
	NegotiateKerberosForSecureChannel   uint32 = 0x80000000 // Z
)

// SyntaxID returns the Netlogon abstract syntax identifier:
// 12345678-1234-abcd-ef00-01234567cffb, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x12345678, B: 0x1234, C: 0xabcd, D: 0xef00, E: 0x01234567cffb},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}
