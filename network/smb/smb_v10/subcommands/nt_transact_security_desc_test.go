package subcommands

import (
	"bytes"
	"testing"
)

func TestNtTransactSecurityDescParametersGolden(t *testing.T) {
	// Query/set the DACL+OWNER of FID 0x4002.
	p := NtTransactSecurityDescParameters{
		FID:                 0x4002,
		SecurityInformation: OWNER_SECURITY_INFORMATION | DACL_SECURITY_INFORMATION, // 0x05
	}
	got, err := p.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x02, 0x40, // FID
		0x00, 0x00, // Reserved
		0x05, 0x00, 0x00, 0x00, // SecurityInformation = OWNER|DACL
	}
	if !bytes.Equal(got, want) {
		t.Errorf("security-desc params:\n got % x\nwant % x", got, want)
	}
	var out NtTransactSecurityDescParameters
	n, err := out.Unmarshal(got)
	if err != nil || n != ntTransactSecurityDescParametersSize {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
	if out != p {
		t.Errorf("round trip: got %+v want %+v", out, p)
	}
}

func TestNtTransactQuerySecurityDescResponseParametersRoundTrip(t *testing.T) {
	p := NtTransactQuerySecurityDescResponseParameters{LengthNeeded: 0x94}
	raw, err := p.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{0x94, 0x00, 0x00, 0x00}
	if !bytes.Equal(raw, want) {
		t.Errorf("LengthNeeded:\n got % x\nwant % x", raw, want)
	}
	var out NtTransactQuerySecurityDescResponseParameters
	if _, err := out.Unmarshal(raw); err != nil || out.LengthNeeded != 0x94 {
		t.Fatalf("round trip: %+v err=%v", out, err)
	}
}
