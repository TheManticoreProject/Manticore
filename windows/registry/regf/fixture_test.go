package regf

import (
	"encoding/binary"
	"unicode/utf16"
)

// This file builds a small but complete, structurally valid REGF hive image entirely in
// memory, used both as the source for the committed golden fixture (testdata/sample.hive,
// see TestGoldenFixture) and indirectly by the navigation tests through that file.
//
// The hive it produces:
//
//	ROOT                       (root key node, UTF-16LE name, class data "classdata", SK record)
//	  └─ Sub                   (compressed-name subkey, shares the SK record)
//	       ├─ DwordVal = 0x12345678        REG_DWORD, inline
//	       ├─ StrVal   = "hello"           REG_SZ, external data cell
//	       └─ BigVal   = 20000 bytes       REG_BINARY, big-data (db) record, 2 segments
//
// Cells are framed with the standard 4-byte size prefix (negative = allocated) and
// 8-byte aligned, inside a single hive bin. Offsets are computed as cells are appended, so
// the layout never has to be hand-counted.

const (
	fixtureBinSize = 32768 // multiple of 4096, large enough for the big-data segments
	bigValSize     = 20000 // > bigDataThreshold, so BigVal needs a db record
)

// bigValBytes is the deterministic content of the BigVal value.
func bigValBytes() []byte {
	b := make([]byte, bigValSize)
	for i := range b {
		b[i] = byte(i % 251)
	}
	return b
}

// utf16le encodes s as UTF-16LE bytes (no terminator added).
func utf16le(s string) []byte {
	u := utf16.Encode([]rune(s))
	b := make([]byte, len(u)*2)
	for i, c := range u {
		binary.LittleEndian.PutUint16(b[i*2:], c)
	}
	return b
}

// hiveAsm assembles hive-bins data: a 32-byte HBIN header followed by size-prefixed cells.
type hiveAsm struct {
	bins []byte
}

func newHiveAsm() *hiveAsm {
	a := &hiveAsm{}
	hb := &HiveBin{Signature: hbinSignature, Offset: 0, Size: fixtureBinSize}
	hdr, _ := hb.Marshal()
	a.bins = append(a.bins, hdr...)
	return a
}

// addCell appends content as one allocated cell and returns the cell's offset relative to
// the start of hive-bins data (i.e. the offset stored in NK/VK/list records).
func (a *hiveAsm) addCell(content []byte) uint32 {
	off := uint32(len(a.bins))
	size := 4 + len(content)
	if r := size % 8; r != 0 {
		size += 8 - r
	}
	cell := make([]byte, size)
	binary.LittleEndian.PutUint32(cell[0:4], uint32(-int32(size))) // negative => allocated
	copy(cell[4:], content)
	a.bins = append(a.bins, cell...)
	return off
}

// buildSampleHive returns the bytes of the in-memory sample hive described above.
func buildSampleHive() []byte {
	a := newHiveAsm()

	// The root key node must be the first cell of the first hive bin: that is how real
	// hives are laid out and how other tools (and our RootKey via RootCellOffset) find it.
	// It references cells that don't exist yet (the subkey list and the class data), so we
	// emit it first with placeholder offsets and back-patch them once those cells are
	// added.
	rootName := utf16le("ROOT")
	root := &KeyNode{
		Signature:        nkSignature,
		Flags:            KeyHiveEntry, // UTF-16LE name (no compressed-name flag)
		NumberOfSubKeys:  1,
		ValuesListOffset: nullCellOffset,
		SecurityOffset:   nullCellOffset, // back-patched to skOff
		ClassNameLength:  uint16(len("classdata")),
		Parent:           nullCellOffset,
		KeyNameLength:    uint16(len(rootName)),
		KeyNameRaw:       rootName,
	}
	rootb, _ := root.Marshal()
	rootOff := a.addCell(rootb)
	rootContent := int(rootOff) + 4 // start of the NK record within a.bins

	classOff := a.addCell([]byte("classdata"))

	// SK record with a minimal valid self-relative SECURITY_DESCRIPTOR (revision 1,
	// SE_SELF_RELATIVE control, no owner/group/SACL/DACL). Shared by ROOT and Sub.
	sd := make([]byte, 20)
	sd[0] = 1                                      // Revision
	binary.LittleEndian.PutUint16(sd[2:4], 0x8000) // Control: SE_SELF_RELATIVE
	sk := &SecurityKey{Signature: skSignature, ReferenceCount: 2, SecurityDescriptorSize: uint32(len(sd)), SecurityDescriptor: sd}
	skb, _ := sk.Marshal()
	skOff := a.addCell(skb)

	szBytes := utf16le("hello\x00") // value data with NUL terminator: 12 bytes
	szOff := a.addCell(szBytes)

	vkD := &KeyValue{
		Signature:  vkSignature,
		NameLength: uint16(len("DwordVal")),
		DataSize:   0x80000000 | 4, // inline flag | 4 bytes
		DataOffset: 0x12345678,     // the inline data itself
		DataType:   RegDword,
		Flags:      ValueCompName,
		NameRaw:    []byte("DwordVal"),
	}
	vkDb, _ := vkD.Marshal()
	vkDOff := a.addCell(vkDb)

	vkS := &KeyValue{
		Signature:  vkSignature,
		NameLength: uint16(len("StrVal")),
		DataSize:   uint32(len(szBytes)),
		DataOffset: szOff,
		DataType:   RegSz,
		Flags:      ValueCompName,
		NameRaw:    []byte("StrVal"),
	}
	vkSb, _ := vkS.Marshal()
	vkSOff := a.addCell(vkSb)

	// Big-data value: 20000 bytes split into two segments (16344 + 3656), a segment-offset
	// list, and a db record. The VK points at the db record.
	big := bigValBytes()
	seg0Off := a.addCell(big[:bigDataThreshold])
	seg1Off := a.addCell(big[bigDataThreshold:])
	segList := make([]byte, 8)
	binary.LittleEndian.PutUint32(segList[0:4], seg0Off)
	binary.LittleEndian.PutUint32(segList[4:8], seg1Off)
	segListOff := a.addCell(segList)
	db := &BigData{Signature: dbSignature, NumberOfSegments: 2, SegmentsListOffset: segListOff}
	dbb, _ := db.Marshal()
	dbOff := a.addCell(dbb)

	vkB := &KeyValue{
		Signature:  vkSignature,
		NameLength: uint16(len("BigVal")),
		DataSize:   bigValSize, // not inline; > bigDataThreshold
		DataOffset: dbOff,
		DataType:   RegBinary,
		Flags:      ValueCompName,
		NameRaw:    []byte("BigVal"),
	}
	vkBb, _ := vkB.Marshal()
	vkBOff := a.addCell(vkBb)

	valueList := make([]byte, 12)
	binary.LittleEndian.PutUint32(valueList[0:4], vkDOff)
	binary.LittleEndian.PutUint32(valueList[4:8], vkSOff)
	binary.LittleEndian.PutUint32(valueList[8:12], vkBOff)
	valueListOff := a.addCell(valueList)

	sub := &KeyNode{
		Signature:         nkSignature,
		Flags:             KeyCompName,
		NumberOfValues:    3,
		ValuesListOffset:  valueListOff,
		SubKeysListOffset: nullCellOffset,
		ClassNameOffset:   nullCellOffset,
		SecurityOffset:    skOff,
		KeyNameLength:     uint16(len("Sub")),
		KeyNameRaw:        []byte("Sub"),
	}
	subb, _ := sub.Marshal()
	subOff := a.addCell(subb)

	lh := &SubKeyList{Signature: lhSig, NumberOfElements: 1, Elements: make([]byte, 8)}
	binary.LittleEndian.PutUint32(lh.Elements[0:4], subOff) // bytes [4:8] are the name hash, ignored
	lhb, _ := lh.Marshal()
	lhOff := a.addCell(lhb)

	// Back-patch the root NK's now-known offsets (SubKeysListOffset at NK+28,
	// SecurityOffset at NK+44, ClassNameOffset at NK+48).
	binary.LittleEndian.PutUint32(a.bins[rootContent+28:rootContent+32], lhOff)
	binary.LittleEndian.PutUint32(a.bins[rootContent+44:rootContent+48], skOff)
	binary.LittleEndian.PutUint32(a.bins[rootContent+48:rootContent+52], classOff)

	if len(a.bins) < fixtureBinSize {
		a.bins = append(a.bins, make([]byte, fixtureBinSize-len(a.bins))...)
	}

	bb := &BaseBlock{
		Signature:               regfSignature,
		PrimarySequenceNumber:   1,
		SecondarySequenceNumber: 1,
		MajorVersion:            1,
		MinorVersion:            3,
		FileType:                0, // primary file
		FileFormat:              1, // direct memory load
		RootCellOffset:          rootOff,
		HiveBinsDataSize:        uint32(len(a.bins)),
		ClusteringFactor:        1,
	}
	bbb, _ := bb.Marshal()

	// XOR-32 checksum over the first 508 bytes (parser does not enforce it, but a real
	// hive carries a valid one).
	var cs uint32
	for i := 0; i < 508; i += 4 {
		cs ^= binary.LittleEndian.Uint32(bbb[i : i+4])
	}
	switch cs {
	case 0:
		cs = 1
	case 0xFFFFFFFF:
		cs = 0xFFFFFFFE
	}
	binary.LittleEndian.PutUint32(bbb[508:512], cs)

	return append(bbb, a.bins...)
}
