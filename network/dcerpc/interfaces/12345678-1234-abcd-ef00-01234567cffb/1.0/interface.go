// Package rpcinterface_123456781234abcdef0001234567cffb_1_0 is the descriptor for the
// Netlogon (logon) RPC interface, abstract syntax 12345678-1234-abcd-ef00-01234567cffb
// version 1.0 ([MS-NRPC]).
//
// An RPC interface is identified by its UUID and version, never by the named pipe it is
// reached over. This package holds only the interface-level descriptor (abstract syntax,
// transport endpoint, opnums, opnum<->name maps, status and negotiate-flag constants). NDR
// types live in windows/protocols/ms-nrpc and method stubs in functions; both depend on
// this package, never the reverse.
package rpcinterface_123456781234abcdef0001234567cffb_1_0

// IDL source: [MS-NRPC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7
// A fetched copy is kept at ms-nrpc.idl in the interface directory.

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the IPC$-relative named pipe for the Netlogon interface ([MS-NRPC] 2.1: the
// Netlogon server listens on the \PIPE\netlogon endpoint over ncacn_np). Netlogon is also
// reachable over ncacn_ip_tcp at a dynamic port resolved through the endpoint mapper.
const PipeName = `\netlogon`

// Opnums for the on-the-wire methods. Opnums 50, 51, 52, 53, 54, 55, 56, 57, 58 are "not used on the wire"
// and are omitted.
const (
	OpnumNetrLogonUasLogon                   uint16 = 0
	OpnumNetrLogonUasLogoff                  uint16 = 1
	OpnumNetrLogonSamLogon                   uint16 = 2
	OpnumNetrLogonSamLogoff                  uint16 = 3
	OpnumNetrServerReqChallenge              uint16 = 4
	OpnumNetrServerAuthenticate              uint16 = 5
	OpnumNetrServerPasswordSet               uint16 = 6
	OpnumNetrDatabaseDeltas                  uint16 = 7
	OpnumNetrDatabaseSync                    uint16 = 8
	OpnumNetrAccountDeltas                   uint16 = 9
	OpnumNetrAccountSync                     uint16 = 10
	OpnumNetrGetDCName                       uint16 = 11
	OpnumNetrLogonControl                    uint16 = 12
	OpnumNetrGetAnyDCName                    uint16 = 13
	OpnumNetrLogonControl2                   uint16 = 14
	OpnumNetrServerAuthenticate2             uint16 = 15
	OpnumNetrDatabaseSync2                   uint16 = 16
	OpnumNetrDatabaseRedo                    uint16 = 17
	OpnumNetrLogonControl2Ex                 uint16 = 18
	OpnumNetrEnumerateTrustedDomains         uint16 = 19
	OpnumDsrGetDcName                        uint16 = 20
	OpnumNetrLogonGetCapabilities            uint16 = 21
	OpnumNetrLogonSetServiceBits             uint16 = 22
	OpnumNetrLogonGetTrustRid                uint16 = 23
	OpnumNetrLogonComputeServerDigest        uint16 = 24
	OpnumNetrLogonComputeClientDigest        uint16 = 25
	OpnumNetrServerAuthenticate3             uint16 = 26
	OpnumDsrGetDcNameEx                      uint16 = 27
	OpnumDsrGetSiteName                      uint16 = 28
	OpnumNetrLogonGetDomainInfo              uint16 = 29
	OpnumNetrServerPasswordSet2              uint16 = 30
	OpnumNetrServerPasswordGet               uint16 = 31
	OpnumNetrLogonSendToSam                  uint16 = 32
	OpnumDsrAddressToSiteNamesW              uint16 = 33
	OpnumDsrGetDcNameEx2                     uint16 = 34
	OpnumNetrLogonGetTimeServiceParentDomain uint16 = 35
	OpnumNetrEnumerateTrustedDomainsEx       uint16 = 36
	OpnumDsrAddressToSiteNamesExW            uint16 = 37
	OpnumDsrGetDcSiteCoverageW               uint16 = 38
	OpnumNetrLogonSamLogonEx                 uint16 = 39
	OpnumDsrEnumerateDomainTrusts            uint16 = 40
	OpnumDsrDeregisterDnsHostRecords         uint16 = 41
	OpnumNetrServerTrustPasswordsGet         uint16 = 42
	OpnumDsrGetForestTrustInformation        uint16 = 43
	OpnumNetrGetForestTrustInformation       uint16 = 44
	OpnumNetrLogonSamLogonWithFlags          uint16 = 45
	OpnumNetrServerGetTrustInfo              uint16 = 46
	OpnumOpnumUnused47                       uint16 = 47
	OpnumDsrUpdateReadOnlyServerDnsRecords   uint16 = 48
	OpnumNetrChainSetClientAttributes        uint16 = 49
	OpnumNetrServerAuthenticateKerberos      uint16 = 59
)

// NTSTATUS values returned by Netlogon methods ([MS-ERREF] 2.3). Netlogon returns NTSTATUS
// (not the Win32 error_status_t that MS-RRP uses). StatusSuccess is the canonical success
// value; StatusAccessDenied is what the server returns for a rejected secure-channel
// credential.
const (
	StatusSuccess           uint32 = 0x00000000 // STATUS_SUCCESS
	StatusInvalidParameter  uint32 = 0xC000000D // STATUS_INVALID_PARAMETER
	StatusAccessDenied      uint32 = 0xC0000022 // STATUS_ACCESS_DENIED
	StatusNoSuchUser        uint32 = 0xC0000064 // STATUS_NO_SUCH_USER
	StatusNoTrustSAMAccount uint32 = 0xC000018B // STATUS_NO_TRUST_SAM_ACCOUNT
	StatusNotSupported      uint32 = 0xC00000BB // STATUS_NOT_SUPPORTED
	StatusMoreEntries       uint32 = 0x00000105 // STATUS_MORE_ENTRIES
	StatusNoMoreEntries     uint32 = 0x8000001A // STATUS_NO_MORE_ENTRIES
	StatusDowngradeDetected uint32 = 0xC0000388 // STATUS_DOWNGRADE_DETECTED
)

// SyntaxID returns the logon abstract syntax identifier:
// 12345678-1234-abcd-ef00-01234567cffb, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x12345678, B: 0x1234, C: 0xabcd, D: 0xef00, E: 0x01234567cffb},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

// StatusString returns a mnemonic for the documented status codes, otherwise the
// hex value.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "STATUS_SUCCESS"
	case StatusInvalidParameter:
		return "STATUS_INVALID_PARAMETER"
	case StatusAccessDenied:
		return "STATUS_ACCESS_DENIED"
	case StatusNoSuchUser:
		return "STATUS_NO_SUCH_USER"
	case StatusNoTrustSAMAccount:
		return "STATUS_NO_TRUST_SAM_ACCOUNT"
	case StatusNotSupported:
		return "STATUS_NOT_SUPPORTED"
	case StatusMoreEntries:
		return "STATUS_MORE_ENTRIES"
	case StatusNoMoreEntries:
		return "STATUS_NO_MORE_ENTRIES"
	case StatusDowngradeDetected:
		return "STATUS_DOWNGRADE_DETECTED"
	default:
		return fmt.Sprintf("0x%08x", status)
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumNetrLogonUasLogon:                   "NetrLogonUasLogon",
	OpnumNetrLogonUasLogoff:                  "NetrLogonUasLogoff",
	OpnumNetrLogonSamLogon:                   "NetrLogonSamLogon",
	OpnumNetrLogonSamLogoff:                  "NetrLogonSamLogoff",
	OpnumNetrServerReqChallenge:              "NetrServerReqChallenge",
	OpnumNetrServerAuthenticate:              "NetrServerAuthenticate",
	OpnumNetrServerPasswordSet:               "NetrServerPasswordSet",
	OpnumNetrDatabaseDeltas:                  "NetrDatabaseDeltas",
	OpnumNetrDatabaseSync:                    "NetrDatabaseSync",
	OpnumNetrAccountDeltas:                   "NetrAccountDeltas",
	OpnumNetrAccountSync:                     "NetrAccountSync",
	OpnumNetrGetDCName:                       "NetrGetDCName",
	OpnumNetrLogonControl:                    "NetrLogonControl",
	OpnumNetrGetAnyDCName:                    "NetrGetAnyDCName",
	OpnumNetrLogonControl2:                   "NetrLogonControl2",
	OpnumNetrServerAuthenticate2:             "NetrServerAuthenticate2",
	OpnumNetrDatabaseSync2:                   "NetrDatabaseSync2",
	OpnumNetrDatabaseRedo:                    "NetrDatabaseRedo",
	OpnumNetrLogonControl2Ex:                 "NetrLogonControl2Ex",
	OpnumNetrEnumerateTrustedDomains:         "NetrEnumerateTrustedDomains",
	OpnumDsrGetDcName:                        "DsrGetDcName",
	OpnumNetrLogonGetCapabilities:            "NetrLogonGetCapabilities",
	OpnumNetrLogonSetServiceBits:             "NetrLogonSetServiceBits",
	OpnumNetrLogonGetTrustRid:                "NetrLogonGetTrustRid",
	OpnumNetrLogonComputeServerDigest:        "NetrLogonComputeServerDigest",
	OpnumNetrLogonComputeClientDigest:        "NetrLogonComputeClientDigest",
	OpnumNetrServerAuthenticate3:             "NetrServerAuthenticate3",
	OpnumDsrGetDcNameEx:                      "DsrGetDcNameEx",
	OpnumDsrGetSiteName:                      "DsrGetSiteName",
	OpnumNetrLogonGetDomainInfo:              "NetrLogonGetDomainInfo",
	OpnumNetrServerPasswordSet2:              "NetrServerPasswordSet2",
	OpnumNetrServerPasswordGet:               "NetrServerPasswordGet",
	OpnumNetrLogonSendToSam:                  "NetrLogonSendToSam",
	OpnumDsrAddressToSiteNamesW:              "DsrAddressToSiteNamesW",
	OpnumDsrGetDcNameEx2:                     "DsrGetDcNameEx2",
	OpnumNetrLogonGetTimeServiceParentDomain: "NetrLogonGetTimeServiceParentDomain",
	OpnumNetrEnumerateTrustedDomainsEx:       "NetrEnumerateTrustedDomainsEx",
	OpnumDsrAddressToSiteNamesExW:            "DsrAddressToSiteNamesExW",
	OpnumDsrGetDcSiteCoverageW:               "DsrGetDcSiteCoverageW",
	OpnumNetrLogonSamLogonEx:                 "NetrLogonSamLogonEx",
	OpnumDsrEnumerateDomainTrusts:            "DsrEnumerateDomainTrusts",
	OpnumDsrDeregisterDnsHostRecords:         "DsrDeregisterDnsHostRecords",
	OpnumNetrServerTrustPasswordsGet:         "NetrServerTrustPasswordsGet",
	OpnumDsrGetForestTrustInformation:        "DsrGetForestTrustInformation",
	OpnumNetrGetForestTrustInformation:       "NetrGetForestTrustInformation",
	OpnumNetrLogonSamLogonWithFlags:          "NetrLogonSamLogonWithFlags",
	OpnumNetrServerGetTrustInfo:              "NetrServerGetTrustInfo",
	OpnumOpnumUnused47:                       "OpnumUnused47",
	OpnumDsrUpdateReadOnlyServerDnsRecords:   "DsrUpdateReadOnlyServerDnsRecords",
	OpnumNetrChainSetClientAttributes:        "NetrChainSetClientAttributes",
	OpnumNetrServerAuthenticateKerberos:      "NetrServerAuthenticateKerberos",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()

// Netlogon negotiate option bits ([MS-NRPC] 3.1.4.2), one per capability letter (A–Z). A
// client advertises the set it supports in the NegotiateFlags argument of
// NetrServerAuthenticate2/3; the server echoes back the subset it agrees to.
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
