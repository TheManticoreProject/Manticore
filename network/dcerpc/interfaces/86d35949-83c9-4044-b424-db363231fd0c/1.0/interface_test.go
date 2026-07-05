package rpcinterface_86d3594983c94044b424db363231fd0c_1_0

import "testing"

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

// TestStatusString verifies mnemonic rendering and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "S_OK" {
		t.Fatalf("StatusString(StatusSuccess) = %s, want S_OK", got)
	}
	if got := StatusString(StatusFalse); got != "S_FALSE" {
		t.Fatalf("StatusString(StatusFalse) = %s, want S_FALSE", got)
	}
	if got := StatusString(SchedEAccountNameNotFound); got != "SCHED_E_ACCOUNT_NAME_NOT_FOUND" {
		t.Fatalf("StatusString(SchedEAccountNameNotFound) = %s, want SCHED_E_ACCOUNT_NAME_NOT_FOUND", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}

// TestIsSuccess verifies the HRESULT success predicate (high bit clear), including
// S_FALSE and SCHED_S_* informational codes.
func TestIsSuccess(t *testing.T) {
	for _, ok := range []uint32{StatusSuccess, StatusFalse, SchedSTaskRunning} {
		if !IsSuccess(ok) {
			t.Fatalf("IsSuccess(0x%08x) = false, want true", ok)
		}
	}
	for _, bad := range []uint32{ErrorAccessDenied, SchedETriggerNotFound} {
		if IsSuccess(bad) {
			t.Fatalf("IsSuccess(0x%08x) = true, want false", bad)
		}
	}
}
