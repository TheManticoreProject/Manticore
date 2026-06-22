package ntds

import (
	"encoding/binary"
	"sort"
	"testing"
	"unicode/utf16"

	"github.com/TheManticoreProject/Manticore/windows/database/ese"
)

// This test builds a synthetic NTDS datatable (an ESE database) carrying a pekList and a
// user object whose unicodePwd/dBCSPwd were encrypted with impacket-generated vectors,
// then runs Dump end-to-end (ESE read -> attribute access -> PEK/hash decryption). The
// blobs come from the same Python (impacket + pycryptodome) generator as the crypto
// tests, so a correct result validates the whole offline pipeline.
const (
	dvBootKey  = "000102030405060708090a0b0c0d0e0f"
	dvPekList  = "0200000000000000aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa01eb46cbb54c926d566dc36e44e20d3946a3ad276548def414bf2828cd1559a6f5c526b2917ec2e28bc891bb52d9c16b467b20ae"
	dvEncNT    = "0000000000000000c1c1c1c1c1c1c1c1c1c1c1c1c1c1c1c15222e8fd0bc1751323e9fc9768131b24"
	dvEncLM    = "0000000000000000c2c2c2c2c2c2c2c2c2c2c2c2c2c2c2c235a006c017bb3f114adcda25a2f8256c"
	dvSID      = "010500000000000515000000010000000200000003000000000003e8" // RID 1000, big-endian (NTDS storage)
	dvExpectNT = "8846f7eaee8fb117ad06bdd830b7586c"
	dvExpectLM = "e52cac67419a9a224a3b108f3fa6cb6d"
	dvRID      = 1000
)

// --- minimal ESE assembler (all-tagged datatable) ---

const (
	edbSig       = 0x89ABCDEF
	edbPageSize  = 8192
	flagRootLeaf = 0x01 | 0x02
	catTypeTable = 1
	catTypeCol   = 2
	colTypeLong  = 4
	colTypeBin   = 9
	colTypeText  = 12
	cpUnicode    = 1200
	ddhSize      = 4
)

func u16(v int) []byte    { b := make([]byte, 2); binary.LittleEndian.PutUint16(b, uint16(v)); return b }
func u32(v uint32) []byte { b := make([]byte, 4); binary.LittleEndian.PutUint32(b, v); return b }
func utf16b(s string) []byte {
	u := utf16.Encode([]rune(s))
	b := make([]byte, len(u)*2)
	for i, c := range u {
		binary.LittleEndian.PutUint16(b[i*2:], c)
	}
	return b
}
func leafEntry(data []byte) []byte { return append(u16(0), data...) }

func catEntry(fixed []byte, name string) []byte {
	out := append([]byte{7, 128}, u16(ddhSize+len(fixed))...) // lastFixed,lastVar=128,varSizeOffset
	out = append(out, fixed...)
	out = append(out, u16(len(name))...)
	return append(out, []byte(name)...)
}
func catTable(objid, dataPage uint32, name string) []byte {
	f := append(u32(objid), u16(catTypeTable)...)
	f = append(f, u32(objid)...)
	f = append(f, u32(dataPage)...)
	f = append(f, u32(0)...)
	return catEntry(f, name)
}
func catColumn(objid, id, colType, codePage uint32, name string) []byte {
	f := append(u32(objid), u16(catTypeCol)...)
	f = append(f, u32(id)...)
	f = append(f, u32(colType)...)
	f = append(f, u32(0)...) // SpaceUsage
	f = append(f, u32(0)...) // ColumnFlags
	f = append(f, u32(codePage)...)
	return catEntry(f, name)
}

// taggedRecord builds an all-tagged data record (no fixed/variable columns).
func taggedRecord(cols map[uint32][]byte) []byte {
	ids := make([]int, 0, len(cols))
	for id := range cols {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	arraySize := len(ids) * 4
	var arr, vals []byte
	off := arraySize
	for _, id := range ids {
		v := cols[uint32(id)]
		arr = append(arr, u16(id)...)
		arr = append(arr, u16(off)...) // flags clear
		vals = append(vals, v...)
		off += len(v)
	}
	rec := append([]byte{0, 127}, u16(ddhSize)...) // lastFixed=0,lastVar=127,varSizeOffset=4
	rec = append(rec, arr...)
	return append(rec, vals...)
}

func buildPage(pageFlags, nextPage uint32, records [][]byte) []byte {
	page := make([]byte, edbPageSize)
	const common, headerLen = 8, 40
	binary.LittleEndian.PutUint32(page[common+12:], nextPage)
	binary.LittleEndian.PutUint16(page[common+26:], uint16(len(records)+1))
	binary.LittleEndian.PutUint32(page[common+28:], pageFlags)
	put := func(i, size, voff int) {
		pos := edbPageSize - 4*(i+1)
		binary.LittleEndian.PutUint16(page[pos:], uint16(size&0x1fff))
		binary.LittleEndian.PutUint16(page[pos+2:], uint16(voff&0x1fff))
	}
	put(0, 0, 0)
	o := headerLen
	for i, r := range records {
		copy(page[o:], r)
		put(i+1, len(r), o-headerLen)
		o += len(r)
	}
	return page
}

func buildNTDSDIT(t *testing.T) []byte {
	hdr := make([]byte, edbPageSize)
	binary.LittleEndian.PutUint32(hdr[4:], edbSig)
	binary.LittleEndian.PutUint32(hdr[8:], 0x620)
	binary.LittleEndian.PutUint32(hdr[232:], 0x11)
	binary.LittleEndian.PutUint32(hdr[236:], edbPageSize)

	catalog := buildPage(flagRootLeaf, 0, [][]byte{
		leafEntry(catTable(16, 5, "datatable")),
		leafEntry(catColumn(16, 256, colTypeBin, 0, AttPEKList)),
		leafEntry(catColumn(16, 257, colTypeText, cpUnicode, AttSAMAccountName)),
		leafEntry(catColumn(16, 258, colTypeBin, 0, AttObjectSid)),
		leafEntry(catColumn(16, 259, colTypeBin, 0, AttUnicodePwd)),
		leafEntry(catColumn(16, 260, colTypeBin, 0, AttDBCSPwd)),
		leafEntry(catColumn(16, 261, colTypeLong, 0, AttUserAccountControl)),
		leafEntry(catColumn(16, 262, colTypeLong, 0, AttSAMAccountType)),
	})
	data := buildPage(flagRootLeaf, 0, [][]byte{
		leafEntry(taggedRecord(map[uint32][]byte{256: hexBytes(dvPekList)})), // pek holder, no sAMAccountName
		leafEntry(taggedRecord(map[uint32][]byte{
			257: utf16b("Administrator"),
			258: hexBytes(dvSID),
			259: hexBytes(dvEncNT),
			260: hexBytes(dvEncLM),
			261: u32(0x200),      // userAccountControl: normal account (enabled)
			262: u32(0x30000000), // sAMAccountType: SAM_NORMAL_USER_ACCOUNT
		})),
	})

	out := append([]byte(nil), hdr...)
	for i := 0; i < 4; i++ {
		out = append(out, make([]byte, edbPageSize)...)
	}
	out = append(out, catalog...)
	return append(out, data...)
}

func TestDumpNTDS(t *testing.T) {
	db, err := ese.OpenBytes(buildNTDSDIT(t))
	if err != nil {
		t.Fatalf("ese.OpenBytes: %v", err)
	}
	defer db.Close()

	var accounts []Account
	if err := Dump(db, hexBytes(dvBootKey), func(a Account) error {
		accounts = append(accounts, a)
		return nil
	}); err != nil {
		t.Fatalf("Dump: %v", err)
	}

	if len(accounts) != 1 {
		t.Fatalf("got %d accounts, want 1 (pek-holder row must be skipped)", len(accounts))
	}
	a := accounts[0]
	if a.SAMAccountName != "Administrator" {
		t.Errorf("sAMAccountName = %q, want Administrator", a.SAMAccountName)
	}
	if a.RID != dvRID {
		t.Errorf("RID = %d, want %d", a.RID, dvRID)
	}
	if got := hexString(a.NTHash); got != dvExpectNT {
		t.Errorf("NT hash = %s, want %s", got, dvExpectNT)
	}
	if got := hexString(a.LMHash); got != dvExpectLM {
		t.Errorf("LM hash = %s, want %s", got, dvExpectLM)
	}
	if a.Disabled() {
		t.Error("account reported disabled, want enabled (uac=0x200)")
	}
	want := "Administrator:1000:" + dvExpectLM + ":" + dvExpectNT + ":::"
	if got := a.SecretsdumpLine(); got != want {
		t.Errorf("SecretsdumpLine = %q, want %q", got, want)
	}
}

func hexString(b []byte) string {
	const hexdigits = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, c := range b {
		out[i*2] = hexdigits[c>>4]
		out[i*2+1] = hexdigits[c&0xf]
	}
	return string(out)
}
