package subcommands

import (
	"bytes"
	"testing"
)

// exampleResumeKey is the 24-byte copychunk resume key from the [MS-SMB] section 2.2.7.2
// FSCTL_SRV_COPYCHUNK example trace.
var exampleResumeKey = [CopychunkResumeKeyLength]byte{
	0x2D, 0x0B, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
	0x59, 0x84, 0x0C, 0x62, 0x1B, 0x84, 0xC6, 0x01,
	0x08, 0x0E, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

func TestSrvCopychunkGolden(t *testing.T) {
	c := SrvCopychunk{SourceOffset: 0x01, DestinationOffset: 0x02, Length: 0x063C}
	got, err := c.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // SourceOffset
		0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // DestinationOffset
		0x3C, 0x06, 0x00, 0x00, // Length
		0x00, 0x00, 0x00, 0x00, // Reserved
	}
	if !bytes.Equal(got, want) {
		t.Errorf("SRV_COPYCHUNK:\n got % x\nwant % x", got, want)
	}

	var out SrvCopychunk
	n, err := out.Unmarshal(got)
	if err != nil || n != srvCopychunkSize {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
	if out != c {
		t.Errorf("round trip: got %+v want %+v", out, c)
	}
}

func TestSrvCopychunkCopyRoundTrip(t *testing.T) {
	in := SrvCopychunkCopy{
		CopychunkResumeKey: exampleResumeKey,
		Chunks: []SrvCopychunk{
			{SourceOffset: 0, DestinationOffset: 0, Length: 0x063C},
		},
	}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// Header: 24-byte key, ChunkCount=1, Reserved=0.
	wantHeader := append(append([]byte{}, exampleResumeKey[:]...), 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)
	if !bytes.Equal(raw[:32], wantHeader) {
		t.Errorf("SRV_COPYCHUNK_COPY header:\n got % x\nwant % x", raw[:32], wantHeader)
	}
	if len(raw) != 32+srvCopychunkSize {
		t.Fatalf("length: got %d want %d", len(raw), 32+srvCopychunkSize)
	}

	var out SrvCopychunkCopy
	n, err := out.Unmarshal(raw)
	if err != nil || n != len(raw) {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
	if out.CopychunkResumeKey != in.CopychunkResumeKey || len(out.Chunks) != 1 || out.Chunks[0] != in.Chunks[0] {
		t.Errorf("round trip mismatch: %+v", out)
	}
}

func TestSrvCopychunkResponseGolden(t *testing.T) {
	r := SrvCopychunkResponse{ChunksWritten: 1, ChunkBytesWritten: 0, TotalBytesWritten: 0x063C}
	got, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x01, 0x00, 0x00, 0x00, // ChunksWritten
		0x00, 0x00, 0x00, 0x00, // ChunkBytesWritten
		0x3C, 0x06, 0x00, 0x00, // TotalBytesWritten
	}
	if !bytes.Equal(got, want) {
		t.Errorf("SRV_COPYCHUNK_RESPONSE:\n got % x\nwant % x", got, want)
	}
	var out SrvCopychunkResponse
	if _, err := out.Unmarshal(got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != r {
		t.Errorf("round trip: got %+v want %+v", out, r)
	}
}

func TestSrvRequestResumeKeyResponseRoundTrip(t *testing.T) {
	in := SrvRequestResumeKeyResponse{CopychunkResumeKey: exampleResumeKey}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// 24-byte key followed by ContextLength = 0.
	want := append(append([]byte{}, exampleResumeKey[:]...), 0x00, 0x00, 0x00, 0x00)
	if !bytes.Equal(raw, want) {
		t.Errorf("resume-key response:\n got % x\nwant % x", raw, want)
	}
	var out SrvRequestResumeKeyResponse
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.CopychunkResumeKey != in.CopychunkResumeKey || len(out.Context) != 0 {
		t.Errorf("round trip mismatch: %+v", out)
	}
}

func TestFSCTLCopychunkCodes(t *testing.T) {
	if FSCTL_SRV_REQUEST_RESUME_KEY != 0x00140078 {
		t.Errorf("FSCTL_SRV_REQUEST_RESUME_KEY: got 0x%08x want 0x00140078", FSCTL_SRV_REQUEST_RESUME_KEY)
	}
	if FSCTL_SRV_COPYCHUNK != 0x001440F2 {
		t.Errorf("FSCTL_SRV_COPYCHUNK: got 0x%08x want 0x001440F2", FSCTL_SRV_COPYCHUNK)
	}
}
