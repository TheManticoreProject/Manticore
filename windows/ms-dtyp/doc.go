// Package msdtyp provides the [MS-DTYP] Windows Data Types as a single canonical set of
// Go types shared across the codebase: the DCE/RPC interfaces and protocols (which
// marshal them with the reflection walker in network/dcerpc/ndr via the struct tags on
// these types), and the fixed-layout consumers (SMB, Active Directory replication).
//
// The package is intentionally dependency-light: it imports only windows/guid and the
// standard library, never network/dcerpc/ndr. The `ndr:"..."` struct tags are inert
// string metadata that the external NDR walker interprets; they do not couple this
// package to the codec, so non-RPC consumers can reuse the same definitions without
// pulling in the RPC stack.
//
// It contains the scalar type aliases ([MS-DTYP] 2.2, e.g. DWORD, WORD, WCHAR, ULONG),
// the common structures ([MS-DTYP] 2.3, e.g. GUID, RPC_SID, RPC_UNICODE_STRING, LUID,
// FILETIME, SYSTEMTIME, LARGE_INTEGER), and their conversion/formatting helpers.
//
// References:
//   - [MS-DTYP] Windows Data Types:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/cca27429-5689-4a16-b2b4-9325d93e4ba2
//   - [C706] DCE 1.1: RPC, Chapter 14 "Transfer Syntax NDR":
//     https://pubs.opengroup.org/onlinepubs/9629399/chap14.htm
package msdtyp
