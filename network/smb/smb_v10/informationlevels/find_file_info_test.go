package informationlevels_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/informationlevels"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

func ft(lo, hi uint32) types.FILETIME { return types.FILETIME{DwLowDateTime: lo, DwHighDateTime: hi} }

func TestFindFileNamesInfoRoundTrip(t *testing.T) {
	name := []types.UCHAR{'a', 0, 'b', 0}
	in := &informationlevels.SMB_FIND_FILE_NAMES_INFO{Nextentryoffset: 0x40, Fileindex: 0, Filename: name}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 12+len(name) {
		t.Fatalf("expected %d bytes, got %d", 12+len(name), len(raw))
	}
	out := &informationlevels.SMB_FIND_FILE_NAMES_INFO{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Nextentryoffset != in.Nextentryoffset || int(out.Filenamelength) != len(name) || !bytes.Equal([]byte(out.Filename), []byte(name)) {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

func TestFindFileDirectoryInfoRoundTrip(t *testing.T) {
	name := []types.UCHAR{'f', 0, 'o', 0, 'o', 0}
	in := &informationlevels.SMB_FIND_FILE_DIRECTORY_INFO{
		Nextentryoffset:    0x80,
		Fileindex:          0,
		Creationtime:       ft(1, 2),
		Lastaccesstime:     ft(3, 4),
		Lastwritetime:      ft(5, 6),
		Lastattrchangetime: ft(7, 8),
		Extfileattributes:  types.ATTR_DIRECTORY,
		Filename:           name,
	}
	in.Endoffile.QuadPart = 0x1111
	in.Allocationsize.QuadPart = 0x2000
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 64+len(name) {
		t.Fatalf("expected %d bytes, got %d", 64+len(name), len(raw))
	}
	out := &informationlevels.SMB_FIND_FILE_DIRECTORY_INFO{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Lastattrchangetime != in.Lastattrchangetime || out.Endoffile.QuadPart != in.Endoffile.QuadPart || !bytes.Equal([]byte(out.Filename), []byte(name)) {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

func TestFindFileFullDirectoryInfoRoundTrip(t *testing.T) {
	name := []types.UCHAR{'b', 0, 'a', 0, 'r', 0}
	in := &informationlevels.SMB_FIND_FILE_FULL_DIRECTORY_INFO{
		Nextentryoffset:    0,
		Creationtime:       ft(9, 10),
		Lastaccesstime:     ft(11, 12),
		Lastwritetime:      ft(13, 14),
		Lastattrchangetime: ft(15, 16),
		Extfileattributes:  types.ATTR_ARCHIVE,
		Easize:             0x55,
		Filename:           name,
	}
	in.Endoffile.QuadPart = 0x3333
	in.Allocationsize.QuadPart = 0x4000
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 68+len(name) {
		t.Fatalf("expected %d bytes, got %d", 68+len(name), len(raw))
	}
	out := &informationlevels.SMB_FIND_FILE_FULL_DIRECTORY_INFO{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Easize != in.Easize || int(out.Filenamelength) != len(name) || !bytes.Equal([]byte(out.Filename), []byte(name)) {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

func TestFindFileBothDirectoryInfoRoundTrip(t *testing.T) {
	name := []types.UCHAR{'l', 0, 'o', 0, 'n', 0, 'g', 0}
	var short [24]types.UCHAR
	copy(short[:], []types.UCHAR{'L', 0, 'O', 0, 'N', 0, 'G', 0, '~', 0, '1', 0})

	in := &informationlevels.SMB_FIND_FILE_BOTH_DIRECTORY_INFO{
		Nextentryoffset:   0,
		Creationtime:      ft(0x11, 0x22),
		Lastaccesstime:    ft(0x33, 0x44),
		Lastwritetime:     ft(0x55, 0x66),
		Lastchangetime:    ft(0x77, 0x88),
		Extfileattributes: types.ATTR_NORMAL,
		Easize:            0,
		Shortnamelength:   12,
		Reserved:          0,
		Shortname:         short,
		Filename:          name,
	}
	in.Endoffile.QuadPart = 0xABCD
	in.Allocationsize.QuadPart = 0x1_0000

	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 94+len(name) {
		t.Fatalf("expected %d bytes, got %d", 94+len(name), len(raw))
	}
	out := &informationlevels.SMB_FIND_FILE_BOTH_DIRECTORY_INFO{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Shortnamelength != 12 || out.Shortname != short {
		t.Errorf("ShortName mismatch: len=%d bytes=% x", out.Shortnamelength, out.Shortname)
	}
	if int(out.Filenamelength) != len(name) || !bytes.Equal([]byte(out.Filename), []byte(name)) {
		t.Errorf("FileName mismatch: len=%d bytes=% x", out.Filenamelength, []byte(out.Filename))
	}
	if out.Endoffile.QuadPart != in.Endoffile.QuadPart || out.Lastchangetime != in.Lastchangetime {
		t.Errorf("fixed field mismatch")
	}
}

func TestFindFileInfoUnmarshalBounds(t *testing.T) {
	if _, err := (&informationlevels.SMB_FIND_FILE_DIRECTORY_INFO{}).Unmarshal(make([]byte, 63)); err == nil {
		t.Error("DIRECTORY: expected error on 63-byte buffer")
	}
	if _, err := (&informationlevels.SMB_FIND_FILE_FULL_DIRECTORY_INFO{}).Unmarshal(make([]byte, 67)); err == nil {
		t.Error("FULL_DIRECTORY: expected error on 67-byte buffer")
	}
	if _, err := (&informationlevels.SMB_FIND_FILE_BOTH_DIRECTORY_INFO{}).Unmarshal(make([]byte, 93)); err == nil {
		t.Error("BOTH_DIRECTORY: expected error on 93-byte buffer")
	}
}
