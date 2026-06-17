// Package dsrepl implements the DS_REPL_*_BLOB structures defined in [MS-ADTS]
// section 2.2. These are the binary (";binary" qualified) LDAP representations of
// the replication state structures otherwise returned by the IDL_DRSGetReplInfo
// RPC method, as read from the rootDSE attributes msDS-ReplAllInboundNeighbors,
// msDS-ReplConnectionFailures, msDS-ReplLinkFailures, msDS-ReplPendingOps,
// msDS-ReplQueueStatistics, msDS-NCReplCursors, msDS-ReplAttributeMetaData, and
// msDS-ReplValueMetaData.
//
// Each structure is a fixed-size header in which string fields are stored as
// 32-bit byte offsets (the osz* members), relative to the start of the structure,
// pointing into a trailing variable-length data region of packed, null-terminated
// UTF-16LE strings. All multibyte fields use little-endian byte ordering.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/a38fb14a-5f54-41ad-8875-c0c716afd53b
package dsrepl

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// readOffsetString reads a null-terminated UTF-16LE string from blob, located at
// the given 32-bit offset relative to the start of blob.
//
// Per [MS-ADTS], an offset of 0 denotes a NULL string (the field is absent); in
// that case nil is returned. A non-zero offset that points directly at a null
// terminator yields a non-nil pointer to an empty string, preserving the
// distinction between an absent and a present-but-empty value.
//
// Parameters:
// - blob: The full structure bytes (offsets are relative to its start).
// - offset: The 32-bit offset of the string, as stored in a header field.
//
// Returns:
// - A pointer to the decoded string, or nil if the offset is 0.
// - An error if the offset is out of range or the string is not null-terminated.
func readOffsetString(blob []byte, offset uint32) (*string, error) {
	if offset == 0 {
		return nil, nil
	}

	if int(offset) > len(blob) {
		return nil, fmt.Errorf("string offset %d is out of range (blob is %d bytes)", offset, len(blob))
	}

	// Scan for the UTF-16LE null terminator (two 0x00 bytes on an even boundary).
	end := -1
	for i := int(offset); i+1 < len(blob); i += 2 {
		if blob[i] == 0x00 && blob[i+1] == 0x00 {
			end = i
			break
		}
	}
	if end == -1 {
		return nil, fmt.Errorf("string at offset %d is not null-terminated", offset)
	}

	s := utf16.DecodeUTF16LE(blob[offset:end])
	return &s, nil
}

// describeOszString formats an optional offset string for human-readable output,
// rendering a NULL (nil) value as <null> and any present value quoted.
func describeOszString(s *string) string {
	if s == nil {
		return "<null>"
	}
	return fmt.Sprintf("%q", *s)
}

// dataRegion accumulates the variable-length trailing data of a DS_REPL_*_BLOB
// and assigns each appended item a 32-bit offset relative to the start of the
// structure. Offsets account for the fixed header that precedes the data region.
type dataRegion struct {
	// base is the size, in bytes, of the fixed header that precedes the data
	// region. Offsets are computed as base + position within buf.
	base uint32
	// buf holds the packed data bytes (strings and buffers).
	buf []byte
}

// newDataRegion creates a dataRegion for a structure whose fixed header is
// headerLen bytes long.
func newDataRegion(headerLen int) *dataRegion {
	return &dataRegion{base: uint32(headerLen), buf: make([]byte, 0)}
}

// addString appends s as a null-terminated UTF-16LE string and returns the
// 32-bit offset to use in the header. A nil pointer is encoded as a NULL pointer
// (offset 0) with nothing appended, matching the [MS-ADTS] semantics that an
// absent optional string is represented by a zero offset. A non-nil pointer to an
// empty string is encoded as a present, zero-length string (just a terminator).
func (d *dataRegion) addString(s *string) uint32 {
	if s == nil {
		return 0
	}

	offset := d.base + uint32(len(d.buf))
	d.buf = append(d.buf, utf16.EncodeUTF16LE(*s)...)
	// UTF-16LE null terminator.
	d.buf = append(d.buf, 0x00, 0x00)

	return offset
}

// addBytes appends a raw buffer and returns the 32-bit offset to use in the
// header. A nil/empty buffer is encoded as a NULL pointer (offset 0). The data
// region is kept 32-bit aligned, as required for the pbData buffer of
// DS_REPL_VALUE_META_DATA_BLOB.
func (d *dataRegion) addBytes(b []byte) uint32 {
	if len(b) == 0 {
		return 0
	}

	offset := d.base + uint32(len(d.buf))
	d.buf = append(d.buf, b...)
	for len(d.buf)%4 != 0 {
		d.buf = append(d.buf, 0x00)
	}

	return offset
}

// bytes returns the accumulated data-region bytes.
func (d *dataRegion) bytes() []byte {
	return d.buf
}
