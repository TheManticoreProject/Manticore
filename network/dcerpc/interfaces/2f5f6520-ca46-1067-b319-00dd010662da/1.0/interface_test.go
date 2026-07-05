package rpcinterface_2f5f6520ca461067b31900dd010662da_1_0

import "testing"

// TestSyntaxID checks the abstract syntax matches [MS-TRP] Appendix A.2 (tapsrv):
// 2f5f6520-ca46-1067-b319-00dd010662da, version 1.0.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "2f5f6520-ca46-1067-b319-00dd010662da" {
		t.Errorf("UUID = %s, want 2f5f6520-ca46-1067-b319-00dd010662da", got)
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
	if OpnumClientAttach != 0 || OpnumClientRequest != 1 || OpnumClientDetach != 2 {
		t.Errorf("opnums = %d/%d/%d, want 0/1/2", OpnumClientAttach, OpnumClientRequest, OpnumClientDetach)
	}
}

// TestStatusString covers the known code and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "STATUS_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want STATUS_SUCCESS", got)
	}
	if got := StatusString(0x80000005); got != "0x80000005" {
		t.Errorf("StatusString(0x80000005) = %q, want hex fallback", got)
	}
}

// TestPipeName pins the well-known tapsrv endpoint ([MS-TRP] 2.1).
func TestPipeName(t *testing.T) {
	if PipeName != `\tapsrv` {
		t.Errorf("PipeName = %q, want \\tapsrv", PipeName)
	}
}
