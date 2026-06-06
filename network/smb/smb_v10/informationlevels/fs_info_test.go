package informationlevels_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/informationlevels"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

func TestFsSizeInfoRoundTrip(t *testing.T) {
	in := &informationlevels.SMB_QUERY_FS_SIZE_INFO{Sectorsperallocationunit: 8, Bytespersector: 512}
	in.Totalallocationunits.QuadPart = 0x10000
	in.Totalfreeallocationunits.QuadPart = 0x8000
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 24 {
		t.Fatalf("expected 24 bytes, got %d", len(raw))
	}
	out := &informationlevels.SMB_QUERY_FS_SIZE_INFO{}
	if n, err := out.Unmarshal(raw); err != nil || n != 24 {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
	if *out != *in {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

func TestFsDeviceInfoRoundTrip(t *testing.T) {
	in := &informationlevels.SMB_QUERY_FS_DEVICE_INFO{Devicetype: 0x0007, Devicecharacteristics: 0x0010}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 8 {
		t.Fatalf("expected 8 bytes, got %d", len(raw))
	}
	out := &informationlevels.SMB_QUERY_FS_DEVICE_INFO{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if *out != *in {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

func TestInfoAllocationRoundTrip(t *testing.T) {
	in := &informationlevels.SMB_INFO_ALLOCATION{
		Idfilesystem:   0,
		Csectorunit:    8,
		Cunit:          0x100000,
		Cunitavailable: 0x80000,
		Cbsector:       512,
	}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 18 {
		t.Fatalf("expected 18 bytes, got %d", len(raw))
	}
	out := &informationlevels.SMB_INFO_ALLOCATION{}
	if n, err := out.Unmarshal(raw); err != nil || n != 18 {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
	if *out != *in {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

func TestFsVolumeInfoRoundTrip(t *testing.T) {
	label := []types.UCHAR{'D', 0, 'A', 0, 'T', 0, 'A', 0}
	in := &informationlevels.SMB_QUERY_FS_VOLUME_INFO{
		Volumecreationtime: types.FILETIME{DwLowDateTime: 0xAA, DwHighDateTime: 0xBB},
		Serialnumber:       0x12345678,
		Reserved:           0,
		Volumelabel:        label,
	}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 18+len(label) {
		t.Fatalf("expected %d bytes, got %d", 18+len(label), len(raw))
	}
	out := &informationlevels.SMB_QUERY_FS_VOLUME_INFO{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Serialnumber != in.Serialnumber || int(out.Volumelabelsize) != len(label) || !bytes.Equal([]byte(out.Volumelabel), []byte(label)) {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

func TestFsAttributeInfoRoundTrip(t *testing.T) {
	name := []types.UCHAR{'N', 0, 'T', 0, 'F', 0, 'S', 0}
	in := &informationlevels.SMB_QUERY_FS_ATTRIBUTE_INFO{
		Filesystemattributes:     0x00000007,
		Maxfilenamelengthinbytes: 255,
		Filesystemname:           name,
	}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 12+len(name) {
		t.Fatalf("expected %d bytes, got %d", 12+len(name), len(raw))
	}
	out := &informationlevels.SMB_QUERY_FS_ATTRIBUTE_INFO{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Filesystemattributes != in.Filesystemattributes || out.Maxfilenamelengthinbytes != in.Maxfilenamelengthinbytes ||
		int(out.Lengthoffilesystemname) != len(name) || !bytes.Equal([]byte(out.Filesystemname), []byte(name)) {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

func TestInfoVolumeRoundTrip(t *testing.T) {
	label := []types.UCHAR{'H', 'O', 'M', 'E'} // OEM, 1 byte/char
	in := &informationlevels.SMB_INFO_VOLUME{Ulvolserialnbr: 0xCAFEBABE, Volumelabel: label}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 5+len(label) {
		t.Fatalf("expected %d bytes, got %d", 5+len(label), len(raw))
	}
	if int(raw[4]) != len(label) {
		t.Errorf("cCharCount byte = %d, want %d", raw[4], len(label))
	}
	out := &informationlevels.SMB_INFO_VOLUME{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Ulvolserialnbr != in.Ulvolserialnbr || int(out.Ccharcount) != len(label) || !bytes.Equal([]byte(out.Volumelabel), []byte(label)) {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

func TestFsInfoUnmarshalBounds(t *testing.T) {
	if _, err := (&informationlevels.SMB_QUERY_FS_SIZE_INFO{}).Unmarshal(make([]byte, 23)); err == nil {
		t.Error("FS_SIZE: expected error on 23-byte buffer")
	}
	if _, err := (&informationlevels.SMB_INFO_ALLOCATION{}).Unmarshal(make([]byte, 17)); err == nil {
		t.Error("INFO_ALLOCATION: expected error on 17-byte buffer")
	}
	if _, err := (&informationlevels.SMB_QUERY_FS_VOLUME_INFO{}).Unmarshal(make([]byte, 17)); err == nil {
		t.Error("FS_VOLUME: expected error on 17-byte buffer")
	}
	if _, err := (&informationlevels.SMB_QUERY_FS_ATTRIBUTE_INFO{}).Unmarshal(make([]byte, 11)); err == nil {
		t.Error("FS_ATTRIBUTE: expected error on 11-byte buffer")
	}
}
