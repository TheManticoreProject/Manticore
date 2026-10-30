package mssrvs

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// diskEnumReply mirrors the NetrServerDiskEnum (opnum 23) response parameters.
type diskEnumReply struct {
	DiskInfoStruct DISK_ENUM_CONTAINER
	TotalEntries   ndr.DWORD
	ResumeHandle   *ndr.DWORD `ndr:"unique"`
	Status         ndr.DWORD  `ndr:"retval"`
}

// diskInfoHolder wraps a lone DISK_INFO so it is encoded as a field (the
// top-level ndr.Marshal walks a struct's fields rather than its hooks).
type diskInfoHolder struct {
	Info DISK_INFO
}

// diskEnumReplyWire builds a spec-format NetrServerDiskEnum reply listing the
// given drive letters, each DISK_INFO encoded as a varying WCHAR array.
func diskEnumReplyWire(letters ...uint16) []byte {
	var b []byte
	u32 := func(v uint32) { b = binary.LittleEndian.AppendUint32(b, v) }
	n := uint32(len(letters))
	u32(n)          // EntriesRead
	u32(0x00020000) // Buffer referent
	u32(n)          // maximum_count
	u32(0)          // offset
	u32(n)          // actual_count
	for _, c := range letters {
		u32(0) // DISK_INFO offset
		u32(3) // DISK_INFO actual_count: letter, ':' and the null terminator
		b = append(b, byte(c), 0, ':', 0, 0, 0)
		b = append(b, 0, 0) // pad to the next 4-octet boundary
	}
	u32(n)          // TotalEntries
	u32(0x00020004) // ResumeHandle referent
	u32(0)          // *ResumeHandle
	u32(0)          // Status (NERR_Success)
	return b
}

// TestDISK_INFO_DecodesVaryingArray decodes a spec-format NetrServerDiskEnum
// reply: each [string] WCHAR Disk[3] carries an offset and actual_count.
func TestDISK_INFO_DecodesVaryingArray(t *testing.T) {
	var r diskEnumReply
	if err := ndr.Unmarshal(diskEnumReplyWire('C', 'D'), &r); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	want := []DISK_INFO{{Disk: [3]uint16{'C', ':', 0}}, {Disk: [3]uint16{'D', ':', 0}}}
	if !reflect.DeepEqual(r.DiskInfoStruct.Buffer, want) {
		t.Errorf("Buffer = %v, want %v", r.DiskInfoStruct.Buffer, want)
	}
	if r.TotalEntries != 2 || r.Status != 0 {
		t.Errorf("TotalEntries = %d, Status = %d, want 2, 0", r.TotalEntries, r.Status)
	}
}

// TestDISK_INFO_EncodesVaryingArray checks a DISK_INFO is written with the
// offset and actual_count of a varying array, and that it round-trips.
func TestDISK_INFO_EncodesVaryingArray(t *testing.T) {
	in := DISK_INFO{Disk: [3]uint16{'C', ':', 0}}
	raw, err := ndr.Marshal(&diskInfoHolder{Info: in})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{0, 0, 0, 0, 3, 0, 0, 0, 'C', 0, ':', 0, 0, 0}
	if !bytes.Equal(raw, want) {
		t.Errorf("Marshal = % x, want % x", raw, want)
	}

	// A terminating empty entry carries only the null terminator.
	raw, err = ndr.Marshal(&diskInfoHolder{})
	if err != nil {
		t.Fatalf("Marshal empty: %v", err)
	}
	if want := []byte{0, 0, 0, 0, 1, 0, 0, 0, 0, 0}; !bytes.Equal(raw, want) {
		t.Errorf("Marshal empty = % x, want % x", raw, want)
	}

	roundTripConnFileSession(t, "DISK_ENUM_CONTAINER", DISK_ENUM_CONTAINER{
		EntriesRead: 3,
		Buffer:      []DISK_INFO{in, {Disk: [3]uint16{'D', ':', 0}}, {}},
	})
}

// TestDISK_INFO_RejectsOversizedCount rejects an actual_count beyond the
// three-WCHAR bound instead of overrunning Disk.
func TestDISK_INFO_RejectsOversizedCount(t *testing.T) {
	wire := []byte{0, 0, 0, 0, 4, 0, 0, 0, 'C', 0, ':', 0, '\\', 0, 0, 0}
	var out diskInfoHolder
	if err := ndr.Unmarshal(wire, &out); err == nil {
		t.Errorf("Unmarshal accepted actual_count 4 for a 3-WCHAR array")
	}
}
