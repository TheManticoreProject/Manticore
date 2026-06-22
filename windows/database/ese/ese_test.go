package ese

import (
	"bytes"
	"encoding/binary"
	"flag"
	"os"
	"testing"
	"unicode/utf16"
)

var update = flag.Bool("update", false, "regenerate testdata golden files")

const edbPageSize = 8192

func u16le(v int) []byte    { b := make([]byte, 2); binary.LittleEndian.PutUint16(b, uint16(v)); return b }
func u32le(v uint32) []byte { b := make([]byte, 4); binary.LittleEndian.PutUint32(b, v); return b }

func utf16leBytes(s string) []byte {
	u := utf16.Encode([]rune(s))
	b := make([]byte, len(u)*2)
	for i, c := range u {
		binary.LittleEndian.PutUint16(b[i*2:], c)
	}
	return b
}

// leaf wraps a record's data definition in a leaf entry (local key size 0, no common key).
func leaf(entryData []byte) []byte {
	return append(u16le(0), entryData...)
}

// catalogTable builds a CATALOG_TYPE_TABLE data definition naming a table whose data
// B-tree root is dataPage.
func catalogTable(objid, dataPage uint32, name string) []byte {
	var fixed []byte
	fixed = append(fixed, u32le(objid)...)            // FatherDataPageID
	fixed = append(fixed, u16le(catalogTypeTable)...) // Type
	fixed = append(fixed, u32le(objid)...)            // Identifier
	fixed = append(fixed, u32le(dataPage)...)         // FatherDataPageNumber
	fixed = append(fixed, u32le(0)...)                // SpaceUsage
	return catalogEntry(fixed, name)
}

// catalogColumn builds a CATALOG_TYPE_COLUMN data definition.
func catalogColumn(objid, colID, colType, spaceUsage, codePage uint32, name string) []byte {
	var fixed []byte
	fixed = append(fixed, u32le(objid)...)             // FatherDataPageID
	fixed = append(fixed, u16le(catalogTypeColumn)...) // Type
	fixed = append(fixed, u32le(colID)...)             // Identifier
	fixed = append(fixed, u32le(colType)...)           // ColumnType
	fixed = append(fixed, u32le(spaceUsage)...)        // SpaceUsage
	fixed = append(fixed, u32le(0)...)                 // ColumnFlags
	fixed = append(fixed, u32le(codePage)...)          // CodePage
	return catalogEntry(fixed, name)
}

// catalogEntry assembles a data definition: a 4-byte header, the fixed fields, then the
// item name as the single variable column (id 128).
func catalogEntry(fixed []byte, name string) []byte {
	variableSizeOffset := dataDefinitionSize + len(fixed)
	ddh := []byte{7, 128} // LastFixedSize (cosmetic), LastVariableDataType=128
	ddh = append(ddh, u16le(variableSizeOffset)...)
	out := append([]byte(nil), ddh...)
	out = append(out, fixed...)
	out = append(out, u16le(len(name))...) // variable-size array: one entry = name length
	out = append(out, []byte(name)...)
	return out
}

// dataRecord builds a datatable row with the fixed/variable/tagged columns used by the
// test schema (col 1 fixed Long, col 128 variable LongText, cols 256/257 tagged).
func dataRecord(fixedNum uint32, varText string, tagBlob []byte, tagText string) []byte {
	varBytes := utf16leBytes(varText)
	tagTextBytes := utf16leBytes(tagText)

	ddh := []byte{1, 128}          // LastFixedSize=1, LastVariableDataType=128
	ddh = append(ddh, u16le(8)...) // VariableSizeOffset = 4 (ddh) + 4 (fixed) = 8

	out := append([]byte(nil), ddh...)
	out = append(out, u32le(fixedNum)...)      // fixed col 1
	out = append(out, u16le(len(varBytes))...) // variable-size array (1 entry)
	out = append(out, varBytes...)             // variable col 128 value

	// tagged index array (2 entries of id+offset, offsets relative to taggedBase) then values.
	off0 := 8 // values start right after the 8-byte array
	off1 := off0 + len(tagBlob)
	out = append(out, u16le(256)...)
	out = append(out, u16le(off0)...) // flags clear (no per-item flag byte)
	out = append(out, u16le(257)...)
	out = append(out, u16le(off1)...)
	out = append(out, tagBlob...)
	out = append(out, tagTextBytes...)
	return out
}

// buildPage lays out one 8192-byte page: a 40-byte header, the records, and the tag
// array at the page end (tag 0 reserved, tags 1..n point at the records).
func buildPage(pageFlags, nextPage uint32, records [][]byte) []byte {
	page := make([]byte, edbPageSize)
	const common = 8
	const headerLen = 40
	binary.LittleEndian.PutUint32(page[common+12:], nextPage)               // NextPageNumber
	binary.LittleEndian.PutUint16(page[common+26:], uint16(len(records)+1)) // FirstAvailablePageTag
	binary.LittleEndian.PutUint32(page[common+28:], pageFlags)              // PageFlags

	putTag := func(i, size, valueOffset int) {
		pos := edbPageSize - 4*(i+1)
		binary.LittleEndian.PutUint16(page[pos:], uint16(size&0x1fff))
		binary.LittleEndian.PutUint16(page[pos+2:], uint16(valueOffset&0x1fff))
	}
	putTag(0, 0, 0) // reserved page tag

	off := headerLen
	for i, r := range records {
		copy(page[off:], r)
		putTag(i+1, len(r), off-headerLen)
		off += len(r)
	}
	return page
}

// buildSampleEDB assembles a complete minimal ESE database: header page, four empty
// pages, the catalog (page 4), and the datatable data (page 5).
func buildSampleEDB() []byte {
	header := make([]byte, edbPageSize)
	binary.LittleEndian.PutUint32(header[4:], eseSignature)
	binary.LittleEndian.PutUint32(header[8:], 0x620)  // Version
	binary.LittleEndian.PutUint32(header[232:], 0x11) // FileFormatRevision
	binary.LittleEndian.PutUint32(header[236:], edbPageSize)

	catalog := buildPage(flagRoot|flagLeaf, 0, [][]byte{
		leaf(catalogTable(16, 5, "datatable")),
		leaf(catalogColumn(16, 1, JetColtypLong, 4, 0, "FixedNum")),
		leaf(catalogColumn(16, 128, JetColtypLongText, 0, codePageUnicode, "VarText")),
		leaf(catalogColumn(16, 256, JetColtypBinary, 0, 0, "TagBlob")),
		leaf(catalogColumn(16, 257, JetColtypLongText, 0, codePageUnicode, "TagText")),
	})
	data := buildPage(flagRoot|flagLeaf, 0, [][]byte{
		leaf(dataRecord(2000, "hello", []byte{0xDE, 0xAD, 0xBE, 0xEF}, "Administrator")),
		leaf(dataRecord(1107, "world", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, "Guest")),
	})

	out := append([]byte(nil), header...)
	for i := 0; i < 4; i++ {
		out = append(out, make([]byte, edbPageSize)...) // pages 0..3 (unused)
	}
	out = append(out, catalog...) // page 4
	out = append(out, data...)    // page 5
	return out
}

const edbGoldenPath = "testdata/sample.edb"

func TestESEGoldenFixture(t *testing.T) {
	want := buildSampleEDB()
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(edbGoldenPath, want, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d bytes to %s", len(want), edbGoldenPath)
	}
	got, err := os.ReadFile(edbGoldenPath)
	if err != nil {
		t.Fatalf("read golden (run with -update): %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s out of date; run: go test -run TestESEGoldenFixture -update", edbGoldenPath)
	}
}

func TestESEReadSample(t *testing.T) {
	db, err := OpenBytes(buildSampleEDB())
	if err != nil {
		t.Fatalf("OpenBytes: %v", err)
	}
	defer db.Close()

	if got := db.TableNames(); len(got) != 1 || got[0] != "datatable" {
		t.Fatalf("TableNames = %v, want [datatable]", got)
	}
	tbl, err := db.Table("datatable")
	if err != nil {
		t.Fatalf("Table: %v", err)
	}
	if len(tbl.Columns()) != 4 {
		t.Fatalf("got %d columns, want 4", len(tbl.Columns()))
	}

	cur, err := tbl.Rows()
	if err != nil {
		t.Fatalf("Rows: %v", err)
	}
	type want struct {
		fixed   uint32
		varText string
		blob    []byte
		tagText string
	}
	wants := []want{
		{2000, "hello", []byte{0xDE, 0xAD, 0xBE, 0xEF}, "Administrator"},
		{1107, "world", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, "Guest"},
	}
	n := 0
	for cur.Next() {
		if n >= len(wants) {
			t.Fatalf("more rows than expected")
		}
		row := cur.Row()
		w := wants[n]
		if v, _ := row.Uint32("FixedNum"); v != w.fixed {
			t.Errorf("row %d FixedNum = %d, want %d", n, v, w.fixed)
		}
		if v, _ := row.String("VarText"); v != w.varText {
			t.Errorf("row %d VarText = %q, want %q", n, v, w.varText)
		}
		if v, _ := row.Raw("TagBlob"); !bytes.Equal(v, w.blob) {
			t.Errorf("row %d TagBlob = % x, want % x", n, v, w.blob)
		}
		if v, _ := row.String("TagText"); v != w.tagText {
			t.Errorf("row %d TagText = %q, want %q", n, v, w.tagText)
		}
		n++
	}
	if err := cur.Err(); err != nil {
		t.Fatalf("cursor err: %v", err)
	}
	if n != 2 {
		t.Errorf("read %d rows, want 2", n)
	}
}

// --- long-value (LV) fixture ---

// catalogLongValue builds a CATALOG_TYPE_LONG_VALUE data definition pointing at the
// table's long-value B-tree root page.
func catalogLongValue(objid, lvPage uint32, name string) []byte {
	var fixed []byte
	fixed = append(fixed, u32le(objid)...) // FatherDataPageID (table objid)
	fixed = append(fixed, u16le(catalogTypeLongValue)...)
	fixed = append(fixed, u32le(objid+1)...) // Identifier
	fixed = append(fixed, u32le(lvPage)...)  // FatherDataPageNumber (LV tree root)
	fixed = append(fixed, u32le(0)...)       // SpaceUsage
	return catalogEntry(fixed, name)
}

// lvRefRecord builds a datatable row: a fixed Long column and a tagged column (256) that
// is a long-value reference (flag 0x04) carrying the 4-byte little-endian LID.
func lvRefRecord(fixedNum uint32, lidLE []byte) []byte {
	ddh := []byte{1, 127}          // LastFixedSize=1, LastVariableDataType=127 (no variable)
	ddh = append(ddh, u16le(8)...) // VariableSizeOffset = 4 (ddh) + 4 (fixed) = 8 -> taggedBase
	out := append([]byte(nil), ddh...)
	out = append(out, u32le(fixedNum)...) // fixed col 1
	// tagged array: one entry (id 256), offset 4 (after the 4-byte array), flag bit set.
	out = append(out, u16le(256)...)
	out = append(out, u16le(4|0x4000)...) // 0x4000 => per-item flag byte present
	out = append(out, 0x04)               // tagged flag: long value
	out = append(out, lidLE...)           // the long-value identifier
	return out
}

// lvLeafKV builds one long-value B-tree leaf entry (explicit local key, no common key).
func lvLeafKV(key, data []byte) []byte {
	out := append(u16le(len(key)), key...)
	return append(out, data...)
}

func bigBlob(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i % 256)
	}
	return b
}

// buildLVSampleEDB builds a database whose single datatable row stores a 600-byte value as
// a long value split across two segments (offsets 0 and 400) in the long-value tree.
func buildLVSampleEDB() ([]byte, []byte) {
	header := make([]byte, edbPageSize)
	binary.LittleEndian.PutUint32(header[4:], eseSignature)
	binary.LittleEndian.PutUint32(header[8:], 0x620)
	binary.LittleEndian.PutUint32(header[232:], 0x11)
	binary.LittleEndian.PutUint32(header[236:], edbPageSize)

	catalog := buildPage(flagRoot|flagLeaf, 0, [][]byte{
		leaf(catalogTable(16, 5, "datatable")),
		leaf(catalogColumn(16, 1, JetColtypLong, 4, 0, "FixedNum")),
		leaf(catalogColumn(16, 256, JetColtypLongBinary, 0, 0, "BigBlob")),
		leaf(catalogLongValue(16, 6, "LV")), // long-value tree at page 6
	})
	data := buildPage(flagRoot|flagLeaf, 0, [][]byte{
		leaf(lvRefRecord(2000, []byte{0x01, 0x00, 0x00, 0x00})), // LID 1
	})

	blob := bigBlob(600)
	pageKey := []byte{0x00, 0x00, 0x00, 0x01} // LID reversed
	segKey := func(off uint32) []byte {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, off)
		return append(append([]byte(nil), pageKey...), b...)
	}
	lv := buildPage(flagRoot|flagLeaf|flagLongValue, 0, [][]byte{
		lvLeafKV(pageKey, u32le(uint32(len(blob)))), // header: total size
		lvLeafKV(segKey(0), blob[:400]),
		lvLeafKV(segKey(400), blob[400:]),
	})

	out := append([]byte(nil), header...)
	for i := 0; i < 4; i++ {
		out = append(out, make([]byte, edbPageSize)...)
	}
	out = append(out, catalog...) // page 4
	out = append(out, data...)    // page 5
	out = append(out, lv...)      // page 6
	return out, blob
}

func TestESELongValue(t *testing.T) {
	image, want := buildLVSampleEDB()
	db, err := OpenBytes(image)
	if err != nil {
		t.Fatalf("OpenBytes: %v", err)
	}
	defer db.Close()
	tbl, _ := db.Table("datatable")
	cur, err := tbl.Rows()
	if err != nil {
		t.Fatalf("Rows: %v", err)
	}
	if !cur.Next() {
		t.Fatalf("no rows (err=%v)", cur.Err())
	}
	got, ok := cur.Row().Raw("BigBlob")
	if !ok {
		t.Fatal("BigBlob not present")
	}
	if !bytes.Equal(got, want) {
		t.Errorf("BigBlob reassembled to %d bytes, want %d (equal=%v)", len(got), len(want), bytes.Equal(got, want))
	}
}
