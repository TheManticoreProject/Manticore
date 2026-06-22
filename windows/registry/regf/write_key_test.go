package regf

import (
	"encoding/binary"
	"reflect"
	"testing"
)

// buildKeyHive builds a minimal writable hive whose root carries a real SK record (so
// created subkeys can inherit it): a single 4096-byte bin with root "ROOT" and an SK.
func buildKeyHive() []byte {
	a := &hiveAsm{}
	hb := &HiveBin{Signature: hbinSignature, Size: 4096}
	hdr, _ := hb.Marshal()
	a.bins = append(a.bins, hdr...)

	rootName := utf16le("ROOT")
	root := &KeyNode{
		Signature:         nkSignature,
		Flags:             KeyHiveEntry,
		SubKeysListOffset: nullCellOffset,
		ValuesListOffset:  nullCellOffset,
		ClassNameOffset:   nullCellOffset,
		SecurityOffset:    nullCellOffset, // back-patched
		Parent:            nullCellOffset,
		KeyNameLength:     uint16(len(rootName)),
		KeyNameRaw:        rootName,
	}
	rootb, _ := root.Marshal()
	rootOff := a.addCell(rootb)
	rootContent := int(rootOff) + 4

	sd := sampleSecurityDescriptorBytes()
	sk := &SecurityKey{Signature: skSignature, ReferenceCount: 1, SecurityDescriptorSize: uint32(len(sd)), SecurityDescriptor: sd}
	skb, _ := sk.Marshal()
	skOff := a.addCell(skb)
	binary.LittleEndian.PutUint32(a.bins[rootContent+44:rootContent+48], skOff) // SecurityOffset at NK+44

	if len(a.bins) < 4096 {
		a.bins = append(a.bins, make([]byte, 4096-len(a.bins))...)
	}

	bb := &BaseBlock{
		Signature:               regfSignature,
		PrimarySequenceNumber:   1,
		SecondarySequenceNumber: 1,
		MajorVersion:            1,
		MinorVersion:            6,
		FileFormat:              1,
		RootCellOffset:          rootOff,
		HiveBinsDataSize:        uint32(len(a.bins)),
		ClusteringFactor:        1,
	}
	bbb, _ := bb.Marshal()
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

func openKeyHive(t *testing.T) *Hive {
	t.Helper()
	h, err := OpenBytes(buildKeyHive())
	if err != nil {
		t.Fatalf("OpenBytes: %v", err)
	}
	return h
}

func TestCreateKeySortedAndInheritsSecurity(t *testing.T) {
	h := openKeyHive(t)
	// Insert out of order; the list must come back sorted case-insensitively.
	for _, name := range []string{"System", "Software", "SAM"} {
		if err := h.CreateKey("", name); err != nil {
			t.Fatalf("CreateKey %s: %v", name, err)
		}
	}
	keys, err := h.EnumKey("")
	if err != nil {
		t.Fatalf("EnumKey: %v", err)
	}
	if !reflect.DeepEqual(keys, []string{"SAM", "Software", "System"}) {
		t.Errorf("EnumKey = %v, want [SAM Software System]", keys)
	}

	sub, err := h.FindKey("Software")
	if err != nil {
		t.Fatalf("FindKey(Software): %v", err)
	}
	if sub.IsRoot() {
		t.Error("created key should not be a root key")
	}
	if subs, _ := sub.SubKeys(); len(subs) != 0 {
		t.Errorf("new key has %d subkeys, want 0", len(subs))
	}

	// The created key inherits the root's security descriptor (shared SK record).
	sd, err := h.GetSecurityDescriptor("Software")
	if err != nil || sd == nil {
		t.Fatalf("GetSecurityDescriptor(Software): sd=%v err=%v", sd, err)
	}
	if o := sd.GetOwner(); o == nil || o.SID.String() != "S-1-5-32-544" {
		t.Errorf("inherited owner = %v, want S-1-5-32-544", o)
	}
}

func TestCreateKeyNestedAndValue(t *testing.T) {
	h := openKeyHive(t)
	if err := h.CreateKey("", "A"); err != nil {
		t.Fatalf("CreateKey A: %v", err)
	}
	if err := h.CreateKey("A", "B"); err != nil {
		t.Fatalf("CreateKey A\\B: %v", err)
	}
	if _, err := h.FindKey("A\\B"); err != nil {
		t.Errorf("FindKey(A\\B): %v", err)
	}
	if keys, _ := h.EnumKey("A"); !reflect.DeepEqual(keys, []string{"B"}) {
		t.Errorf("EnumKey(A) = %v, want [B]", keys)
	}

	// Values can be written to a freshly created key.
	if err := h.SetValue("A\\B", "v", RegDword, le32(0x2026)); err != nil {
		t.Fatalf("SetValue on created key: %v", err)
	}
	if _, d := mustGet(t, h, "A\\B", "v"); binary.LittleEndian.Uint32(d) != 0x2026 {
		t.Errorf("created-key value = % x, want 26 20 00 00", d)
	}
}

func TestCreateKeyDuplicate(t *testing.T) {
	h := openKeyHive(t)
	if err := h.CreateKey("", "Dup"); err != nil {
		t.Fatalf("CreateKey: %v", err)
	}
	if err := h.CreateKey("", "dup"); err == nil {
		t.Error("CreateKey of an existing key (case-insensitive) returned nil error")
	}
}

func TestDeleteKeyRecursive(t *testing.T) {
	h := openKeyHive(t)
	if err := h.CreateKey("", "P"); err != nil {
		t.Fatal(err)
	}
	if err := h.CreateKey("P", "C"); err != nil {
		t.Fatal(err)
	}
	if err := h.SetValue("P", "pv", RegDword, le32(1)); err != nil {
		t.Fatal(err)
	}
	if err := h.SetValue("P\\C", "cv", RegDword, le32(2)); err != nil {
		t.Fatal(err)
	}
	// Keep a sibling to confirm only the target subtree is removed.
	if err := h.CreateKey("", "Keep"); err != nil {
		t.Fatal(err)
	}

	if err := h.DeleteKey("", "P"); err != nil {
		t.Fatalf("DeleteKey: %v", err)
	}
	if _, err := h.FindKey("P"); err == nil {
		t.Error("P still present after delete")
	}
	if _, err := h.FindKey("P\\C"); err == nil {
		t.Error("P\\C still present after recursive delete")
	}
	keys, _ := h.EnumKey("")
	if !reflect.DeepEqual(keys, []string{"Keep"}) {
		t.Errorf("after delete, EnumKey = %v, want [Keep]", keys)
	}
	if err := h.DeleteKey("", "Nope"); err == nil {
		t.Error("DeleteKey of a missing key returned nil error")
	}
}

func TestCreateKeyPersistsAndReopens(t *testing.T) {
	h := openKeyHive(t)
	for _, n := range []string{"Alpha", "Beta"} {
		if err := h.CreateKey("", n); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.CreateKey("Alpha", "Child"); err != nil {
		t.Fatal(err)
	}
	raw, err := h.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	h2, err := OpenBytes(raw)
	if err != nil {
		t.Fatalf("re-open: %v", err)
	}
	if keys, _ := h2.EnumKey(""); !reflect.DeepEqual(keys, []string{"Alpha", "Beta"}) {
		t.Errorf("reopened EnumKey = %v, want [Alpha Beta]", keys)
	}
	if keys, _ := h2.EnumKey("Alpha"); !reflect.DeepEqual(keys, []string{"Child"}) {
		t.Errorf("reopened EnumKey(Alpha) = %v, want [Child]", keys)
	}
}
