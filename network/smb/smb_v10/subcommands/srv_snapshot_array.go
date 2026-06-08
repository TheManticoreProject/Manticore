package subcommands

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

// FSCTL_SRV_ENUMERATE_SNAPSHOTS enumerates the available previous-version (snapshot /
// shadow-copy) timestamps of an open file or directory ([MS-SMB] section 2.2.7.3). Like
// the copychunk FSCTLs, it is carried as the NT_TRANSACT_IOCTL FunctionCode of an
// SMB_COM_NT_TRANSACT request; the SrvSnapshotArray below is the response NT_Trans_Data.
const FSCTL_SRV_ENUMERATE_SNAPSHOTS uint32 = 0x00144064

const srvSnapshotArrayHeaderSize = 12 // NumberOfSnapShots(4) + NumberOfSnapShotsReturned(4) + SnapShotArraySize(4)

// SrvSnapshotArray is the NT_Trans_Data of an FSCTL_SRV_ENUMERATE_SNAPSHOTS response
// ([MS-SMB] section 2.2.7.3.2 SRV_SNAPSHOT_ARRAY). The on-the-wire SnapShotArraySize and
// the SnapShotMultiSZ list are derived from / parsed into Snapshots.
type SrvSnapshotArray struct {
	// NumberOfSnapShots (4 bytes): total number of snapshots the object store has of this
	// file. May exceed len(Snapshots) if the client's buffer could not hold them all.
	NumberOfSnapShots uint32

	// NumberOfSnapShotsReturned (4 bytes): the number of snapshots actually returned in this
	// response (equals len(Snapshots) on marshal).
	NumberOfSnapShotsReturned uint32

	// Snapshots: the returned snapshot labels, each of the form "@GMT-YYYY.MM.DD-HH.MM.SS".
	// On the wire they are a NUL-terminated UTF-16LE multi-SZ (an extra terminating NUL
	// closes the list; an empty list is two NUL characters).
	Snapshots []string
}

// snapshotMultiSZ builds the UTF-16LE multi-SZ encoding of the snapshot list: each entry
// NUL-terminated, the whole list closed by one additional NUL; an empty list is two NULs
// ([MS-SMB] section 2.2.7.3.2).
func (a *SrvSnapshotArray) snapshotMultiSZ() []uint16 {
	if len(a.Snapshots) == 0 {
		return []uint16{0, 0}
	}
	units := []uint16{}
	for _, s := range a.Snapshots {
		units = append(units, utf16.Encode([]rune(s))...)
		units = append(units, 0) // per-entry NUL terminator
	}
	units = append(units, 0) // additional list-terminating NUL
	return units
}

// Marshal serializes the SRV_SNAPSHOT_ARRAY: the three counts followed by the multi-SZ.
// NumberOfSnapShotsReturned and SnapShotArraySize are set from the Snapshots list, while
// NumberOfSnapShots is taken from the field (it can legitimately exceed the returned count).
func (a *SrvSnapshotArray) Marshal() ([]byte, error) {
	units := a.snapshotMultiSZ()
	multiSZ := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(multiSZ[i*2:], u)
	}

	b := make([]byte, srvSnapshotArrayHeaderSize+len(multiSZ))
	binary.LittleEndian.PutUint32(b[0:4], a.NumberOfSnapShots)
	binary.LittleEndian.PutUint32(b[4:8], uint32(len(a.Snapshots)))
	binary.LittleEndian.PutUint32(b[8:12], uint32(len(multiSZ)))
	copy(b[srvSnapshotArrayHeaderSize:], multiSZ)
	return b, nil
}

// Unmarshal parses an SRV_SNAPSHOT_ARRAY, returning the number of bytes consumed. The
// multi-SZ is decoded into Snapshots; a probe response that carries only the header (no
// room for the list) yields an empty Snapshots slice.
func (a *SrvSnapshotArray) Unmarshal(data []byte) (int, error) {
	if len(data) < srvSnapshotArrayHeaderSize {
		return 0, fmt.Errorf("subcommands: SRV_SNAPSHOT_ARRAY requires at least %d bytes, got %d", srvSnapshotArrayHeaderSize, len(data))
	}
	a.NumberOfSnapShots = binary.LittleEndian.Uint32(data[0:4])
	a.NumberOfSnapShotsReturned = binary.LittleEndian.Uint32(data[4:8])
	arraySize := int(binary.LittleEndian.Uint32(data[8:12]))

	// The list occupies SnapShotArraySize octets, but a buffer-probe response may omit it,
	// so read only what is actually present.
	avail := len(data) - srvSnapshotArrayHeaderSize
	if arraySize > avail {
		arraySize = avail
	}
	multiSZ := data[srvSnapshotArrayHeaderSize : srvSnapshotArrayHeaderSize+arraySize]

	a.Snapshots = nil
	var cur []uint16
	for i := 0; i+1 < len(multiSZ); i += 2 {
		u := binary.LittleEndian.Uint16(multiSZ[i:])
		if u == 0 {
			if len(cur) == 0 {
				break // empty entry → list terminator
			}
			a.Snapshots = append(a.Snapshots, string(utf16.Decode(cur)))
			cur = nil
			continue
		}
		cur = append(cur, u)
	}
	return srvSnapshotArrayHeaderSize + arraySize, nil
}
