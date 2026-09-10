package rpcinterface_86d3594983c94044b424db363231fd0c_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of
// ITaskSchedulerService.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "86d35949-83c9-4044-b424-db363231fd0c" {
		t.Fatalf("UUID = %s, want 86d35949-83c9-4044-b424-db363231fd0c", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName verifies there is no well-known named pipe: ITaskSchedulerService is an
// ncacn_ip_tcp dynamic-endpoint interface ([MS-TSCH] 2.1).
func TestPipeName(t *testing.T) {
	if PipeName != "" {
		t.Fatalf("PipeName = %q, want empty (ncacn_ip_tcp dynamic endpoint)", PipeName)
	}
}

// TestOpnums verifies the 20 implemented opnums and the name mapping round trip.
func TestOpnums(t *testing.T) {
	if OpnumSchRpcHighestVersion != 0 || OpnumSchRpcEnableTask != 19 {
		t.Fatalf("opnum bounds = %d..%d, want 0..19", OpnumSchRpcHighestVersion, OpnumSchRpcEnableTask)
	}
	if len(OpnumToName) != 20 {
		t.Fatalf("OpnumToName has %d entries, want 20", len(OpnumToName))
	}
	if OpnumToName[12] != "SchRpcRun" || NameToOpnum["SchRpcRegisterTask"] != 1 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusCodesResolveThroughHRESULT pins that every HRESULT this interface used
// to declare resolves through the shared [MS-ERREF] 2.1.1 table under the name the
// specification gives that value, in both directions, and that StatusString reports
// the same name.
//
// The two entries that carry the regression test for the mispairing this table
// replaces are 0x00041306 and 0x00041302: the interface declared the first as
// SCHED_S_TASK_DISABLED and never declared the second at all, where the
// specification names 0x00041306 SCHED_S_TASK_TERMINATED and 0x00041302
// SCHED_S_TASK_DISABLED.
func TestStatusCodesResolveThroughHRESULT(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "S_OK",
		0x00000001: "S_FALSE",
		0x00041300: "SCHED_S_TASK_READY",
		0x00041301: "SCHED_S_TASK_RUNNING",
		0x00041302: "SCHED_S_TASK_DISABLED",
		0x00041303: "SCHED_S_TASK_HAS_NOT_RUN",
		0x00041304: "SCHED_S_TASK_NO_MORE_RUNS",
		0x00041305: "SCHED_S_TASK_NOT_SCHEDULED",
		0x00041306: "SCHED_S_TASK_TERMINATED",
		0x80070005: "E_ACCESSDENIED",
		0x80070057: "E_INVALIDARG",
		0x80041309: "SCHED_E_TRIGGER_NOT_FOUND",
		0x8004130A: "SCHED_E_TASK_NOT_READY",
		0x8004130B: "SCHED_E_TASK_NOT_RUNNING",
		0x8004130C: "SCHED_E_SERVICE_NOT_INSTALLED",
		0x8004130D: "SCHED_E_CANNOT_OPEN_TASK",
		0x8004130E: "SCHED_E_INVALID_TASK",
		0x8004130F: "SCHED_E_ACCOUNT_INFORMATION_NOT_SET",
		0x80041310: "SCHED_E_ACCOUNT_NAME_NOT_FOUND",
		0x80041315: "SCHED_E_SERVICE_NOT_RUNNING",
		0x80041316: "SCHED_E_UNEXPECTEDNODE",
		0x80041317: "SCHED_E_NAMESPACE",
		0x80041318: "SCHED_E_INVALIDVALUE",
		0x80041319: "SCHED_E_MISSINGNODE",
		0x8004131A: "SCHED_E_MALFORMEDXML",
		0x8004131D: "SCHED_E_TOO_MANY_NODES",
		0x8004131E: "SCHED_E_PAST_END_BOUNDARY",
		0x8004131F: "SCHED_E_ALREADY_RUNNING",
		0x80041320: "SCHED_E_USER_NOT_LOGGED_ON",
		0x80041321: "SCHED_E_INVALID_TASK_HASH",
		0x80041322: "SCHED_E_SERVICE_NOT_AVAILABLE",
		0x80041323: "SCHED_E_SERVICE_TOO_BUSY",
		0x80041324: "SCHED_E_TASK_ATTEMPTED",
	}
	for code, name := range documented {
		if got := hresult.HRESULT(code).String(); got != name {
			t.Errorf("hresult.HRESULT(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := hresult.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("hresult.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
	}

	// The mispairing itself, pinned in both directions: the value the interface
	// rendered as SCHED_S_TASK_DISABLED is SCHED_S_TASK_TERMINATED, and the value
	// that really is SCHED_S_TASK_DISABLED was absent from the subset and rendered
	// as undecoded hex.
	if got := StatusString(0x00041306); got != "SCHED_S_TASK_TERMINATED" {
		t.Errorf("StatusString(0x00041306) = %q, want SCHED_S_TASK_TERMINATED", got)
	}
	if got := StatusString(0x00041302); got != "SCHED_S_TASK_DISABLED" {
		t.Errorf("StatusString(0x00041302) = %q, want SCHED_S_TASK_DISABLED", got)
	}

	// Values the subset never enumerated now resolve by name rather than as hex.
	for code, name := range map[uint32]string{
		0x00041307: "SCHED_S_TASK_NO_VALID_TRIGGERS",
		0x00041308: "SCHED_S_EVENT_TRIGGER",
		0x0004131B: "SCHED_S_SOME_TRIGGERS_FAILED",
		0x0004131C: "SCHED_S_BATCH_LOGON_PROBLEM",
		0x80041311: "SCHED_E_ACCOUNT_DBASE_CORRUPT",
		0x80041312: "SCHED_E_NO_SECURITY_SERVICES",
		0x80041313: "SCHED_E_UNKNOWN_OBJECT_VERSION",
		0x80041314: "SCHED_E_UNSUPPORTED_ACCOUNT_OPTION",
	} {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
	}

	// The FACILITY_WIN32 codes are derived from the Win32 table rather than named by
	// [MS-ERREF] 2.1.1, so they render as the HRESULT_FROM_WIN32 of the code they
	// wrap and unwrap back to it.
	for code, name := range map[uint32]string{
		0x80070002: "HRESULT_FROM_WIN32(ERROR_FILE_NOT_FOUND)",
		0x800700B7: "HRESULT_FROM_WIN32(ERROR_ALREADY_EXISTS)",
	} {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
	}
	if got := hresult.FromWin32(win32.ERROR_FILE_NOT_FOUND); uint32(got) != 0x80070002 {
		t.Errorf("hresult.FromWin32(ERROR_FILE_NOT_FOUND) = 0x%08x, want 0x80070002", uint32(got))
	}
	if got, wraps := hresult.HRESULT(0x800700B7).ToWin32(); !wraps || got != win32.ERROR_ALREADY_EXISTS {
		t.Errorf("hresult.HRESULT(0x800700b7).ToWin32() = %d, %v; want ERROR_ALREADY_EXISTS, true", got, wraps)
	}

	// A value [MS-ERREF] 2.1.1 does not define still renders as hex.
	if got := StatusString(0x0EADBEEF); got != "0x0eadbeef" {
		t.Errorf("StatusString(0x0eadbeef) = %q, want hex", got)
	}
}

// TestStatusStringDecodesWinErrorOnlyCodes pins the one reason StatusString still
// exists: WinError.h defines these three Task Scheduler errors, the [MS-ERREF] 2.1.1
// SCHED_* block ends at SCHED_E_TASK_ATTEMPTED (0x80041324) and names none of them,
// so the shared table cannot decode them.
func TestStatusStringDecodesWinErrorOnlyCodes(t *testing.T) {
	local := map[uint32]string{
		SchedETaskDisabled:    "SCHED_E_TASK_DISABLED",
		SchedETaskNotV1Compat: "SCHED_E_TASK_NOT_V1_COMPAT",
		SchedEStartOnDemand:   "SCHED_E_START_ON_DEMAND",
	}
	if len(local) != 3 {
		t.Fatalf("local table has %d entries, want 3", len(local))
	}
	for code, name := range local {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
		if _, defined := hresult.FromName(name); defined {
			t.Errorf("hresult.FromName(%q) resolves; the code belongs in the shared table instead", name)
		}
		if got := hresult.HRESULT(code).String(); got == name {
			t.Errorf("hresult.HRESULT(0x%08x).String() = %q; the local decode is redundant", code, got)
		}
	}
	if SchedETaskDisabled != 0x80041326 || SchedETaskNotV1Compat != 0x80041327 || SchedEStartOnDemand != 0x80041328 {
		t.Fatalf("WinError.h values drifted: 0x%08x, 0x%08x, 0x%08x", SchedETaskDisabled, SchedETaskNotV1Compat, SchedEStartOnDemand)
	}
}

// TestIsSuccessIsARange verifies that the shared predicate the stubs now use reads
// bit 31 rather than comparing against S_OK, so S_FALSE and the SCHED_S_*
// informational codes are successes and Error reports nil for them.
func TestIsSuccessIsARange(t *testing.T) {
	for _, ok := range []uint32{0x00000000, 0x00000001, 0x00041301, 0x00041306, 0x0004131B} {
		if !hresult.HRESULT(ok).IsSuccess() {
			t.Errorf("hresult.HRESULT(0x%08x).IsSuccess() = false, want true", ok)
		}
		if err := hresult.HRESULT(ok).Error(); err != nil {
			t.Errorf("hresult.HRESULT(0x%08x).Error() = %v, want nil", ok, err)
		}
	}
	for _, bad := range []uint32{0x80070005, 0x80041309, SchedEStartOnDemand} {
		if hresult.HRESULT(bad).IsSuccess() {
			t.Errorf("hresult.HRESULT(0x%08x).IsSuccess() = true, want false", bad)
		}
		if hresult.HRESULT(bad).Error() == nil {
			t.Errorf("hresult.HRESULT(0x%08x).Error() = nil, want an error", bad)
		}
	}
}
