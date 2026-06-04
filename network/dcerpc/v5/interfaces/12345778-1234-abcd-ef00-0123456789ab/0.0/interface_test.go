package rpcinterface_123457781234abcdef000123456789ab_0_0

import "testing"

func TestStatusString(t *testing.T) {
	if got := StatusString(StatusAccessDenied); got != "STATUS_ACCESS_DENIED" {
		t.Errorf("StatusString(0xC0000022) = %q", got)
	}
	if got := StatusString(0x12345678); got != "0x12345678" {
		t.Errorf("StatusString(unknown) = %q", got)
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	// 12345778-1234-abcd-ef00-0123456789ab, version 0.0.
	if id.UUID.A != 0x12345778 || id.UUID.B != 0x1234 || id.UUID.C != 0xabcd ||
		id.UUID.D != 0xef00 || id.UUID.E != 0x0123456789ab {
		t.Errorf("SyntaxID UUID = %+v, want 12345778-1234-abcd-ef00-0123456789ab", id.UUID)
	}
	if id.MajorVersion != 0 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 0.0", id.MajorVersion, id.MinorVersion)
	}
}
