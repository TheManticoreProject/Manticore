package rpcinterface_5b821720f63b11d0aad200c04fc324db_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// TestSyntaxID verifies the abstract syntax identifier matches the IDL: UUID
// 5b821720-f63b-11d0-aad2-00c04fc324db, version 1.0 ([MS-DHCPM]).
func TestSyntaxID(t *testing.T) {
	sid := SyntaxID()
	want := guid.GUID{A: 0x5b821720, B: 0xf63b, C: 0x11d0, D: 0xaad2, E: 0x00c04fc324db}
	if sid.UUID != want {
		t.Errorf("UUID = %+v, want %+v", sid.UUID, want)
	}
	if sid.MajorVersion != 1 || sid.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", sid.MajorVersion, sid.MinorVersion)
	}
}

// TestOpnumNameRoundTrip confirms OpnumToName and its derived reverse map agree, and that
// the wire opnums are the dense range 0..132 the dhcpsrv2 interface defines.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 133 {
		t.Fatalf("OpnumToName has %d entries, want 133", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Fatalf("NameToOpnum has %d entries, want %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d (ok=%v), want %d", name, got, ok, op)
		}
	}
	for op := uint16(0); op < 133; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that the Win32 codes this interface used to
// declare privately resolve through the shared [MS-ERREF] 2.2 table under their
// specification names, that a code the private subset never enumerated now renders by
// name rather than as undecoded hex, and that a value the specification does not define
// still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	migrated := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x0000000D: "ERROR_INVALID_DATA",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x000000B7: "ERROR_ALREADY_EXISTS",
		0x000000EA: "ERROR_MORE_DATA",
		0x00000103: "ERROR_NO_MORE_ITEMS",
	}
	if len(migrated) != 8 {
		t.Fatalf("migrated table has %d entries, want 8", len(migrated))
	}
	for code, name := range migrated {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
	}

	// The two pagination sentinels the Enum* guards tolerate alongside success are
	// the shared table's own codes, so the guards read them from there directly.
	if uint32(win32.ERROR_MORE_DATA) != 0x000000EA {
		t.Errorf("win32.ERROR_MORE_DATA = 0x%08x, want 0x000000ea", uint32(win32.ERROR_MORE_DATA))
	}
	if uint32(win32.ERROR_NO_MORE_ITEMS) != 0x00000103 {
		t.Errorf("win32.ERROR_NO_MORE_ITEMS = 0x%08x, want 0x00000103", uint32(win32.ERROR_NO_MORE_ITEMS))
	}

	// ERROR_INVALID_HANDLE was outside the subset this interface used to declare, so it
	// rendered as hex; it resolves by name now, through StatusString as well.
	if got := win32.WIN32_ERROR(0x00000006).String(); got != "ERROR_INVALID_HANDLE" {
		t.Errorf("win32.WIN32_ERROR(0x00000006).String() = %q, want ERROR_INVALID_HANDLE", got)
	}
	if got := StatusString(0x00000006); got != "ERROR_INVALID_HANDLE" {
		t.Errorf("StatusString(0x00000006) = %q, want ERROR_INVALID_HANDLE", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0x12345678).String(); got != "0x12345678" {
		t.Errorf("win32.WIN32_ERROR(0x12345678).String() = %q, want hex", got)
	}
	if got := StatusString(0x12345678); got != "0x12345678" {
		t.Errorf("StatusString(0x12345678) = %q, want hex", got)
	}
}

// TestStatusStringDecodesDhcpSpecificErrors pins the one reason StatusString still
// exists: [MS-DHCPM] 2.2.1.2 assigns the DHCP-specific errors values in the
// 0x00004E20..0x00004E4D range, and [MS-ERREF] 2.2 carries no row for any value in
// it, so these codes cannot be read out of the shared table. The win32.Lookup
// assertion below fails loudly should a future [MS-ERREF] revision assign the range.
func TestStatusStringDecodesDhcpSpecificErrors(t *testing.T) {
	dhcpSpecific := map[uint32]string{
		ErrorDhcpRegistryInitFailed:    "ERROR_DHCP_REGISTRY_INIT_FAILED",
		ErrorDhcpDatabaseInitFailed:    "ERROR_DHCP_DATABASE_INIT_FAILED",
		ErrorDhcpRpcInitFailed:         "ERROR_DHCP_RPC_INIT_FAILED",
		ErrorDhcpNetworkInitFailed:     "ERROR_DHCP_NETWORK_INIT_FAILED",
		ErrorDhcpSubnetExists:          "ERROR_DHCP_SUBNET_EXISTS",
		ErrorDhcpSubnetNotPresent:      "ERROR_DHCP_SUBNET_NOT_PRESENT",
		ErrorDhcpPrimaryNotFound:       "ERROR_DHCP_PRIMARY_NOT_FOUND",
		ErrorDhcpElementCantRemove:     "ERROR_DHCP_ELEMENT_CANT_REMOVE",
		ErrorDhcpOptionExists:          "ERROR_DHCP_OPTION_EXISTS",
		ErrorDhcpOptionNotPresent:      "ERROR_DHCP_OPTION_NOT_PRESENT",
		ErrorDhcpAddressNotAvailable:   "ERROR_DHCP_ADDRESS_NOT_AVAILABLE",
		ErrorDhcpRangeFull:             "ERROR_DHCP_RANGE_FULL",
		ErrorDhcpJetError:              "ERROR_DHCP_JET_ERROR",
		ErrorDhcpClientExists:          "ERROR_DHCP_CLIENT_EXISTS",
		ErrorDhcpInvalidDhcpMessage:    "ERROR_DHCP_INVALID_DHCP_MESSAGE",
		ErrorDhcpInvalidDhcpClient:     "ERROR_DHCP_INVALID_DHCP_CLIENT",
		ErrorDhcpServicePaused:         "ERROR_DHCP_SERVICE_PAUSED",
		ErrorDhcpNotReservedClient:     "ERROR_DHCP_NOT_RESERVED_CLIENT",
		ErrorDhcpReservedClient:        "ERROR_DHCP_RESERVED_CLIENT",
		ErrorDhcpRangeTooSmall:         "ERROR_DHCP_RANGE_TOO_SMALL",
		ErrorDhcpIPRangeExists:         "ERROR_DHCP_IPRANGE_EXISTS",
		ErrorDhcpReservedIPExists:      "ERROR_DHCP_RESERVEDIP_EXISTS",
		ErrorDhcpInvalidRange:          "ERROR_DHCP_INVALID_RANGE",
		ErrorDhcpRangeExtended:         "ERROR_DHCP_RANGE_EXTENDED",
		ErrorDhcpSuperScopeNameTooLong: "ERROR_DHCP_SUPER_SCOPE_NAME_TOO_LONG",
		ErrorDhcpClassNotFound:         "ERROR_DHCP_CLASS_NOT_FOUND",
		ErrorDhcpClassAlreadyExists:    "ERROR_DHCP_CLASS_ALREADY_EXISTS",
	}
	if len(dhcpSpecific) != 27 {
		t.Fatalf("DHCP-specific table has %d entries, want 27", len(dhcpSpecific))
	}
	for code, name := range dhcpSpecific {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
		if code < 0x00004E20 || code > 0x00004E4D {
			t.Errorf("%s = 0x%08x, outside the DHCP error range 0x00004E20..0x00004E4D", name, code)
		}
		if entry, defined := win32.Lookup(win32.WIN32_ERROR(code)); defined {
			t.Errorf("win32.Lookup(0x%08x) resolves to %q; [MS-ERREF] 2.2 now assigns the DHCP range, so %s belongs in the shared table instead", code, entry.Name, name)
		}
		if _, defined := win32.FromName(name); defined {
			t.Errorf("win32.FromName(%q) resolves; the code belongs in the shared table instead", name)
		}
	}

	// The two endpoints of the range, to pin what the shared table does with a value
	// it has no row for: bare hex, which is why the local table is still needed.
	if got := win32.WIN32_ERROR(ErrorDhcpRegistryInitFailed).String(); got != "0x00004e20" {
		t.Errorf("win32.WIN32_ERROR(0x00004e20).String() = %q, want hex", got)
	}
	if got := win32.WIN32_ERROR(ErrorDhcpClassAlreadyExists).String(); got != "0x00004e4d" {
		t.Errorf("win32.WIN32_ERROR(0x00004e4d).String() = %q, want hex", got)
	}
}
