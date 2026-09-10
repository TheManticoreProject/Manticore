package rpcinterface_ea0a3165483411d2a6f800c04fa346cc_4_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// TestSyntaxID verifies the abstract syntax identifier matches the IDL: UUID
// ea0a3165-4834-11d2-a6f8-00c04fa346cc, version 4.0 ([MS-FAX]).
func TestSyntaxID(t *testing.T) {
	sid := SyntaxID()
	want := guid.GUID{A: 0xea0a3165, B: 0x4834, C: 0x11d2, D: 0xa6f8, E: 0x00c04fa346cc}
	if sid.UUID != want {
		t.Errorf("UUID = %+v, want %+v", sid.UUID, want)
	}
	if sid.MajorVersion != 4 || sid.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 4.0", sid.MajorVersion, sid.MinorVersion)
	}
}

// TestOpnumNameRoundTrip confirms OpnumToName and its derived reverse map agree, and that
// the wire opnums are the dense range 0..103 the fax interface defines.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 104 {
		t.Fatalf("OpnumToName has %d entries, want 104", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Fatalf("NameToOpnum has %d entries, want %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d (ok=%v), want %d", name, got, ok, op)
		}
	}
	for op := uint16(0); op < 104; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that the common Win32 codes [MS-FAX]
// documents for the fax interface resolve through the shared [MS-ERREF] 2.2 table
// under their specification names, that a code the interface never enumerated now
// renders by name rather than as undecoded hex, and that a value the specification
// does not define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000008: "ERROR_NOT_ENOUGH_MEMORY",
		0x0000000D: "ERROR_INVALID_DATA",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000006F: "ERROR_BUFFER_OVERFLOW",
		0x0000007A: "ERROR_INSUFFICIENT_BUFFER",
		0x0000007C: "ERROR_INVALID_LEVEL",
		0x000000B7: "ERROR_ALREADY_EXISTS",
		0x00000103: "ERROR_NO_MORE_ITEMS",
	}
	for code, name := range documented {
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

	// ERROR_GEN_FAILURE was outside the subset this interface used to declare, so it
	// rendered as hex; it resolves by name now, through StatusString as well.
	if got := win32.WIN32_ERROR(0x0000001F).String(); got != "ERROR_GEN_FAILURE" {
		t.Errorf("win32.WIN32_ERROR(0x0000001f).String() = %q, want ERROR_GEN_FAILURE", got)
	}
	if got := StatusString(0x0000001F); got != "ERROR_GEN_FAILURE" {
		t.Errorf("StatusString(0x0000001f) = %q, want ERROR_GEN_FAILURE", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0x12345678).String(); got != "0x12345678" {
		t.Errorf("win32.WIN32_ERROR(0x12345678).String() = %q, want hex", got)
	}
	if got := StatusString(0x12345678); got != "0x12345678" {
		t.Errorf("StatusString(0x12345678) = %q, want hex", got)
	}
}

// TestStatusStringDecodesFaxSpecificErrors pins the one reason StatusString still
// exists: [MS-FAX] section 2.2.52 assigns the fax-specific errors values that
// [MS-ERREF] 2.2 already assigns to the unrelated Terminal Services ERROR_CTX_*
// errors, so these codes cannot be read out of the shared table.
func TestStatusStringDecodesFaxSpecificErrors(t *testing.T) {
	faxSpecific := map[uint32]string{
		FaxErrSrvOutOfMemory:        "FAX_ERR_SRV_OUTOFMEMORY",
		FaxErrGroupNotFound:         "FAX_ERR_GROUP_NOT_FOUND",
		FaxErrBadGroupConfiguration: "FAX_ERR_BAD_GROUP_CONFIGURATION",
		FaxErrGroupInUse:            "FAX_ERR_GROUP_IN_USE",
		FaxErrRuleNotFound:          "FAX_ERR_RULE_NOT_FOUND",
		FaxErrNotNTFS:               "FAX_ERR_NOT_NTFS",
		FaxErrDirectoryInUse:        "FAX_ERR_DIRECTORY_IN_USE",
		FaxErrFileAccessDenied:      "FAX_ERR_FILE_ACCESS_DENIED",
		FaxErrMessageNotFound:       "FAX_ERR_MESSAGE_NOT_FOUND",
		FaxErrDeviceNumLimit:        "FAX_ERR_DEVICE_NUM_LIMIT_EXCEEDED",
		FaxErrNotSupportedOnThisSKU: "FAX_ERR_NOT_SUPPORTED_ON_THIS_SKU",
		FaxErrVersionMismatch:       "FAX_ERR_VERSION_MISMATCH",
		FaxErrRecipientsLimit:       "FAX_ERR_RECIPIENTS_LIMIT",
	}
	if len(faxSpecific) != 13 {
		t.Fatalf("fax-specific table has %d entries, want 13", len(faxSpecific))
	}
	for code, name := range faxSpecific {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
		if code < 0x00001B59 || code > 0x00001B65 {
			t.Errorf("%s = 0x%08x, outside the fax error range 0x1B59..0x1B65", name, code)
		}
		if _, defined := win32.FromName(name); defined {
			t.Errorf("win32.FromName(%q) resolves; the code belongs in the shared table instead", name)
		}
	}

	// The collision that makes the local table necessary: the shared table names
	// 0x1B59 and 0x1B5E for Terminal Services, not for fax.
	if got := win32.WIN32_ERROR(FaxErrSrvOutOfMemory).String(); got != "ERROR_CTX_WINSTATION_NAME_INVALID" {
		t.Errorf("win32.WIN32_ERROR(0x00001b59).String() = %q, want ERROR_CTX_WINSTATION_NAME_INVALID", got)
	}
	if got := win32.WIN32_ERROR(FaxErrNotNTFS).String(); got != "ERROR_CTX_SERVICE_NAME_COLLISION" {
		t.Errorf("win32.WIN32_ERROR(0x00001b5e).String() = %q, want ERROR_CTX_SERVICE_NAME_COLLISION", got)
	}
}
