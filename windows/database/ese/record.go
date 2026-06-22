package ese

import (
	"encoding/binary"
	"unicode/utf16"
)

// extendedTags reports whether the database uses the Windows-7+ large-page tag encoding
// (tag flags always present, value-embedded page flags).
func (d *Database) extendedTags() bool {
	return d.version == 0x620 && d.revision >= 0x11 && d.pageSize > 8192
}

// Row is a decoded data record: column values by column ID, as raw on-disk bytes.
type Row struct {
	table  *Table
	values map[uint32][]byte
}

// RawByID returns the raw bytes of the column with the given ID.
func (r *Row) RawByID(id uint32) ([]byte, bool) {
	v, ok := r.values[id]
	return v, ok && v != nil
}

// Raw returns the raw bytes of the named column.
func (r *Row) Raw(name string) ([]byte, bool) {
	c, ok := r.table.columnByName[name]
	if !ok {
		return nil, false
	}
	return r.RawByID(c.ID)
}

// Has reports whether the named column is present (non-empty) in this row.
func (r *Row) Has(name string) bool {
	_, ok := r.Raw(name)
	return ok
}

// String decodes the named text column to a Go string, honouring its code page
// (UTF-16LE for Unicode columns, otherwise byte-for-byte).
func (r *Row) String(name string) (string, bool) {
	c, ok := r.table.columnByName[name]
	if !ok {
		return "", false
	}
	v, ok := r.RawByID(c.ID)
	if !ok {
		return "", false
	}
	if c.CodePage == codePageUnicode {
		u := make([]uint16, 0, len(v)/2)
		for i := 0; i+1 < len(v); i += 2 {
			u = append(u, binary.LittleEndian.Uint16(v[i:i+2]))
		}
		return string(utf16.Decode(u)), true
	}
	return string(v), true
}

// Uint32 returns the named column decoded as a little-endian uint32.
func (r *Row) Uint32(name string) (uint32, bool) {
	v, ok := r.Raw(name)
	if !ok || len(v) < 4 {
		return 0, false
	}
	return binary.LittleEndian.Uint32(v[:4]), true
}

// Int64 returns the named column decoded as a little-endian int64.
func (r *Row) Int64(name string) (int64, bool) {
	v, ok := r.Raw(name)
	if !ok || len(v) < 8 {
		return 0, false
	}
	return int64(binary.LittleEndian.Uint64(v[:8])), true
}

type taggedItem struct {
	id           uint32
	offset       int
	flagsPresent bool
}

// decodeRecord decodes a leaf-entry data record into a Row, following the ESE record
// layout: a data-definition header, then fixed-size columns (ID ≤ LastFixedSize),
// variable-size columns (127 < ID ≤ LastVariableDataType) addressed via a length array,
// then tagged columns (ID > 255) addressed via a tag array. Columns are processed in
// ascending-ID (catalog) order, which the running offsets rely on.
func decodeRecord(t *Table, payload []byte) (*Row, error) {
	values := make(map[uint32][]byte)
	if len(payload) < dataDefinitionSize {
		return &Row{table: t, values: values}, nil
	}
	lastFixed := uint32(payload[0])
	lastVariable := payload[1]
	variableSizeOffset := int(binary.LittleEndian.Uint16(payload[2:4]))

	varCount := 0
	if lastVariable > 127 {
		varCount = int(lastVariable) - 127
	}
	variableDataBytesProcessed := varCount * 2
	prevItemLen := 0
	fixedSizeOffset := dataDefinitionSize

	var tagged []taggedItem
	taggedParsed := false
	taggedBase := 0

	for _, col := range t.columns {
		id := col.ID
		switch {
		case id <= lastFixed:
			end := fixedSizeOffset + int(col.SpaceUsage)
			if fixedSizeOffset >= 0 && end <= len(payload) {
				values[id] = append([]byte(nil), payload[fixedSizeOffset:end]...)
			}
			fixedSizeOffset = end

		case id > 127 && id <= uint32(lastVariable):
			idx := int(id) - 127 - 1
			p := variableSizeOffset + idx*2
			if p+2 > len(payload) {
				continue
			}
			itemLen := int(binary.LittleEndian.Uint16(payload[p : p+2]))
			if itemLen&0x8000 != 0 {
				itemLen = prevItemLen // empty item
			} else {
				start := variableSizeOffset + variableDataBytesProcessed
				vlen := itemLen - prevItemLen
				if vlen >= 0 && start >= 0 && start+vlen <= len(payload) {
					values[id] = append([]byte(nil), payload[start:start+vlen]...)
				}
				variableDataBytesProcessed += vlen
			}
			prevItemLen = itemLen

		case id > 255:
			if !taggedParsed {
				taggedBase = variableSizeOffset + variableDataBytesProcessed
				tagged = parseTaggedItems(payload, taggedBase, t.db.extendedTags())
				taggedParsed = true
			}
			if v, isLV, ok := lookupTagged(payload, taggedBase, tagged, id); ok {
				if isLV {
					// The value is a long-value identifier; reassemble it from the
					// long-value tree, falling back to the raw reference on failure.
					if resolved, err := t.resolveLongValue(v); err == nil {
						values[id] = resolved
					} else {
						values[id] = v
					}
				} else {
					values[id] = v
				}
			}
		}
	}
	return &Row{table: t, values: values}, nil
}

// parseTaggedItems reads the tagged-data index array that begins at base: a sequence of
// (2-byte identifier, 2-byte offset/flags) entries terminated where the first value
// begins.
func parseTaggedItems(payload []byte, base int, extended bool) []taggedItem {
	if base < 0 || base+4 > len(payload) {
		return nil
	}
	firstOffsetTag := int(binary.LittleEndian.Uint16(payload[base+2:base+4])&0x3fff) + base
	if firstOffsetTag > len(payload) {
		firstOffsetTag = len(payload)
	}
	var items []taggedItem
	index := base
	for index+4 <= len(payload) {
		id := binary.LittleEndian.Uint16(payload[index : index+2])
		w := binary.LittleEndian.Uint16(payload[index+2 : index+4])
		offset := int(w & 0x3fff)
		flagsPresent := extended || (w&0x4000 != 0)
		items = append(items, taggedItem{id: uint32(id), offset: offset, flagsPresent: flagsPresent})
		index += 4
		if index >= firstOffsetTag {
			break
		}
	}
	return items
}

// lookupTagged returns the value bytes of tagged column id, computing its length from the
// next tagged entry (or end of record for the last), and skipping a leading flag byte when
// present. It also reports whether the value is a long-value reference (flag 0x04), in
// which case the returned bytes are the long-value identifier. Compressed inline values
// (flag 0x02) are unsupported (reported absent).
func lookupTagged(payload []byte, base int, items []taggedItem, id uint32) (value []byte, isLongValue bool, ok bool) {
	for k, it := range items {
		if it.id != id {
			continue
		}
		offsetItem := base + it.offset
		var size int
		if k+1 < len(items) {
			size = items[k+1].offset - it.offset
		} else {
			size = len(payload) - offsetItem
		}
		if it.flagsPresent {
			if offsetItem >= len(payload) {
				return nil, false, false
			}
			flag := payload[offsetItem]
			offsetItem++
			size--
			if flag&taggedCompressed != 0 {
				return nil, false, false // compressed tagged data not supported
			}
			isLongValue = flag&taggedLongValue != 0
		}
		if size < 0 {
			size = 0
		}
		if offsetItem < 0 || offsetItem > len(payload) {
			return nil, false, false
		}
		end := offsetItem + size
		if end > len(payload) {
			end = len(payload)
		}
		return append([]byte(nil), payload[offsetItem:end]...), isLongValue, true
	}
	return nil, false, false
}
