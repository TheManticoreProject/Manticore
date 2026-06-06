package informationlevels_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/informationlevels"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

func dosDate(y uint16, m, d uint8) types.SMB_DATE {
	return types.SMB_DATE{Year: y, Month: m, Day: d}
}
func dosTime(h, m, twoSec uint8) types.SMB_TIME_DOS {
	return types.SMB_TIME_DOS{Hours: h, Minutes: m, TwoSeconds: twoSec}
}

// TestInfoStandardRoundTrip verifies SMB_INFO_STANDARD is 12 bytes (six 2-byte DOS
// date/time fields) and round-trips, confirming the 2-byte SMB_TIME_DOS layout.
func TestInfoStandardRoundTrip(t *testing.T) {
	in := &informationlevels.SMB_INFO_STANDARD{
		Creationdate:   dosDate(2021, 6, 15),
		Creationtime:   dosTime(12, 30, 10),
		Lastaccessdate: dosDate(2022, 1, 2),
		Lastaccesstime: dosTime(1, 2, 3),
		Lastwritedate:  dosDate(2023, 12, 31),
		Lastwritetime:  dosTime(23, 59, 29),
	}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 12 {
		t.Fatalf("expected 12 bytes, got %d", len(raw))
	}
	out := &informationlevels.SMB_INFO_STANDARD{}
	if n, err := out.Unmarshal(raw); err != nil || n != 12 {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
	if *out != *in {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

// TestInfoQueryEaSizeRoundTrip verifies SMB_INFO_QUERY_EA_SIZE is 26 bytes and
// round-trips, including the 2-byte date/time fields and SMB_FILE_ATTRIBUTES.
func TestInfoQueryEaSizeRoundTrip(t *testing.T) {
	in := &informationlevels.SMB_INFO_QUERY_EA_SIZE{
		Creationdate:   dosDate(2021, 6, 15),
		Creationtime:   dosTime(12, 30, 10),
		Lastaccessdate: dosDate(2022, 1, 2),
		Lastaccesstime: dosTime(1, 2, 3),
		Lastwritedate:  dosDate(2023, 12, 31),
		Lastwritetime:  dosTime(23, 59, 29),
		Filedatasize:   0x1000,
		Allocationsize: 0x200,
		Easize:         0x40,
	}
	in.Attributes.SetAttributes(0x0020)

	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 26 {
		t.Fatalf("expected 26 bytes, got %d", len(raw))
	}
	out := &informationlevels.SMB_INFO_QUERY_EA_SIZE{}
	if n, err := out.Unmarshal(raw); err != nil || n != 26 {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
	if *out != *in {
		t.Errorf("mismatch:\n in  %+v\n out %+v", in, out)
	}
}

// TestInfoIsNameValidIsEmpty verifies SMB_INFO_IS_NAME_VALID marshals to no bytes
// and unmarshals consuming nothing (success is conveyed via the SMB Header Status).
func TestInfoIsNameValidIsEmpty(t *testing.T) {
	in := &informationlevels.SMB_INFO_IS_NAME_VALID{}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 0 {
		t.Fatalf("expected 0 bytes, got %d", len(raw))
	}
	out := &informationlevels.SMB_INFO_IS_NAME_VALID{}
	if n, err := out.Unmarshal(raw); err != nil || n != 0 {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
}

// TestInfoEaListRoundTrip verifies the three SMB_FEA_LIST-wrapping information
// levels round-trip and that SizeOfListInBytes includes the 4-byte size field.
func TestInfoEaListRoundTrip(t *testing.T) {
	payload := []types.UCHAR{0x00, 0x04, 0x05, 0x00, 'F', 'O', 'O', 0x00, 'b', 'a', 'r', 0x00}

	t.Run("QueryAllEas", func(t *testing.T) {
		in := &informationlevels.SMB_INFO_QUERY_ALL_EAS{}
		in.Extendedattributelist.FEAList = payload
		raw, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if len(raw) != 4+len(payload) {
			t.Fatalf("expected %d bytes, got %d", 4+len(payload), len(raw))
		}
		out := &informationlevels.SMB_INFO_QUERY_ALL_EAS{}
		if _, err := out.Unmarshal(raw); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if int(out.Extendedattributelist.SizeOfListInBytes) != 4+len(payload) {
			t.Errorf("SizeOfListInBytes = %d, want %d", out.Extendedattributelist.SizeOfListInBytes, 4+len(payload))
		}
		if !bytes.Equal([]byte(out.Extendedattributelist.FEAList), []byte(payload)) {
			t.Errorf("FEAList mismatch: % x", []byte(out.Extendedattributelist.FEAList))
		}
	})

	t.Run("QueryEasFromList", func(t *testing.T) {
		in := &informationlevels.SMB_INFO_QUERY_EAS_FROM_LIST{}
		in.Extendedattributelist.FEAList = payload
		raw, _ := in.Marshal()
		out := &informationlevels.SMB_INFO_QUERY_EAS_FROM_LIST{}
		if _, err := out.Unmarshal(raw); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if !bytes.Equal([]byte(out.Extendedattributelist.FEAList), []byte(payload)) {
			t.Errorf("FEAList mismatch")
		}
	})

	t.Run("SetEas", func(t *testing.T) {
		in := &informationlevels.SMB_INFO_SET_EAS{}
		in.Extendedattributelist.FEAList = payload
		raw, _ := in.Marshal()
		out := &informationlevels.SMB_INFO_SET_EAS{}
		if _, err := out.Unmarshal(raw); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if !bytes.Equal([]byte(out.Extendedattributelist.FEAList), []byte(payload)) {
			t.Errorf("FEAList mismatch")
		}
	})
}

func TestLegacyEaInfoUnmarshalBounds(t *testing.T) {
	if _, err := (&informationlevels.SMB_INFO_STANDARD{}).Unmarshal(make([]byte, 11)); err == nil {
		t.Error("INFO_STANDARD: expected error on 11-byte buffer")
	}
	if _, err := (&informationlevels.SMB_INFO_QUERY_EA_SIZE{}).Unmarshal(make([]byte, 25)); err == nil {
		t.Error("INFO_QUERY_EA_SIZE: expected error on 25-byte buffer")
	}
	if _, err := (&informationlevels.SMB_INFO_QUERY_ALL_EAS{}).Unmarshal(make([]byte, 3)); err == nil {
		t.Error("QUERY_ALL_EAS: expected error on 3-byte buffer")
	}
}
