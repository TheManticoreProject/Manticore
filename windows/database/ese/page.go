package ese

import (
	"encoding/binary"
	"fmt"
)

// page is a parsed ESE page: its header fields plus the raw bytes (tags are read on
// demand from the end of the page).
type page struct {
	db                    *Database
	data                  []byte
	headerLen             int
	firstAvailablePageTag uint16
	pageFlags             uint32
	nextPageNumber        uint32
}

// newPage parses the page header. The layout depends on the format version/revision and
// page size (see impacket ESENT_PAGE_HEADER): the leading checksum field is 8 bytes on
// Windows 7+ format (4+4 or 4 on older), and pages larger than 8192 carry a trailing
// extended-checksum block, all of which only affects where the page body (tag values)
// begins.
func (d *Database) newPage(data []byte) (*page, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("ese: page too short: %d bytes", len(data))
	}
	var checksumLen int
	switch {
	case d.version < 0x620 || (d.version == 0x620 && d.revision < 0x0b):
		checksumLen = 8 // CheckSum(4) + PageNumber(4)
	case d.version == 0x620 && d.revision < 0x11:
		checksumLen = 8 // CheckSum(4) + ECCCheckSum(4)
	default:
		checksumLen = 8 // Windows 7+: single 8-byte CheckSum
	}
	// common block (32 bytes) immediately follows the checksum block.
	common := checksumLen
	headerLen := checksumLen + 32
	if d.version >= 0x620 && d.revision >= 0x11 && d.pageSize > 8192 {
		headerLen += 40 // extended_win7 trailing block
	}
	if len(data) < headerLen {
		return nil, fmt.Errorf("ese: page shorter than header (%d < %d)", len(data), headerLen)
	}

	p := &page{db: d, data: data, headerLen: headerLen}
	// common: LastModificationTime(8) Prev(4) Next(4) Father(4) AvailDataSize(2)
	//         AvailUncommitted(2) FirstAvailDataOffset(2) FirstAvailPageTag(2) PageFlags(4)
	p.nextPageNumber = binary.LittleEndian.Uint32(data[common+12 : common+16])
	p.firstAvailablePageTag = binary.LittleEndian.Uint16(data[common+26 : common+28])
	p.pageFlags = binary.LittleEndian.Uint32(data[common+28 : common+32])
	return p, nil
}

func (p *page) isLeaf() bool { return p.pageFlags&flagLeaf != 0 }

// getTag returns the page-tag flags and the tag's value bytes for tag number tagNum.
// Tags are a 4-byte array at the very end of the page, indexed from the end (tag 0 is
// the last 4 bytes). Each tag encodes a value size and an offset (relative to the page
// header length); the mask layout differs for >8192-byte Windows-7+ pages.
func (p *page) getTag(tagNum int) (flags uint8, value []byte, err error) {
	if tagNum >= int(p.firstAvailablePageTag) {
		return 0, nil, fmt.Errorf("ese: tag %d out of range (have %d)", tagNum, p.firstAvailablePageTag)
	}
	// tag entries occupy the last 4*FirstAvailablePageTag bytes; tag i is at
	// data[len - 4*(i+1) : len - 4*i].
	end := len(p.data) - 4*tagNum
	tag := p.data[end-4 : end]
	w0 := binary.LittleEndian.Uint16(tag[0:2])
	w1 := binary.LittleEndian.Uint16(tag[2:4])

	extended := p.db.version == 0x620 && p.db.revision >= 0x11 && p.db.pageSize > 8192
	var valueSize, valueOffset int
	if extended {
		valueSize = int(w0 & 0x7fff)
		valueOffset = int(w1 & 0x7fff)
	} else {
		valueSize = int(w0 & 0x1fff)
		flags = uint8((w1 & 0xe000) >> 13)
		valueOffset = int(w1 & 0x1fff)
	}
	start := p.headerLen + valueOffset
	if start < 0 || start+valueSize > len(p.data) {
		return 0, nil, fmt.Errorf("ese: tag %d value out of bounds", tagNum)
	}
	value = append([]byte(nil), p.data[start:start+valueSize]...)
	if extended {
		// For >8192 pages the per-tag flags live in the high bits of the value's second
		// byte; lift them out and clear them so the value parses cleanly.
		if len(value) > 1 {
			flags = value[1] >> 5
			value[1] &= 0x1f
		}
	}
	return flags, value, nil
}

// entryData strips the leaf/branch entry framing (optional common-key-size prefix, then
// the local page key) and returns the trailing payload. For a leaf entry that payload is
// the record's data definition; for a branch entry it is followed by the child page
// number (see branchChild).
func entryData(flags uint8, data []byte) ([]byte, error) {
	off := 0
	if flags&tagCommon != 0 {
		if len(data) < 2 {
			return nil, fmt.Errorf("ese: entry too short for common key size")
		}
		off += 2 // CommonPageKeySize
	}
	if len(data) < off+2 {
		return nil, fmt.Errorf("ese: entry too short for local key size")
	}
	localKeySize := int(binary.LittleEndian.Uint16(data[off : off+2]))
	off += 2 + localKeySize
	if off > len(data) {
		return nil, fmt.Errorf("ese: entry local key runs past end")
	}
	return data[off:], nil
}

// branchChild returns the child page number from a branch-page entry.
func branchChild(flags uint8, data []byte) (uint32, error) {
	payload, err := entryData(flags, data)
	if err != nil {
		return 0, err
	}
	if len(payload) < 4 {
		return 0, fmt.Errorf("ese: branch entry missing child page number")
	}
	return binary.LittleEndian.Uint32(payload[len(payload)-4:]), nil
}
