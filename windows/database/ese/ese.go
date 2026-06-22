// Package ese implements a read-only parser for Extensible Storage Engine (ESE / JET
// Blue) database files (EDB), the on-disk format used by Active Directory's NTDS.dit
// (as well as Windows Search, Exchange, etc.).
//
// It exposes enough of the format to enumerate tables and iterate rows with typed
// column access — the subset needed to read NTDS.dit's datatable and link_table. The
// implementation follows the behaviour of impacket's ese.py and the libesedb format
// documentation.
//
// References:
//   - libesedb "Extensible Storage Engine (ESE) Database File (EDB) format"
//     https://github.com/libyal/libesedb
//   - impacket ese.py (ESENT_DB)
package ese

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	eseSignature       = 0x89ABCDEF
	catalogPageNumber  = 4
	dataDefinitionSize = 4  // ESENT_DATA_DEFINITION_HEADER: LastFixedSize(1)+LastVariableDataType(1)+VariableSizeOffset(2)
	maxCatalogDepth    = 32 // guard against a cyclic catalog B-tree

	// Page flags ([MS] FLAGS_*).
	flagRoot      = 0x01
	flagLeaf      = 0x02
	flagSpaceTree = 0x20
	flagIndex     = 0x40
	flagLongValue = 0x80

	// Tag flags (the 3 high bits of a tag, or per-page-flag bits).
	tagCommon = 0x04

	// Catalog entry types.
	catalogTypeTable     = 1
	catalogTypeColumn    = 2
	catalogTypeIndex     = 3
	catalogTypeLongValue = 4

	// Tagged data type flags.
	taggedCompressed = 0x02
	taggedLongValue  = 0x04 // value is a long-value identifier (data lives in the LV tree)
	taggedMultiValue = 0x08
)

// JET column types (subset).
const (
	JetColtypNil           = 0
	JetColtypBit           = 1
	JetColtypUnsignedByte  = 2
	JetColtypShort         = 3
	JetColtypLong          = 4
	JetColtypCurrency      = 5
	JetColtypIEEESingle    = 6
	JetColtypIEEEDouble    = 7
	JetColtypDateTime      = 8
	JetColtypBinary        = 9
	JetColtypText          = 10
	JetColtypLongBinary    = 11
	JetColtypLongText      = 12
	JetColtypUnsignedLong  = 14
	JetColtypLongLong      = 15
	JetColtypGUID          = 16
	JetColtypUnsignedShort = 17
)

// Code pages used by text columns.
const (
	codePageUnicode = 1200
	codePageASCII   = 20127
	codePageWestern = 1252
)

// Database is an opened ESE database.
type Database struct {
	r            io.ReaderAt
	closer       io.Closer
	pageSize     int
	version      uint32
	revision     uint32
	tables       map[string]*Table
	tableOrder   []string
	currentTable *Table // most-recently-seen table while building the catalog
}

// Column describes one column of a table (from the catalog).
type Column struct {
	Name       string
	ID         uint32
	Type       uint32
	SpaceUsage uint32
	CodePage   uint32
}

// Table describes one table and its columns (from the catalog).
type Table struct {
	Name           string
	fatherDataPage uint32
	longValuePage  uint32 // root of the table's long-value B-tree (0 if none)
	columns        []*Column
	columnByID     map[uint32]*Column
	columnByName   map[string]*Column
	db             *Database

	lvLeaves []lvEntry // cached leaf entries of the long-value tree (lazy)
	lvLoaded bool
}

// Columns returns the table's columns in catalog order.
func (t *Table) Columns() []*Column { return t.columns }

// Open opens and parses the ESE database at path. Call Close when done.
func Open(path string) (*Database, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ese: open %s: %w", path, err)
	}
	db, err := newDatabase(f, f)
	if err != nil {
		f.Close()
		return nil, err
	}
	return db, nil
}

// OpenBytes parses an ESE database from an in-memory image.
func OpenBytes(data []byte) (*Database, error) {
	return newDatabase(bytes.NewReader(data), nil)
}

// Close releases the underlying file (no-op for OpenBytes).
func (d *Database) Close() error {
	if d.closer != nil {
		return d.closer.Close()
	}
	return nil
}

// PageSize returns the database page size in bytes.
func (d *Database) PageSize() int { return d.pageSize }

// TableNames returns the table names in catalog order.
func (d *Database) TableNames() []string { return append([]string(nil), d.tableOrder...) }

// Table returns the named table, or an error if it does not exist.
func (d *Database) Table(name string) (*Table, error) {
	t, ok := d.tables[name]
	if !ok {
		return nil, fmt.Errorf("ese: table %q not found", name)
	}
	return t, nil
}

func newDatabase(r io.ReaderAt, closer io.Closer) (*Database, error) {
	// The database header lives in the first 4096 bytes; read enough to learn the page
	// size, then bootstrap.
	hdr := make([]byte, 4096)
	if _, err := r.ReadAt(hdr, 0); err != nil && err != io.EOF {
		return nil, fmt.Errorf("ese: read header: %w", err)
	}
	if binary.LittleEndian.Uint32(hdr[4:8]) != eseSignature {
		return nil, fmt.Errorf("ese: bad signature 0x%08X (expected 0x%08X)",
			binary.LittleEndian.Uint32(hdr[4:8]), eseSignature)
	}
	d := &Database{
		r:        r,
		closer:   closer,
		version:  binary.LittleEndian.Uint32(hdr[8:12]),
		revision: binary.LittleEndian.Uint32(hdr[232:236]),
		pageSize: int(binary.LittleEndian.Uint32(hdr[236:240])),
		tables:   make(map[string]*Table),
	}
	if d.pageSize == 0 || d.pageSize%2 != 0 {
		return nil, fmt.Errorf("ese: invalid page size %d", d.pageSize)
	}
	if err := d.parseCatalog(catalogPageNumber, 0); err != nil {
		return nil, err
	}
	return d, nil
}

// getPage reads page pageNum (0-based; the header is conceptually page -1, so the byte
// offset is (pageNum+1)*pageSize).
func (d *Database) getPage(pageNum uint32) (*page, error) {
	buf := make([]byte, d.pageSize)
	off := int64(pageNum+1) * int64(d.pageSize)
	if _, err := d.r.ReadAt(buf, off); err != nil && err != io.EOF {
		return nil, fmt.Errorf("ese: read page %d: %w", pageNum, err)
	}
	return d.newPage(buf)
}
