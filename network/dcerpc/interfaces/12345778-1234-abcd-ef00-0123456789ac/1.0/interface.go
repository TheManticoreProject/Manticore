// Package rpcinterface_123457781234abcdef000123456789ac_1_0 is the descriptor for the
// Security Account Manager (SAM) Remote Protocol (samr) RPC interface, abstract syntax
// 12345778-1234-abcd-ef00-0123456789ac version 1.0 ([MS-SAMR]).
//
// An RPC interface is identified by its UUID and version, never by the named pipe it is
// reached over: the "\samr" pipe carries this interface (note the UUID differs from
// lsarpc's only in the last nibble, ...ac vs ...ab); the directory is named after the
// interface UUID with the version in the nested 1.0/ directory.
//
// This package holds only the interface-level descriptor: the abstract syntax
// identifier, the transport endpoint (PipeName), the opnum constants and opnum<->name
// maps, and the NTSTATUS return codes. NDR types live in the structures subpackage and
// the method stubs in functions; both depend on this package, never the reverse.
//
// References:
//   - [MS-SAMR] Security Account Manager (SAM) Remote Protocol:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/4df07fab-1bbc-452f-8e92-7853a3c7e380
package rpcinterface_123457781234abcdef000123456789ac_1_0

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the IPC$-relative named pipe for the samr interface.
const PipeName = `\samr`

// Opnums for the on-the-wire methods ([MS-SAMR] section 3.1.5). Opnums 4, 42, 43, 59-61,
// 63, 68-72, 75, 76 are "not used on the wire" and are omitted.
const (
	OpnumSamrConnect                                 uint16 = 0
	OpnumSamrCloseHandle                             uint16 = 1
	OpnumSamrSetSecurityObject                       uint16 = 2
	OpnumSamrQuerySecurityObject                     uint16 = 3
	OpnumSamrLookupDomainInSamServer                 uint16 = 5
	OpnumSamrEnumerateDomainsInSamServer             uint16 = 6
	OpnumSamrOpenDomain                              uint16 = 7
	OpnumSamrQueryInformationDomain                  uint16 = 8
	OpnumSamrSetInformationDomain                    uint16 = 9
	OpnumSamrCreateGroupInDomain                     uint16 = 10
	OpnumSamrEnumerateGroupsInDomain                 uint16 = 11
	OpnumSamrCreateUserInDomain                      uint16 = 12
	OpnumSamrEnumerateUsersInDomain                  uint16 = 13
	OpnumSamrCreateAliasInDomain                     uint16 = 14
	OpnumSamrEnumerateAliasesInDomain                uint16 = 15
	OpnumSamrGetAliasMembership                      uint16 = 16
	OpnumSamrLookupNamesInDomain                     uint16 = 17
	OpnumSamrLookupIdsInDomain                       uint16 = 18
	OpnumSamrOpenGroup                               uint16 = 19
	OpnumSamrQueryInformationGroup                   uint16 = 20
	OpnumSamrSetInformationGroup                     uint16 = 21
	OpnumSamrAddMemberToGroup                        uint16 = 22
	OpnumSamrDeleteGroup                             uint16 = 23
	OpnumSamrRemoveMemberFromGroup                   uint16 = 24
	OpnumSamrGetMembersInGroup                       uint16 = 25
	OpnumSamrSetMemberAttributesOfGroup              uint16 = 26
	OpnumSamrOpenAlias                               uint16 = 27
	OpnumSamrQueryInformationAlias                   uint16 = 28
	OpnumSamrSetInformationAlias                     uint16 = 29
	OpnumSamrDeleteAlias                             uint16 = 30
	OpnumSamrAddMemberToAlias                        uint16 = 31
	OpnumSamrRemoveMemberFromAlias                   uint16 = 32
	OpnumSamrGetMembersInAlias                       uint16 = 33
	OpnumSamrOpenUser                                uint16 = 34
	OpnumSamrDeleteUser                              uint16 = 35
	OpnumSamrQueryInformationUser                    uint16 = 36
	OpnumSamrSetInformationUser                      uint16 = 37
	OpnumSamrChangePasswordUser                      uint16 = 38
	OpnumSamrGetGroupsForUser                        uint16 = 39
	OpnumSamrQueryDisplayInformation                 uint16 = 40
	OpnumSamrGetDisplayEnumerationIndex              uint16 = 41
	OpnumSamrGetUserDomainPasswordInformation        uint16 = 44
	OpnumSamrRemoveMemberFromForeignDomain           uint16 = 45
	OpnumSamrQueryInformationDomain2                 uint16 = 46
	OpnumSamrQueryInformationUser2                   uint16 = 47
	OpnumSamrQueryDisplayInformation2                uint16 = 48
	OpnumSamrGetDisplayEnumerationIndex2             uint16 = 49
	OpnumSamrCreateUser2InDomain                     uint16 = 50
	OpnumSamrQueryDisplayInformation3                uint16 = 51
	OpnumSamrAddMultipleMembersToAlias               uint16 = 52
	OpnumSamrRemoveMultipleMembersFromAlias          uint16 = 53
	OpnumSamrOemChangePasswordUser2                  uint16 = 54
	OpnumSamrUnicodeChangePasswordUser2              uint16 = 55
	OpnumSamrGetDomainPasswordInformation            uint16 = 56
	OpnumSamrConnect2                                uint16 = 57
	OpnumSamrSetInformationUser2                     uint16 = 58
	OpnumSamrConnect4                                uint16 = 62
	OpnumSamrConnect5                                uint16 = 64
	OpnumSamrRidToSid                                uint16 = 65
	OpnumSamrSetDSRMPassword                         uint16 = 66
	OpnumSamrValidatePassword                        uint16 = 67
	OpnumSamrUnicodeChangePasswordUser4              uint16 = 73
	OpnumSamrValidateComputerAccountReuseAttempt     uint16 = 74
	OpnumSamrAccountIsDelegatedManagedServiceAccount uint16 = 77
)

// Common NTSTATUS codes returned by samr methods ([MS-ERREF] 2.3.1). samr methods return
// an NTSTATUS (the IDL "long" return value).
const (
	StatusSuccess            uint32 = 0x00000000
	StatusMoreEntries        uint32 = 0x00000105
	StatusSomeNotMapped      uint32 = 0x00000107
	StatusNoMoreEntries      uint32 = 0x8000001A
	StatusInvalidHandle      uint32 = 0xC0000008
	StatusInvalidParameter   uint32 = 0xC000000D
	StatusAccessDenied       uint32 = 0xC0000022
	StatusObjectNameNotFound uint32 = 0xC0000034
	StatusNoSuchUser         uint32 = 0xC0000064
	StatusNoneMapped         uint32 = 0xC0000073
	StatusNoSuchDomain       uint32 = 0xC00000DF
	StatusNoSuchAlias        uint32 = 0xC0000151
	StatusNoSuchGroup        uint32 = 0xC0000066
	StatusUserExists         uint32 = 0xC0000063
	StatusGroupExists        uint32 = 0xC0000065
	StatusAliasExists        uint32 = 0xC0000154
	StatusWrongPassword      uint32 = 0xC000006A
	StatusNotSupported       uint32 = 0xC00000BB
)

// SyntaxID returns the samr abstract syntax identifier:
// 12345778-1234-abcd-ef00-0123456789ac, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x12345778, B: 0x1234, C: 0xabcd, D: 0xef00, E: 0x0123456789ac},
		MajorVersion: 1,
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
	case StatusSomeNotMapped:
		return "STATUS_SOME_NOT_MAPPED"
	case StatusNoMoreEntries:
		return "STATUS_NO_MORE_ENTRIES"
	case StatusInvalidHandle:
		return "STATUS_INVALID_HANDLE"
	case StatusInvalidParameter:
		return "STATUS_INVALID_PARAMETER"
	case StatusAccessDenied:
		return "STATUS_ACCESS_DENIED"
	case StatusObjectNameNotFound:
		return "STATUS_OBJECT_NAME_NOT_FOUND"
	case StatusNoSuchUser:
		return "STATUS_NO_SUCH_USER"
	case StatusNoneMapped:
		return "STATUS_NONE_MAPPED"
	case StatusNoSuchDomain:
		return "STATUS_NO_SUCH_DOMAIN"
	case StatusNoSuchAlias:
		return "STATUS_NO_SUCH_ALIAS"
	case StatusNoSuchGroup:
		return "STATUS_NO_SUCH_GROUP"
	case StatusUserExists:
		return "STATUS_USER_EXISTS"
	case StatusGroupExists:
		return "STATUS_GROUP_EXISTS"
	case StatusAliasExists:
		return "STATUS_ALIAS_EXISTS"
	case StatusWrongPassword:
		return "STATUS_WRONG_PASSWORD"
	case StatusNotSupported:
		return "STATUS_NOT_SUPPORTED"
	default:
		return fmt.Sprintf("0x%08x", status)
	}
}

// OpnumToName maps each on-the-wire opnum to its [MS-SAMR] method name; the single
// source of truth.
var OpnumToName = map[uint16]string{
	OpnumSamrConnect:                                 "SamrConnect",
	OpnumSamrCloseHandle:                             "SamrCloseHandle",
	OpnumSamrSetSecurityObject:                       "SamrSetSecurityObject",
	OpnumSamrQuerySecurityObject:                     "SamrQuerySecurityObject",
	OpnumSamrLookupDomainInSamServer:                 "SamrLookupDomainInSamServer",
	OpnumSamrEnumerateDomainsInSamServer:             "SamrEnumerateDomainsInSamServer",
	OpnumSamrOpenDomain:                              "SamrOpenDomain",
	OpnumSamrQueryInformationDomain:                  "SamrQueryInformationDomain",
	OpnumSamrSetInformationDomain:                    "SamrSetInformationDomain",
	OpnumSamrCreateGroupInDomain:                     "SamrCreateGroupInDomain",
	OpnumSamrEnumerateGroupsInDomain:                 "SamrEnumerateGroupsInDomain",
	OpnumSamrCreateUserInDomain:                      "SamrCreateUserInDomain",
	OpnumSamrEnumerateUsersInDomain:                  "SamrEnumerateUsersInDomain",
	OpnumSamrCreateAliasInDomain:                     "SamrCreateAliasInDomain",
	OpnumSamrEnumerateAliasesInDomain:                "SamrEnumerateAliasesInDomain",
	OpnumSamrGetAliasMembership:                      "SamrGetAliasMembership",
	OpnumSamrLookupNamesInDomain:                     "SamrLookupNamesInDomain",
	OpnumSamrLookupIdsInDomain:                       "SamrLookupIdsInDomain",
	OpnumSamrOpenGroup:                               "SamrOpenGroup",
	OpnumSamrQueryInformationGroup:                   "SamrQueryInformationGroup",
	OpnumSamrSetInformationGroup:                     "SamrSetInformationGroup",
	OpnumSamrAddMemberToGroup:                        "SamrAddMemberToGroup",
	OpnumSamrDeleteGroup:                             "SamrDeleteGroup",
	OpnumSamrRemoveMemberFromGroup:                   "SamrRemoveMemberFromGroup",
	OpnumSamrGetMembersInGroup:                       "SamrGetMembersInGroup",
	OpnumSamrSetMemberAttributesOfGroup:              "SamrSetMemberAttributesOfGroup",
	OpnumSamrOpenAlias:                               "SamrOpenAlias",
	OpnumSamrQueryInformationAlias:                   "SamrQueryInformationAlias",
	OpnumSamrSetInformationAlias:                     "SamrSetInformationAlias",
	OpnumSamrDeleteAlias:                             "SamrDeleteAlias",
	OpnumSamrAddMemberToAlias:                        "SamrAddMemberToAlias",
	OpnumSamrRemoveMemberFromAlias:                   "SamrRemoveMemberFromAlias",
	OpnumSamrGetMembersInAlias:                       "SamrGetMembersInAlias",
	OpnumSamrOpenUser:                                "SamrOpenUser",
	OpnumSamrDeleteUser:                              "SamrDeleteUser",
	OpnumSamrQueryInformationUser:                    "SamrQueryInformationUser",
	OpnumSamrSetInformationUser:                      "SamrSetInformationUser",
	OpnumSamrChangePasswordUser:                      "SamrChangePasswordUser",
	OpnumSamrGetGroupsForUser:                        "SamrGetGroupsForUser",
	OpnumSamrQueryDisplayInformation:                 "SamrQueryDisplayInformation",
	OpnumSamrGetDisplayEnumerationIndex:              "SamrGetDisplayEnumerationIndex",
	OpnumSamrGetUserDomainPasswordInformation:        "SamrGetUserDomainPasswordInformation",
	OpnumSamrRemoveMemberFromForeignDomain:           "SamrRemoveMemberFromForeignDomain",
	OpnumSamrQueryInformationDomain2:                 "SamrQueryInformationDomain2",
	OpnumSamrQueryInformationUser2:                   "SamrQueryInformationUser2",
	OpnumSamrQueryDisplayInformation2:                "SamrQueryDisplayInformation2",
	OpnumSamrGetDisplayEnumerationIndex2:             "SamrGetDisplayEnumerationIndex2",
	OpnumSamrCreateUser2InDomain:                     "SamrCreateUser2InDomain",
	OpnumSamrQueryDisplayInformation3:                "SamrQueryDisplayInformation3",
	OpnumSamrAddMultipleMembersToAlias:               "SamrAddMultipleMembersToAlias",
	OpnumSamrRemoveMultipleMembersFromAlias:          "SamrRemoveMultipleMembersFromAlias",
	OpnumSamrOemChangePasswordUser2:                  "SamrOemChangePasswordUser2",
	OpnumSamrUnicodeChangePasswordUser2:              "SamrUnicodeChangePasswordUser2",
	OpnumSamrGetDomainPasswordInformation:            "SamrGetDomainPasswordInformation",
	OpnumSamrConnect2:                                "SamrConnect2",
	OpnumSamrSetInformationUser2:                     "SamrSetInformationUser2",
	OpnumSamrConnect4:                                "SamrConnect4",
	OpnumSamrConnect5:                                "SamrConnect5",
	OpnumSamrRidToSid:                                "SamrRidToSid",
	OpnumSamrSetDSRMPassword:                         "SamrSetDSRMPassword",
	OpnumSamrValidatePassword:                        "SamrValidatePassword",
	OpnumSamrUnicodeChangePasswordUser4:              "SamrUnicodeChangePasswordUser4",
	OpnumSamrValidateComputerAccountReuseAttempt:     "SamrValidateComputerAccountReuseAttempt",
	OpnumSamrAccountIsDelegatedManagedServiceAccount: "SamrAccountIsDelegatedManagedServiceAccount",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
