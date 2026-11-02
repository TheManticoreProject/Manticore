package pac

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// TestForgeAddsHardeningBuffersByDefault checks Forge emits PAC_ATTRIBUTES_INFO
// and PAC_REQUESTOR, which DCs enforcing the November 2021 hardening require,
// and that OldPAC omits them.
func TestForgeAddsHardeningBuffersByDefault(t *testing.T) {
	info := sampleInfo(t)
	key := bytes.Repeat([]byte{0x42}, 16)

	p, err := Forge(info, "Administrator", time.Unix(1700000000, 0), 23)
	if err != nil {
		t.Fatalf("Forge: %v", err)
	}
	signed, err := p.Sign(key, key)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(signed)
	if err != nil {
		t.Fatal(err)
	}

	attr, ok := parsed.Buffer(BufferAttributesInfo)
	if !ok {
		t.Fatal("no PAC_ATTRIBUTES_INFO buffer")
	}
	if len(attr.Data) != 8 || binary.LittleEndian.Uint32(attr.Data[0:4]) != 2 || binary.LittleEndian.Uint32(attr.Data[4:8]) != PACWasRequested {
		t.Errorf("PAC_ATTRIBUTES_INFO = % x", attr.Data)
	}
	req, ok := parsed.Buffer(BufferRequestorSID)
	if !ok {
		t.Fatal("no PAC_REQUESTOR buffer")
	}
	wantSID, _ := info.UserRPCSID().Marshal()
	if !bytes.Equal(req.Data, wantSID) {
		t.Errorf("PAC_REQUESTOR SID = % x, want % x", req.Data, wantSID)
	}

	old, err := ForgeWithExtras(info, "Administrator", time.Unix(1700000000, 0), 23, ForgeExtras{OldPAC: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, typ := range []uint32{BufferAttributesInfo, BufferRequestorSID} {
		if _, ok := old.Buffer(typ); ok {
			t.Errorf("OldPAC still carries buffer type %#x", typ)
		}
	}
	// The signatures must still verify with the added buffers.
	if err := parsed.VerifyServerSignature(key); err != nil {
		t.Errorf("server signature: %v", err)
	}
	if err := parsed.VerifyKDCSignature(key); err != nil {
		t.Errorf("KDC signature: %v", err)
	}
}

// TestForgeUPNDNSInfo checks the UPN_DNS_INFO buffer: the fixed header's
// lengths and offsets point at the UTF-16LE strings and the packed SID, each
// 8-byte aligned, with the SAM-name-and-SID flag set.
func TestForgeUPNDNSInfo(t *testing.T) {
	sid, err := msdtyp.ParseSID("S-1-5-21-1111111111-2222222222-3333333333-500")
	if err != nil {
		t.Fatal(err)
	}
	u := UPNDNSInfo{UPN: "administrator@corp.local", DNSDomainName: "CORP.LOCAL", SamName: "Administrator", SID: &sid}
	data, err := MarshalUPNDNSInfo(u)
	if err != nil {
		t.Fatalf("MarshalUPNDNSInfo: %v", err)
	}

	field := func(lenOff, offOff int) []byte {
		n := int(binary.LittleEndian.Uint16(data[lenOff:]))
		o := int(binary.LittleEndian.Uint16(data[offOff:]))
		if o%8 != 0 {
			t.Errorf("offset %d is not 8-byte aligned", o)
		}
		if o+n > len(data) {
			t.Fatalf("field [%d:%d] out of bounds (%d)", o, o+n, len(data))
		}
		return data[o : o+n]
	}
	if got := decodeUTF16LE(field(0, 2)); got != u.UPN {
		t.Errorf("UPN = %q, want %q", got, u.UPN)
	}
	if got := decodeUTF16LE(field(4, 6)); got != u.DNSDomainName {
		t.Errorf("DNS domain = %q, want %q", got, u.DNSDomainName)
	}
	if binary.LittleEndian.Uint32(data[8:12]) != upnDNSFlagExtended {
		t.Errorf("Flags = %#x, want the extended flag", binary.LittleEndian.Uint32(data[8:12]))
	}
	if got := decodeUTF16LE(field(12, 14)); got != u.SamName {
		t.Errorf("SAM name = %q, want %q", got, u.SamName)
	}
	wantSID, _ := sid.Marshal()
	if got := field(16, 18); !bytes.Equal(got, wantSID) {
		t.Errorf("SID = % x, want % x", got, wantSID)
	}

	if _, err := MarshalUPNDNSInfo(UPNDNSInfo{UPN: "x"}); err == nil {
		t.Error("MarshalUPNDNSInfo accepted a nil SID")
	}
}

// decodeUTF16LE decodes UTF-16LE bytes to a string (test helper).
func decodeUTF16LE(b []byte) string {
	units := make([]uint16, len(b)/2)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(b[2*i:])
	}
	return string(utf16Runes(units))
}

func utf16Runes(units []uint16) []rune {
	var out []rune
	for _, u := range units {
		out = append(out, rune(u))
	}
	return out
}
