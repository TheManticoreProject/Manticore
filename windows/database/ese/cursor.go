package ese

// Cursor iterates the rows of a table's data B-tree, left to right across leaf pages.
// The zero value is not usable; obtain one from Table.Rows.
type Cursor struct {
	table *Table
	page  *page
	tag   int
	row   *Row
	err   error
}

// Rows returns a cursor positioned before the first row of the table. It descends the
// data B-tree to the left-most leaf page.
func (t *Table) Rows() (*Cursor, error) {
	pageNum := t.fatherDataPage
	var p *page
	for depth := 0; ; depth++ {
		if depth > maxCatalogDepth {
			break // cycle guard
		}
		pp, err := t.db.getPage(pageNum)
		if err != nil {
			return nil, err
		}
		p = pp
		if p.firstAvailablePageTag <= 1 {
			break // empty page: no records
		}
		if p.isLeaf() {
			break
		}
		flags, data, err := p.getTag(1)
		if err != nil {
			break
		}
		child, err := branchChild(flags, data)
		if err != nil {
			break
		}
		pageNum = child
	}
	return &Cursor{table: t, page: p, tag: 0}, nil
}

// Next advances to the next row, returning false at the end (check Err afterwards).
func (c *Cursor) Next() bool {
	for {
		c.tag++
		if c.tag >= int(c.page.firstAvailablePageTag) {
			if c.page.nextPageNumber == 0 {
				return false
			}
			np, err := c.table.db.getPage(c.page.nextPageNumber)
			if err != nil {
				c.err = err
				return false
			}
			c.page = np
			c.tag = 0
			continue
		}
		if !c.page.isLeaf() {
			continue
		}
		// Skip non-data leaf pages (space tree / index / long value).
		if c.page.pageFlags&(flagSpaceTree|flagIndex|flagLongValue) != 0 {
			continue
		}
		flags, data, err := c.page.getTag(c.tag)
		if err != nil {
			continue
		}
		payload, err := entryData(flags, data)
		if err != nil {
			continue
		}
		row, err := decodeRecord(c.table, payload)
		if err != nil {
			continue
		}
		c.row = row
		return true
	}
}

// Row returns the current row (valid after Next returns true).
func (c *Cursor) Row() *Row { return c.row }

// Err returns the first error encountered during iteration, if any.
func (c *Cursor) Err() error { return c.err }
