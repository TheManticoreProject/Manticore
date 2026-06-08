package subcommands

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// sidS110 is the binary form of the well-known SID S-1-1-0 (Everyone): Revision(1),
// SubAuthorityCount(1), IdentifierAuthority(6, big-endian = 1), one SubAuthority(4 = 0).
var sidS110 = []byte{0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}

func TestNtTransQueryQuotaRequestParametersGolden(t *testing.T) {
	p := NtTransQueryQuotaRequestParameters{FID: 0x4002, ReturnSingleEntry: true}
	got, err := p.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x02, 0x40, // FID
		0x01,                   // ReturnSingleEntry = true
		0x00,                   // RestartScan = false
		0x00, 0x00, 0x00, 0x00, // SidListLength
		0x00, 0x00, 0x00, 0x00, // StartSidLength
		0x00, 0x00, 0x00, 0x00, // StartSidOffset
	}
	if !bytes.Equal(got, want) {
		t.Errorf("NT_TRANSACT_QUERY_QUOTA params:\n got % x\nwant % x", got, want)
	}
	var out NtTransQueryQuotaRequestParameters
	if _, err := out.Unmarshal(got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != p {
		t.Errorf("round trip: got %+v want %+v", out, p)
	}
}

func TestNtTransQuotaResponseParametersRoundTrip(t *testing.T) {
	p := NtTransQuotaResponseParameters{DataLength: 52}
	raw, err := p.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 4 || binary.LittleEndian.Uint32(raw) != 52 {
		t.Fatalf("DataLength marshal: % x", raw)
	}
	var out NtTransQuotaResponseParameters
	if _, err := out.Unmarshal(raw); err != nil || out.DataLength != 52 {
		t.Fatalf("round trip: %+v err=%v", out, err)
	}
}

func TestFileQuotaInformationRoundTrip(t *testing.T) {
	q := FileQuotaInformation{
		NextEntryOffset: 0,
		ChangeTime:      0x01C6841B620C8459,
		QuotaUsed:       1024,
		QuotaThreshold:  -1, // no warning threshold
		QuotaLimit:      -2, // delete the entry
		Sid:             sidS110,
	}
	raw, err := q.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != fileQuotaInformationFixedSize+len(sidS110) {
		t.Fatalf("length: got %d want %d", len(raw), fileQuotaInformationFixedSize+len(sidS110))
	}
	// SidLength is derived into octets [4:8].
	if got := binary.LittleEndian.Uint32(raw[4:8]); int(got) != len(sidS110) {
		t.Errorf("SidLength: got %d want %d", got, len(sidS110))
	}
	if !bytes.Equal(raw[fileQuotaInformationFixedSize:], sidS110) {
		t.Errorf("Sid tail:\n got % x\nwant % x", raw[fileQuotaInformationFixedSize:], sidS110)
	}

	var out FileQuotaInformation
	n, err := out.Unmarshal(raw)
	if err != nil || n != len(raw) {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
	if out.ChangeTime != q.ChangeTime || out.QuotaUsed != q.QuotaUsed ||
		out.QuotaThreshold != q.QuotaThreshold || out.QuotaLimit != q.QuotaLimit ||
		!bytes.Equal(out.Sid, q.Sid) {
		t.Errorf("round trip mismatch: %+v", out)
	}
}

func TestFileGetQuotaInformationRoundTrip(t *testing.T) {
	e := FileGetQuotaInformation{NextEntryOffset: 0, Sid: sidS110}
	raw, err := e.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != fileGetQuotaInformationFixedSize+len(sidS110) {
		t.Fatalf("length: got %d want %d", len(raw), fileGetQuotaInformationFixedSize+len(sidS110))
	}
	if got := binary.LittleEndian.Uint32(raw[4:8]); int(got) != len(sidS110) {
		t.Errorf("SidLength: got %d want %d", got, len(sidS110))
	}
	var out FileGetQuotaInformation
	n, err := out.Unmarshal(raw)
	if err != nil || n != len(raw) {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
	if out.NextEntryOffset != 0 || !bytes.Equal(out.Sid, sidS110) {
		t.Errorf("round trip mismatch: %+v", out)
	}
}
