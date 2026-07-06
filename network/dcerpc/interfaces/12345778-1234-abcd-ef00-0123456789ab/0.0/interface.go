// Package rpcinterface_123457781234abcdef000123456789ab_0_0 is the descriptor for the
// LSA Directory (lsarpc) RPC interface, abstract syntax
// 12345778-1234-abcd-ef00-0123456789ab version 0.0 ([MS-LSAD] / [MS-LSAT]).
//
// An RPC interface is identified by its UUID and version, never by the named pipe it
// happens to be reached over: the "\lsarpc" pipe multiplexes several distinct
// interfaces, so the directory is named after the interface UUID (with the version in
// the nested 0.0/ directory) rather than after the pipe.
//
// This package holds only the interface-level descriptor: the abstract syntax
// identifier, the transport endpoint (PipeName), the opnum constants, and the status
// and access-mask values shared across methods. The NDR types live in the structures
// subpackage and the method stubs in the functions subpackage; both depend on this
// package, never the reverse.
//
// References:
//   - [MS-LSAD] Local Security Authority (Domain Policy) Remote Protocol:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/1b5471ef-4c33-4a91-b079-dfcbb82f05cc
//   - [MS-LSAD] LsarOpenPolicy2 (Opnum 44):
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/9456a963-7c21-4710-af77-d0a2f5a72d6b
//   - [MS-LSAD] LsarClose (Opnum 0):
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/a0ad7d8b-bb3b-4d3d-8c2e-e0c5b3a2a8b4
package rpcinterface_123457781234abcdef000123456789ab_0_0

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the named pipe to open for the lsarpc interface. When the SMB tree is
// connected to IPC$, the pipe is opened by its name relative to that share (the
// "\PIPE" namespace is implicit), so the NT_CREATE_ANDX FileName is "\lsarpc" rather
// than the MS-RPCE well-known endpoint "\pipe\lsarpc".
const PipeName = `\lsarpc`

// Opnums for the interface methods ([MS-LSAD] / [MS-LSAT] section 3.1.4). Opnums not
// listed here are "not used on the wire" in the IDL.
const (
	OpnumLsarClose                          uint16 = 0
	OpnumLsarEnumeratePrivileges            uint16 = 2
	OpnumLsarQuerySecurityObject            uint16 = 3
	OpnumLsarSetSecurityObject              uint16 = 4
	OpnumLsarOpenPolicy                     uint16 = 6
	OpnumLsarQueryInformationPolicy         uint16 = 7
	OpnumLsarSetInformationPolicy           uint16 = 8
	OpnumLsarCreateAccount                  uint16 = 10
	OpnumLsarEnumerateAccounts              uint16 = 11
	OpnumLsarCreateTrustedDomain            uint16 = 12
	OpnumLsarEnumerateTrustedDomains        uint16 = 13
	OpnumLsarLookupNames                    uint16 = 14
	OpnumLsarLookupSids                     uint16 = 15
	OpnumLsarCreateSecret                   uint16 = 16
	OpnumLsarOpenAccount                    uint16 = 17
	OpnumLsarEnumeratePrivilegesAccount     uint16 = 18
	OpnumLsarAddPrivilegesToAccount         uint16 = 19
	OpnumLsarRemovePrivilegesFromAccount    uint16 = 20
	OpnumLsarGetSystemAccessAccount         uint16 = 23
	OpnumLsarSetSystemAccessAccount         uint16 = 24
	OpnumLsarOpenTrustedDomain              uint16 = 25
	OpnumLsarQueryInfoTrustedDomain         uint16 = 26
	OpnumLsarSetInformationTrustedDomain    uint16 = 27
	OpnumLsarOpenSecret                     uint16 = 28
	OpnumLsarSetSecret                      uint16 = 29
	OpnumLsarQuerySecret                    uint16 = 30
	OpnumLsarLookupPrivilegeValue           uint16 = 31
	OpnumLsarLookupPrivilegeName            uint16 = 32
	OpnumLsarLookupPrivilegeDisplayName     uint16 = 33
	OpnumLsarDeleteObject                   uint16 = 34
	OpnumLsarEnumerateAccountsWithUserRight uint16 = 35
	OpnumLsarEnumerateAccountRights         uint16 = 36
	OpnumLsarAddAccountRights               uint16 = 37
	OpnumLsarRemoveAccountRights            uint16 = 38
	OpnumLsarQueryTrustedDomainInfo         uint16 = 39
	OpnumLsarSetTrustedDomainInfo           uint16 = 40
	OpnumLsarDeleteTrustedDomain            uint16 = 41
	OpnumLsarStorePrivateData               uint16 = 42
	OpnumLsarRetrievePrivateData            uint16 = 43
	OpnumLsarOpenPolicy2                    uint16 = 44
	OpnumLsarGetUserName                    uint16 = 45
	OpnumLsarQueryInformationPolicy2        uint16 = 46
	OpnumLsarSetInformationPolicy2          uint16 = 47
	OpnumLsarQueryTrustedDomainInfoByName   uint16 = 48
	OpnumLsarSetTrustedDomainInfoByName     uint16 = 49
	OpnumLsarEnumerateTrustedDomainsEx      uint16 = 50
	OpnumLsarCreateTrustedDomainEx          uint16 = 51
	OpnumLsarQueryDomainInformationPolicy   uint16 = 53
	OpnumLsarSetDomainInformationPolicy     uint16 = 54
	OpnumLsarOpenTrustedDomainByName        uint16 = 55
	OpnumLsarLookupSids2                    uint16 = 57
	OpnumLsarLookupNames2                   uint16 = 58
	OpnumLsarCreateTrustedDomainEx2         uint16 = 59
	OpnumLsarLookupNames3                   uint16 = 68
	OpnumLsarQueryForestTrustInformation    uint16 = 73
	OpnumLsarSetForestTrustInformation      uint16 = 74
	OpnumLsarLookupSids3                    uint16 = 76
	OpnumLsarLookupNames4                   uint16 = 77
	OpnumLsarCreateTrustedDomainEx3         uint16 = 129
	OpnumLsarOpenPolicy3                    uint16 = 130
	OpnumLsarQueryForestTrustInformation2   uint16 = 132
	OpnumLsarSetForestTrustInformation2     uint16 = 133
	OpnumLsarOpenPolicyWithCreds            uint16 = 135
	OpnumLsarOpenSecret2                    uint16 = 136
	OpnumLsarCreateSecret2                  uint16 = 137
	OpnumLsarSetSecret2                     uint16 = 138
	OpnumLsarQuerySecret2                   uint16 = 139
	OpnumLsarStorePrivateData2              uint16 = 140
	OpnumLsarRetrievePrivateData2           uint16 = 141
)

// Standard access rights ([MS-DTYP] 2.4.3 ACCESS_MASK), plus the generic
// MAXIMUM_ALLOWED bit.
const (
	Delete          uint32 = 0x00010000
	ReadControl     uint32 = 0x00020000
	WriteDac        uint32 = 0x00040000
	WriteOwner      uint32 = 0x00080000
	MaximumAllowed  uint32 = 0x02000000
	AccessSystemSec uint32 = 0x01000000
	GenericAll      uint32 = 0x10000000
	GenericExecute  uint32 = 0x20000000
	GenericWrite    uint32 = 0x40000000
	GenericRead     uint32 = 0x80000000
)

// Policy object access rights ([MS-LSAD] section 2.2.1.1.2).
const (
	PolicyViewLocalInformation  uint32 = 0x00000001
	PolicyViewAuditInformation  uint32 = 0x00000002
	PolicyGetPrivateInformation uint32 = 0x00000004
	PolicyTrustAdmin            uint32 = 0x00000008
	PolicyCreateAccount         uint32 = 0x00000010
	PolicyCreateSecret          uint32 = 0x00000020
	PolicyCreatePrivilege       uint32 = 0x00000040
	PolicySetDefaultQuotaLimits uint32 = 0x00000080
	PolicySetAuditRequirements  uint32 = 0x00000100
	PolicyAuditLogAdmin         uint32 = 0x00000200
	PolicyServerAdmin           uint32 = 0x00000400
	PolicyLookupNames           uint32 = 0x00000800
	PolicyNotification          uint32 = 0x00001000
)

// Account / secret / trusted-domain object access rights ([MS-LSAD] 2.2.1.1.3–2.2.1.1.5).
const (
	AccountViewAccount        uint32 = 0x00000001
	AccountAdjustPrivileges   uint32 = 0x00000002
	AccountAdjustQuotas       uint32 = 0x00000004
	AccountAdjustSystemAccess uint32 = 0x00000008
	SecretSetValue            uint32 = 0x00000001
	SecretQueryValue          uint32 = 0x00000002
	TrustedQueryDomainName    uint32 = 0x00000001
	TrustedQueryControllers   uint32 = 0x00000002
	TrustedSetControllers     uint32 = 0x00000004
	TrustedQueryPosix         uint32 = 0x00000008
	TrustedSetPosix           uint32 = 0x00000010
	TrustedSetAuth            uint32 = 0x00000020
	TrustedQueryAuth          uint32 = 0x00000040
)

// Common NTSTATUS codes returned by these methods ([MS-LSAD] return values and
// [MS-ERREF] 2.3.1).
const (
	StatusSuccess            uint32 = 0x00000000
	StatusMoreEntries        uint32 = 0x00000105
	StatusNoMoreEntries      uint32 = 0x8000001A
	StatusObjectNameNotFound uint32 = 0xC0000034
	StatusAccessDenied       uint32 = 0xC0000022
	StatusInvalidParameter   uint32 = 0xC000000D
	StatusInvalidHandle      uint32 = 0xC0000008
	StatusNoSuchPrivilege    uint32 = 0xC0000060
	StatusNoSuchDomain       uint32 = 0xC00000DF
	StatusNoneMapped         uint32 = 0xC0000073
	StatusSomeNotMapped      uint32 = 0x00000107
	StatusInvalidSid         uint32 = 0xC0000078
	StatusNotSupported       uint32 = 0xC00000BB
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

// StatusString returns a mnemonic for the documented status codes, otherwise the hex
// value.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "STATUS_SUCCESS"
	case StatusMoreEntries:
		return "STATUS_MORE_ENTRIES"
	case StatusNoMoreEntries:
		return "STATUS_NO_MORE_ENTRIES"
	case StatusSomeNotMapped:
		return "STATUS_SOME_NOT_MAPPED"
	case StatusObjectNameNotFound:
		return "STATUS_OBJECT_NAME_NOT_FOUND"
	case StatusAccessDenied:
		return "STATUS_ACCESS_DENIED"
	case StatusInvalidParameter:
		return "STATUS_INVALID_PARAMETER"
	case StatusInvalidHandle:
		return "STATUS_INVALID_HANDLE"
	case StatusNoSuchPrivilege:
		return "STATUS_NO_SUCH_PRIVILEGE"
	case StatusNoSuchDomain:
		return "STATUS_NO_SUCH_DOMAIN"
	case StatusNoneMapped:
		return "STATUS_NONE_MAPPED"
	case StatusInvalidSid:
		return "STATUS_INVALID_SID"
	case StatusNotSupported:
		return "STATUS_NOT_SUPPORTED"
	default:
		return fmt.Sprintf("0x%08x", status)
	}
}

// OpnumToName maps each on-the-wire opnum to its [MS-LSAD]/[MS-LSAT] method name.
// Opnums that are "not used on the wire" in the IDL are absent. It is the single source
// of truth; NameToOpnum is derived from it.
var OpnumToName = map[uint16]string{
	OpnumLsarClose:                          "LsarClose",
	OpnumLsarEnumeratePrivileges:            "LsarEnumeratePrivileges",
	OpnumLsarQuerySecurityObject:            "LsarQuerySecurityObject",
	OpnumLsarSetSecurityObject:              "LsarSetSecurityObject",
	OpnumLsarOpenPolicy:                     "LsarOpenPolicy",
	OpnumLsarQueryInformationPolicy:         "LsarQueryInformationPolicy",
	OpnumLsarSetInformationPolicy:           "LsarSetInformationPolicy",
	OpnumLsarCreateAccount:                  "LsarCreateAccount",
	OpnumLsarEnumerateAccounts:              "LsarEnumerateAccounts",
	OpnumLsarCreateTrustedDomain:            "LsarCreateTrustedDomain",
	OpnumLsarEnumerateTrustedDomains:        "LsarEnumerateTrustedDomains",
	OpnumLsarLookupNames:                    "LsarLookupNames",
	OpnumLsarLookupSids:                     "LsarLookupSids",
	OpnumLsarCreateSecret:                   "LsarCreateSecret",
	OpnumLsarOpenAccount:                    "LsarOpenAccount",
	OpnumLsarEnumeratePrivilegesAccount:     "LsarEnumeratePrivilegesAccount",
	OpnumLsarAddPrivilegesToAccount:         "LsarAddPrivilegesToAccount",
	OpnumLsarRemovePrivilegesFromAccount:    "LsarRemovePrivilegesFromAccount",
	OpnumLsarGetSystemAccessAccount:         "LsarGetSystemAccessAccount",
	OpnumLsarSetSystemAccessAccount:         "LsarSetSystemAccessAccount",
	OpnumLsarOpenTrustedDomain:              "LsarOpenTrustedDomain",
	OpnumLsarQueryInfoTrustedDomain:         "LsarQueryInfoTrustedDomain",
	OpnumLsarSetInformationTrustedDomain:    "LsarSetInformationTrustedDomain",
	OpnumLsarOpenSecret:                     "LsarOpenSecret",
	OpnumLsarSetSecret:                      "LsarSetSecret",
	OpnumLsarQuerySecret:                    "LsarQuerySecret",
	OpnumLsarLookupPrivilegeValue:           "LsarLookupPrivilegeValue",
	OpnumLsarLookupPrivilegeName:            "LsarLookupPrivilegeName",
	OpnumLsarLookupPrivilegeDisplayName:     "LsarLookupPrivilegeDisplayName",
	OpnumLsarDeleteObject:                   "LsarDeleteObject",
	OpnumLsarEnumerateAccountsWithUserRight: "LsarEnumerateAccountsWithUserRight",
	OpnumLsarEnumerateAccountRights:         "LsarEnumerateAccountRights",
	OpnumLsarAddAccountRights:               "LsarAddAccountRights",
	OpnumLsarRemoveAccountRights:            "LsarRemoveAccountRights",
	OpnumLsarQueryTrustedDomainInfo:         "LsarQueryTrustedDomainInfo",
	OpnumLsarSetTrustedDomainInfo:           "LsarSetTrustedDomainInfo",
	OpnumLsarDeleteTrustedDomain:            "LsarDeleteTrustedDomain",
	OpnumLsarStorePrivateData:               "LsarStorePrivateData",
	OpnumLsarRetrievePrivateData:            "LsarRetrievePrivateData",
	OpnumLsarOpenPolicy2:                    "LsarOpenPolicy2",
	OpnumLsarGetUserName:                    "LsarGetUserName",
	OpnumLsarQueryInformationPolicy2:        "LsarQueryInformationPolicy2",
	OpnumLsarSetInformationPolicy2:          "LsarSetInformationPolicy2",
	OpnumLsarQueryTrustedDomainInfoByName:   "LsarQueryTrustedDomainInfoByName",
	OpnumLsarSetTrustedDomainInfoByName:     "LsarSetTrustedDomainInfoByName",
	OpnumLsarEnumerateTrustedDomainsEx:      "LsarEnumerateTrustedDomainsEx",
	OpnumLsarCreateTrustedDomainEx:          "LsarCreateTrustedDomainEx",
	OpnumLsarQueryDomainInformationPolicy:   "LsarQueryDomainInformationPolicy",
	OpnumLsarSetDomainInformationPolicy:     "LsarSetDomainInformationPolicy",
	OpnumLsarOpenTrustedDomainByName:        "LsarOpenTrustedDomainByName",
	OpnumLsarLookupSids2:                    "LsarLookupSids2",
	OpnumLsarLookupNames2:                   "LsarLookupNames2",
	OpnumLsarCreateTrustedDomainEx2:         "LsarCreateTrustedDomainEx2",
	OpnumLsarLookupNames3:                   "LsarLookupNames3",
	OpnumLsarQueryForestTrustInformation:    "LsarQueryForestTrustInformation",
	OpnumLsarSetForestTrustInformation:      "LsarSetForestTrustInformation",
	OpnumLsarLookupSids3:                    "LsarLookupSids3",
	OpnumLsarLookupNames4:                   "LsarLookupNames4",
	OpnumLsarCreateTrustedDomainEx3:         "LsarCreateTrustedDomainEx3",
	OpnumLsarOpenPolicy3:                    "LsarOpenPolicy3",
	OpnumLsarQueryForestTrustInformation2:   "LsarQueryForestTrustInformation2",
	OpnumLsarSetForestTrustInformation2:     "LsarSetForestTrustInformation2",
	OpnumLsarOpenPolicyWithCreds:            "LsarOpenPolicyWithCreds",
	OpnumLsarOpenSecret2:                    "LsarOpenSecret2",
	OpnumLsarCreateSecret2:                  "LsarCreateSecret2",
	OpnumLsarSetSecret2:                     "LsarSetSecret2",
	OpnumLsarQuerySecret2:                   "LsarQuerySecret2",
	OpnumLsarStorePrivateData2:              "LsarStorePrivateData2",
	OpnumLsarRetrievePrivateData2:           "LsarRetrievePrivateData2",
}

// NameToOpnum is the reverse of OpnumToName: method name to opnum. It is built from
// OpnumToName at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
