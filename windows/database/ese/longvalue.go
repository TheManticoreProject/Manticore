package ese

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// lvEntry is one reconstructed leaf entry of a long-value B-tree: its full key and value.
type lvEntry struct {
	key  []byte
	data []byte
}

// parseEntry splits a leaf/branch entry into its full key and trailing payload. When the
// entry carries a common-key-size prefix (TAG_COMMON), the page's common key (its tag 0
// value) supplies the shared key prefix.
func parseEntry(flags uint8, data, commonKey []byte) (key, payload []byte, err error) {
	off := 0
	commonSize := 0
	if flags&tagCommon != 0 {
		if len(data) < 2 {
			return nil, nil, fmt.Errorf("ese: entry too short for common key size")
		}
		commonSize = int(binary.LittleEndian.Uint16(data[0:2]))
		off += 2
	}
	if len(data) < off+2 {
		return nil, nil, fmt.Errorf("ese: entry too short for local key size")
	}
	localSize := int(binary.LittleEndian.Uint16(data[off : off+2]))
	off += 2
	if off+localSize > len(data) {
		return nil, nil, fmt.Errorf("ese: entry local key runs past end")
	}
	localKey := data[off : off+localSize]
	off += localSize
	if commonSize > len(commonKey) {
		commonSize = len(commonKey)
	}
	key = append(append([]byte(nil), commonKey[:commonSize]...), localKey...)
	return key, data[off:], nil
}

// loadLongValueLeaves walks the table's long-value B-tree once and caches every leaf
// entry (key + data) for subsequent reassembly lookups.
func (t *Table) loadLongValueLeaves() {
	if t.lvLoaded {
		return
	}
	t.lvLoaded = true
	if t.longValuePage == 0 || t.longValuePage == nullCellPage {
		return
	}
	t.db.walkLongValueTree(t.longValuePage, 0, &t.lvLeaves)
}

// nullCellPage is the sentinel for "no page".
const nullCellPage = 0xFFFFFFFF

// walkLongValueTree recursively collects the leaf entries of a B-tree rooted at pageNum.
func (d *Database) walkLongValueTree(pageNum uint32, depth int, out *[]lvEntry) {
	if depth > maxCatalogDepth {
		return
	}
	p, err := d.getPage(pageNum)
	if err != nil {
		return
	}
	var commonKey []byte
	if p.firstAvailablePageTag > 0 {
		if _, v, err := p.getTag(0); err == nil {
			commonKey = v // page external header = common page key
		}
	}
	for i := 1; i < int(p.firstAvailablePageTag); i++ {
		flags, data, err := p.getTag(i)
		if err != nil {
			continue
		}
		if p.isLeaf() {
			key, payload, err := parseEntry(flags, data, commonKey)
			if err != nil {
				continue
			}
			*out = append(*out, lvEntry{key: key, data: payload})
		} else {
			if child, err := branchChild(flags, data); err == nil {
				d.walkLongValueTree(child, depth+1, out)
			}
		}
	}
}

// resolveLongValue reassembles the long value referenced by lid (the in-record reference
// bytes) from table t's long-value tree. The LV-tree page key is the LID reversed; data
// segments are keyed by that page key followed by a 4-byte big-endian segment offset, and
// are concatenated in offset order.
func (t *Table) resolveLongValue(lid []byte) ([]byte, error) {
	t.loadLongValueLeaves()
	if len(t.lvLeaves) == 0 {
		return nil, fmt.Errorf("ese: no long-value data for id % x", lid)
	}
	prefix := make([]byte, len(lid))
	for i := range lid {
		prefix[i] = lid[len(lid)-1-i] // reverse the LID -> page key
	}

	type seg struct {
		offset int
		data   []byte
	}
	var segs []seg
	end := 0
	for _, e := range t.lvLeaves {
		if len(e.key) != len(prefix)+4 || !bytes.Equal(e.key[:len(prefix)], prefix) {
			continue // header entry (key==prefix) or a different LID
		}
		offset := int(binary.BigEndian.Uint32(e.key[len(prefix):]))
		data := e.data
		if len(data) > 0 && data[0] == 0x18 {
			return nil, fmt.Errorf("ese: compressed long value (LZXPRESS) not supported")
		}
		segs = append(segs, seg{offset: offset, data: data})
		if offset+len(data) > end {
			end = offset + len(data)
		}
	}
	if len(segs) == 0 {
		return nil, fmt.Errorf("ese: long value % x not found", lid)
	}
	out := make([]byte, end)
	for _, s := range segs {
		copy(out[s.offset:], s.data)
	}
	return out, nil
}
