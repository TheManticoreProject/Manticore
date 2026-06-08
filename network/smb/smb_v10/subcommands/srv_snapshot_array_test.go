package subcommands

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

func TestFSCTLEnumerateSnapshotsCode(t *testing.T) {
	if FSCTL_SRV_ENUMERATE_SNAPSHOTS != 0x00144064 {
		t.Errorf("FSCTL_SRV_ENUMERATE_SNAPSHOTS: got 0x%08x want 0x00144064", FSCTL_SRV_ENUMERATE_SNAPSHOTS)
	}
}

func TestSrvSnapshotArrayRoundTrip(t *testing.T) {
	in := SrvSnapshotArray{
		NumberOfSnapShots: 2,
		Snapshots: []string{
			"@GMT-2020.01.02-03.04.05",
			"@GMT-2021.06.07-08.09.10",
		},
	}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Header: NumberOfSnapShots=2, NumberOfSnapShotsReturned=2, SnapShotArraySize=len(multiSZ).
	if got := binary.LittleEndian.Uint32(raw[0:4]); got != 2 {
		t.Errorf("NumberOfSnapShots: got %d want 2", got)
	}
	if got := binary.LittleEndian.Uint32(raw[4:8]); got != 2 {
		t.Errorf("NumberOfSnapShotsReturned: got %d want 2", got)
	}
	// Each label is 24 UTF-16 units; multi-SZ = (24+1)*2 entries-with-NUL + 1 final NUL.
	wantUnits := (24 + 1) + (24 + 1) + 1
	if got := binary.LittleEndian.Uint32(raw[8:12]); int(got) != wantUnits*2 {
		t.Errorf("SnapShotArraySize: got %d want %d", got, wantUnits*2)
	}

	var out SrvSnapshotArray
	n, err := out.Unmarshal(raw)
	if err != nil || n != len(raw) {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
	if out.NumberOfSnapShots != 2 || out.NumberOfSnapShotsReturned != 2 || len(out.Snapshots) != 2 {
		t.Fatalf("round trip counts: %+v", out)
	}
	if out.Snapshots[0] != in.Snapshots[0] || out.Snapshots[1] != in.Snapshots[1] {
		t.Errorf("snapshots round trip: got %q want %q", out.Snapshots, in.Snapshots)
	}
}

func TestSrvSnapshotArrayEmpty(t *testing.T) {
	// A response advertising 5 available snapshots but returning none: the multi-SZ is two
	// UTF-16 NUL characters (4 octets).
	in := SrvSnapshotArray{NumberOfSnapShots: 5}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := make([]byte, 12+4)
	binary.LittleEndian.PutUint32(want[0:4], 5)
	binary.LittleEndian.PutUint32(want[4:8], 0)
	binary.LittleEndian.PutUint32(want[8:12], 4)
	// trailing 4 octets are the two NUL WCHARs (already zero)
	if !bytes.Equal(raw, want) {
		t.Errorf("empty SRV_SNAPSHOT_ARRAY:\n got % x\nwant % x", raw, want)
	}

	var out SrvSnapshotArray
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.NumberOfSnapShots != 5 || len(out.Snapshots) != 0 {
		t.Errorf("empty round trip: %+v", out)
	}
}

// TestSrvSnapshotArrayGoldenList checks the exact multi-SZ encoding of a single label.
func TestSrvSnapshotArrayGoldenList(t *testing.T) {
	in := SrvSnapshotArray{NumberOfSnapShots: 1, Snapshots: []string{"@GMT-2020.01.02-03.04.05"}}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	multiSZ := raw[12:]
	want := []byte{}
	for _, u := range utf16.Encode([]rune("@GMT-2020.01.02-03.04.05")) {
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, u)
		want = append(want, b...)
	}
	want = append(want, 0x00, 0x00) // per-entry NUL
	want = append(want, 0x00, 0x00) // list-terminating NUL
	if !bytes.Equal(multiSZ, want) {
		t.Errorf("multi-SZ:\n got % x\nwant % x", multiSZ, want)
	}
}
