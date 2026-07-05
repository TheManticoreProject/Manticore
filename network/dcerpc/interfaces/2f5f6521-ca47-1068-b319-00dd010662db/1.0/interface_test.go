package rpcinterface_2f5f6521ca471068b31900dd010662db_1_0

import "testing"

// TestSyntaxID checks the abstract syntax matches [MS-TRP] Appendix A.1 (remotesp):
// 2f5f6521-ca47-1068-b319-00dd010662db, version 1.0.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "2f5f6521-ca47-1068-b319-00dd010662db" {
		t.Errorf("UUID = %s, want 2f5f6521-ca47-1068-b319-00dd010662db", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are consistent inverses and
// cover the three on-the-wire opnums.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 3 {
		t.Fatalf("OpnumToName has %d entries, want 3", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	if OpnumRemoteSPAttach != 0 || OpnumRemoteSPEventProc != 1 || OpnumRemoteSPDetach != 2 {
		t.Errorf("opnums = %d/%d/%d, want 0/1/2", OpnumRemoteSPAttach, OpnumRemoteSPEventProc, OpnumRemoteSPDetach)
	}
}

// TestStatusString covers the known code and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "STATUS_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want STATUS_SUCCESS", got)
	}
	if got := StatusString(0xdeadbeef); got != "0xdeadbeef" {
		t.Errorf("StatusString(0xdeadbeef) = %q, want hex fallback", got)
	}
}

// TestPipeName documents that remotesp has no fixed named pipe: [MS-TRP] 2.1 uses a
// client-specified endpoint for the reverse callback connection, so PipeName is empty.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (client-specified reverse endpoint)", PipeName)
	}
}
