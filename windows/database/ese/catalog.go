package ese

import (
	"encoding/binary"
	"fmt"
)

// parseCatalog walks the catalog B-tree starting at pageNum, registering tables and
// their columns. Leaf pages hold catalog entries; branch pages point to child pages.
func (d *Database) parseCatalog(pageNum uint32, depth int) error {
	if depth > maxCatalogDepth {
		return fmt.Errorf("ese: catalog nesting too deep at page %d (possible cycle)", pageNum)
	}
	p, err := d.getPage(pageNum)
	if err != nil {
		return err
	}
	for i := 1; i < int(p.firstAvailablePageTag); i++ {
		flags, data, err := p.getTag(i)
		if err != nil {
			continue
		}
		if p.isLeaf() {
			payload, err := entryData(flags, data)
			if err != nil {
				continue
			}
			d.addCatalogEntry(payload)
		} else {
			child, err := branchChild(flags, data)
			if err != nil {
				continue
			}
			if err := d.parseCatalog(child, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

// addCatalogEntry parses one catalog record (a data definition) and registers the table,
// column, index, or long-value it describes. Layout (after the 4-byte data-definition
// header): FatherDataPageID(4), Type(2), Identifier(4), then type-specific fixed fields,
// with the item name in the variable-data area.
func (d *Database) addCatalogEntry(payload []byte) {
	if len(payload) < dataDefinitionSize+10 {
		return
	}
	lastVariable := payload[1]
	variableSizeOffset := binary.LittleEndian.Uint16(payload[2:4])

	fixed := payload[dataDefinitionSize:] // FatherDataPageID(4) Type(2) Identifier(4) ...
	entryType := binary.LittleEndian.Uint16(fixed[4:6])
	identifier := binary.LittleEndian.Uint32(fixed[6:10])
	name := parseItemName(payload, variableSizeOffset, lastVariable)

	switch entryType {
	case catalogTypeTable:
		if len(fixed) < 14 {
			return
		}
		t := &Table{
			Name:           name,
			fatherDataPage: binary.LittleEndian.Uint32(fixed[10:14]),
			columnByID:     make(map[uint32]*Column),
			columnByName:   make(map[string]*Column),
			db:             d,
		}
		if _, exists := d.tables[name]; !exists {
			d.tableOrder = append(d.tableOrder, name)
		}
		d.tables[name] = t
		d.currentTable = t

	case catalogTypeColumn:
		if d.currentTable == nil || len(fixed) < 26 {
			return
		}
		c := &Column{
			Name:       name,
			ID:         identifier,
			Type:       binary.LittleEndian.Uint32(fixed[10:14]),
			SpaceUsage: binary.LittleEndian.Uint32(fixed[14:18]),
			CodePage:   binary.LittleEndian.Uint32(fixed[22:26]),
		}
		t := d.currentTable
		if _, exists := t.columnByID[c.ID]; !exists {
			t.columns = append(t.columns, c)
		}
		t.columnByID[c.ID] = c
		t.columnByName[c.Name] = c

	case catalogTypeIndex, catalogTypeLongValue:
		// Not needed for row reading.
	}
}

// parseItemName extracts the item (table/column) name from the variable-data area of a
// catalog record: a numEntries-long array of uint16 cumulative lengths, then the values;
// the name is the first variable value.
func parseItemName(payload []byte, variableSizeOffset uint16, lastVariable byte) string {
	numEntries := int(lastVariable)
	if lastVariable > 127 {
		numEntries = int(lastVariable) - 127
	}
	vo := int(variableSizeOffset)
	if vo+2 > len(payload) {
		return ""
	}
	itemLen := int(binary.LittleEndian.Uint16(payload[vo : vo+2]))
	start := vo + 2*numEntries
	if start < 0 || start > len(payload) {
		return ""
	}
	end := start + itemLen
	if end > len(payload) {
		end = len(payload)
	}
	return string(payload[start:end])
}
