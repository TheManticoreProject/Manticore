package informationlevels_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/informationlevels"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

func TestQueryFileBasicInfoRoundTrip(t *testing.T) {
	in := &informationlevels.SMB_QUERY_FILE_BASIC_INFO{
		Creationtime:      types.FILETIME{DwLowDateTime: 1, DwHighDateTime: 2},
		Lastaccesstime:    types.FILETIME{DwLowDateTime: 3, DwHighDateTime: 4},
		Lastwritetime:     types.FILETIME{DwLowDateTime: 5, DwHighDateTime: 6},
		Lastchangetime:    types.FILETIME{DwLowDateTime: 7, DwHighDateTime: 8},
		Extfileattributes: types.ATTR_DIRECTORY,
		Reserved:          0,
	}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 40 {
		t.Fatalf("expected 40 bytes, got %d", len(raw))
	}
	out := &informationlevels.SMB_QUERY_FILE_BASIC_INFO{}
	if n, err := out.Unmarshal(raw); err != nil || n != 40 {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
	if *out != *in {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

func TestQueryFileStandardInfoRoundTrip(t *testing.T) {
	in := &informationlevels.SMB_QUERY_FILE_STANDARD_INFO{
		Numberoflinks: 3,
		Deletepending: 1,
		Directory:     1,
	}
	in.Allocationsize.QuadPart = 0x1000
	in.Endoffile.QuadPart = 0x0ABC
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 22 {
		t.Fatalf("expected 22 bytes, got %d", len(raw))
	}
	out := &informationlevels.SMB_QUERY_FILE_STANDARD_INFO{}
	if n, err := out.Unmarshal(raw); err != nil || n != 22 {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
	if *out != *in {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

func TestQueryFileEaInfoRoundTrip(t *testing.T) {
	in := &informationlevels.SMB_QUERY_FILE_EA_INFO{Easize: 0xDEADBEEF}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(raw))
	}
	out := &informationlevels.SMB_QUERY_FILE_EA_INFO{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Easize != in.Easize {
		t.Errorf("EaSize = 0x%X, want 0x%X", out.Easize, in.Easize)
	}
}

func TestQueryFileCompressionInfoRoundTrip(t *testing.T) {
	in := &informationlevels.SMB_QUERY_FILE_COMRESSION_INFO{
		Compressionformat:    0x0002,
		Compressionunitshift: 16,
		Chunkshift:           12,
		Clustershift:         4,
		Reserved:             [3]types.UCHAR{0, 0, 0},
	}
	in.Compressedfilesize.QuadPart = 0x123456
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 16 {
		t.Fatalf("expected 16 bytes, got %d", len(raw))
	}
	out := &informationlevels.SMB_QUERY_FILE_COMRESSION_INFO{}
	if n, err := out.Unmarshal(raw); err != nil || n != 16 {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
	if *out != *in {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

// nameInfoLike is satisfied by both SMB_QUERY_FILE_NAME_INFO and
// SMB_QUERY_FILE_ALT_NAME_INFO (identical layouts) for shared round-trip testing.
func TestQueryFileNameInfoRoundTrip(t *testing.T) {
	// UTF-16LE "ab.txt"
	name := []types.UCHAR{'a', 0, 'b', 0, '.', 0, 't', 0, 'x', 0, 't', 0}

	in := &informationlevels.SMB_QUERY_FILE_NAME_INFO{Filename: name}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 4+len(name) {
		t.Fatalf("expected %d bytes, got %d", 4+len(name), len(raw))
	}
	out := &informationlevels.SMB_QUERY_FILE_NAME_INFO{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if int(out.Filenamelength) != len(name) || !bytes.Equal([]byte(out.Filename), []byte(name)) {
		t.Errorf("FileName mismatch: len=%d bytes=% x", out.Filenamelength, []byte(out.Filename))
	}

	alt := &informationlevels.SMB_QUERY_FILE_ALT_NAME_INFO{Filename: name}
	altRaw, err := alt.Marshal()
	if err != nil {
		t.Fatalf("ALT Marshal: %v", err)
	}
	altOut := &informationlevels.SMB_QUERY_FILE_ALT_NAME_INFO{}
	if _, err := altOut.Unmarshal(altRaw); err != nil {
		t.Fatalf("ALT Unmarshal: %v", err)
	}
	if !bytes.Equal([]byte(altOut.Filename), []byte(name)) {
		t.Errorf("ALT FileName mismatch: % x", []byte(altOut.Filename))
	}
}

func TestQueryFileStreamInfoRoundTrip(t *testing.T) {
	// UTF-16LE "::$DATA"
	streamName := []types.UCHAR{':', 0, ':', 0, '$', 0, 'D', 0, 'A', 0, 'T', 0, 'A', 0}

	in := &informationlevels.SMB_QUERY_FILE_STREAM_INFO{
		Nextentryoffset: 0,
		Streamname:      streamName,
	}
	in.Streamsize.QuadPart = 0x4000
	in.Streamallocationsize.QuadPart = 0x4096
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 24+len(streamName) {
		t.Fatalf("expected %d bytes, got %d", 24+len(streamName), len(raw))
	}
	out := &informationlevels.SMB_QUERY_FILE_STREAM_INFO{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Streamsize.QuadPart != in.Streamsize.QuadPart ||
		out.Streamallocationsize.QuadPart != in.Streamallocationsize.QuadPart ||
		int(out.Streamnamelength) != len(streamName) ||
		!bytes.Equal([]byte(out.Streamname), []byte(streamName)) {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

func TestQueryFileAllInfoRoundTrip(t *testing.T) {
	name := []types.UCHAR{'f', 0, 'o', 0, 'o', 0}

	in := &informationlevels.SMB_QUERY_FILE_ALL_INFO{
		Creationtime:      types.FILETIME{DwLowDateTime: 0x11, DwHighDateTime: 0x22},
		Lastaccesstime:    types.FILETIME{DwLowDateTime: 0x33, DwHighDateTime: 0x44},
		Lastwritetime:     types.FILETIME{DwLowDateTime: 0x55, DwHighDateTime: 0x66},
		Lastchangetime:    types.FILETIME{DwLowDateTime: 0x77, DwHighDateTime: 0x88},
		Extfileattributes: types.ATTR_ARCHIVE,
		Reserved1:         0,
		Numberoflinks:     1,
		Deletepending:     0,
		Directory:         0,
		Reserved2:         0,
		Easize:            0,
		Filename:          name,
	}
	in.Allocationsize.QuadPart = 0x2000
	in.Endoffile.QuadPart = 0x1234

	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// 4 FILETIMEs (32) + 40 fixed bytes + FileName.
	if len(raw) != 72+len(name) {
		t.Fatalf("expected %d bytes, got %d", 72+len(name), len(raw))
	}
	out := &informationlevels.SMB_QUERY_FILE_ALL_INFO{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Allocationsize.QuadPart != in.Allocationsize.QuadPart {
		t.Errorf("AllocationSize = 0x%X, want 0x%X", out.Allocationsize.QuadPart, in.Allocationsize.QuadPart)
	}
	if out.Endoffile.QuadPart != in.Endoffile.QuadPart {
		t.Errorf("EndOfFile = 0x%X, want 0x%X", out.Endoffile.QuadPart, in.Endoffile.QuadPart)
	}
	if int(out.Filenamelength) != len(name) || !bytes.Equal([]byte(out.Filename), []byte(name)) {
		t.Errorf("FileName mismatch: len=%d bytes=% x", out.Filenamelength, []byte(out.Filename))
	}
	if out.Creationtime != in.Creationtime || out.Lastchangetime != in.Lastchangetime {
		t.Errorf("timestamp mismatch")
	}
}

func TestQueryFileInfoUnmarshalBounds(t *testing.T) {
	if _, err := (&informationlevels.SMB_QUERY_FILE_STANDARD_INFO{}).Unmarshal(make([]byte, 21)); err == nil {
		t.Error("STANDARD: expected error on 21-byte buffer")
	}
	if _, err := (&informationlevels.SMB_QUERY_FILE_COMRESSION_INFO{}).Unmarshal(make([]byte, 15)); err == nil {
		t.Error("COMPRESSION: expected error on 15-byte buffer")
	}
	// NAME_INFO with a length that overruns the buffer.
	bad := []byte{0xFF, 0x00, 0x00, 0x00} // FileNameLength=255, no payload
	if _, err := (&informationlevels.SMB_QUERY_FILE_NAME_INFO{}).Unmarshal(bad); err == nil {
		t.Error("NAME: expected error when FileNameLength overruns buffer")
	}
}
