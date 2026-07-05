package rpcinterface_1ff706820a5130e8076d740be8cee98b_1_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of atsvc.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "1ff70682-0a51-30e8-076d-740be8cee98b" {
		t.Fatalf("UUID = %s, want 1ff70682-0a51-30e8-076d-740be8cee98b", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName verifies the well-known ncacn_np endpoint shared with SASec.
func TestPipeName(t *testing.T) {
	if PipeName != `\atsvc` {
		t.Fatalf("PipeName = %q, want \\atsvc", PipeName)
	}
}

// TestOpnums verifies the implemented opnums and the name mapping round trip.
func TestOpnums(t *testing.T) {
	if OpnumNetrJobAdd != 0 || OpnumNetrJobDel != 1 || OpnumNetrJobEnum != 2 || OpnumNetrJobGetInfo != 3 {
		t.Fatalf("opnums = %d/%d/%d/%d, want 0/1/2/3", OpnumNetrJobAdd, OpnumNetrJobDel, OpnumNetrJobEnum, OpnumNetrJobGetInfo)
	}
	if OpnumToName[0] != "NetrJobAdd" || NameToOpnum["NetrJobGetInfo"] != 3 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusString verifies mnemonic rendering and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Fatalf("StatusString(StatusSuccess) = %s, want ERROR_SUCCESS", got)
	}
	if got := StatusString(ErrorMoreData); got != "ERROR_MORE_DATA" {
		t.Fatalf("StatusString(ErrorMoreData) = %s, want ERROR_MORE_DATA", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}
