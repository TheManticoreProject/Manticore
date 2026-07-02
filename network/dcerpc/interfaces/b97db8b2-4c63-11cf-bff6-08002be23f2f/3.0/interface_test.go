package rpcinterface_b97db8b24c6311cfbff608002be23f2f_3_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the clusapi interface
// (b97db8b2-4c63-11cf-bff6-08002be23f2f v3.0, [MS-CMRP]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "b97db8b2-4c63-11cf-bff6-08002be23f2f" {
		t.Errorf("UUID = %s, want b97db8b2-4c63-11cf-bff6-08002be23f2f", got)
	}
	if s.MajorVersion != 3 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 3.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// that a couple of anchor opnums resolve to the expected method names.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) differ in size",
			len(OpnumToName), len(NameToOpnum))
	}
	if OpnumToName[OpnumApiOpenCluster] != "ApiOpenCluster" {
		t.Errorf("opnum 0 = %q, want ApiOpenCluster", OpnumToName[OpnumApiOpenCluster])
	}
	if OpnumToName[OpnumApiCloseCluster] != "ApiCloseCluster" {
		t.Errorf("opnum 1 = %q, want ApiCloseCluster", OpnumToName[OpnumApiCloseCluster])
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestOpnumGaps checks that the "not used on the wire" opnums (e.g. 80, 264) are not
// present in the opnum map, while their neighbors are.
func TestOpnumGaps(t *testing.T) {
	for _, gap := range []uint16{80, 264} {
		if name, ok := OpnumToName[gap]; ok {
			t.Errorf("opnum %d should be absent (not used on the wire), got %q", gap, name)
		}
	}
	if _, ok := OpnumToName[79]; !ok {
		t.Errorf("opnum 79 (ApiNodeControl) should be present")
	}
}

// TestStatusString checks known Win32/cluster mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(StatusAccessDenied); got != "ERROR_ACCESS_DENIED" {
		t.Errorf("StatusString(5) = %q, want ERROR_ACCESS_DENIED", got)
	}
	if got := StatusString(StatusGroupNotFound); got != "ERROR_GROUP_NOT_FOUND" {
		t.Errorf("StatusString(0x1395) = %q, want ERROR_GROUP_NOT_FOUND", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}

// TestPipeName documents that this interface has no named pipe: MS-CMRP v3.0 uses
// ncacn_ip_tcp with a dynamic endpoint, so PipeName is intentionally empty.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (ncacn_ip_tcp dynamic endpoint)", PipeName)
	}
}
