package mssrvs

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// DISK_INFO contains the drive letter of a disk on the server ([MS-SRVS]
// 2.2.4.86). Disk is a fixed array of three WCHARs (drive letter, ':' and a
// null terminator).
//
// The IDL declares the field as [string] WCHAR Disk[3]. A fixed-bound [string]
// array is encoded in NDR as a *varying* array ([C706] 14.3.3.2): an offset and
// an actual_count (characters, including the null terminator) followed by that
// many WCHARs, with no maximum_count. The codec's plain fixed-array handling
// omits both counts, so DISK_INFO marshals itself.
type DISK_INFO struct {
	Disk [3]uint16
}

// Compile-time assertion that DISK_INFO encodes itself as NDR.
var _ ndr.Marshaler = (*DISK_INFO)(nil)

// AlignmentNDR reports the 4-octet alignment of the leading offset/actual_count words.
func (*DISK_INFO) AlignmentNDR() int { return 4 }

// MarshalNDR writes Disk as a varying WCHAR array: offset 0, actual_count (the
// characters up to and including the first null terminator, or all three when
// there is none), then those characters.
func (di *DISK_INFO) MarshalNDR(e *ndr.Encoder) error {
	n := len(di.Disk)
	for i, c := range di.Disk {
		if c == 0 {
			n = i + 1
			break
		}
	}
	writeDiskInfoCount(e, 0) // offset
	writeDiskInfoCount(e, uint64(n))
	for _, c := range di.Disk[:n] {
		e.WriteUint16(c)
	}
	return nil
}

// UnmarshalNDR reads a varying WCHAR array (offset, actual_count, characters)
// into Disk, rejecting a range that does not fit the three-WCHAR bound.
func (di *DISK_INFO) UnmarshalNDR(d *ndr.Decoder) error {
	offset, err := readDiskInfoCount(d)
	if err != nil {
		return fmt.Errorf("mssrvs: DISK_INFO offset: %w", err)
	}
	count, err := readDiskInfoCount(d)
	if err != nil {
		return fmt.Errorf("mssrvs: DISK_INFO actual_count: %w", err)
	}
	if offset > uint64(len(di.Disk)) || count > uint64(len(di.Disk))-offset {
		return fmt.Errorf("mssrvs: DISK_INFO offset %d + actual_count %d exceeds bound %d", offset, count, len(di.Disk))
	}
	di.Disk = [3]uint16{}
	for i := offset; i < offset+count; i++ {
		c, err := d.ReadUint16()
		if err != nil {
			return fmt.Errorf("mssrvs: DISK_INFO: %w", err)
		}
		di.Disk[i] = c
	}
	return nil
}

// writeDiskInfoCount writes a varying-array offset or actual_count: 4 octets
// 4-aligned under NDR20, 8 octets 8-aligned under NDR64 ([MS-RPCE] 2.2.5).
func writeDiskInfoCount(e *ndr.Encoder, v uint64) {
	if e.Syntax() == ndr.NDR64 {
		e.Align(8)
		e.WriteUint64(v)
		return
	}
	e.Align(4)
	e.WriteUint32(uint32(v))
}

// readDiskInfoCount reads a varying-array offset or actual_count in the
// decoder's transfer syntax.
func readDiskInfoCount(d *ndr.Decoder) (uint64, error) {
	if d.Syntax() == ndr.NDR64 {
		d.Align(8)
		return d.ReadUint64()
	}
	d.Align(4)
	v, err := d.ReadUint32()
	return uint64(v), err
}
