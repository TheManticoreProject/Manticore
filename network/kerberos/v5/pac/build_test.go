package pac

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

func sampleInfo(t *testing.T) *KERB_VALIDATION_INFO {
	t.Helper()
	dom, err := msdtyp.ParseSID("S-1-5-21-1111111111-2222222222-3333333333")
	if err != nil {
		t.Fatalf("ParseSID: %v", err)
	}
	extra, err := msdtyp.ParseSID("S-1-5-21-1111111111-2222222222-3333333333-519")
	if err != nil {
		t.Fatalf("ParseSID extra: %v", err)
	}
	never := NeverExpireFileTime()
	return &KERB_VALIDATION_INFO{
		LogonTime:          FileTimeFromTime(time.Unix(1700000000, 0)),
		LogoffTime:         never,
		KickOffTime:        never,
		PasswordMustChange: never,
		EffectiveName:      NewUnicodeString("Administrator"),
		FullName:           NewUnicodeString(""),
		LogonScript:        NewUnicodeString(""),
		ProfilePath:        NewUnicodeString(""),
		HomeDirectory:      NewUnicodeString(""),
		HomeDirectoryDrive: NewUnicodeString(""),
		UserId:             500,
		PrimaryGroupId:     513,
		GroupCount:         3,
		GroupIds: []GROUP_MEMBERSHIP{
			{RelativeId: 513, Attributes: DefaultGroupAttributes},
			{RelativeId: 512, Attributes: DefaultGroupAttributes},
			{RelativeId: 519, Attributes: DefaultGroupAttributes},
		},
		UserFlags:          0x20,
		LogonServer:        NewUnicodeString("DC01"),
		LogonDomainName:    NewUnicodeString("CORP"),
		LogonDomainId:      &dom,
		UserAccountControl: 0x00010200,
		SidCount:           1,
		ExtraSids: []KERB_SID_AND_ATTRIBUTES{
			{Sid: &extra, Attributes: DefaultGroupAttributes},
		},
	}
}

func TestKerbValidationInfoRoundtrip(t *testing.T) {
	info := sampleInfo(t)
	data, err := MarshalKerbValidationInfo(info)
	if err != nil {
		t.Fatalf("MarshalKerbValidationInfo: %v", err)
	}

	// Type-serialization header sanity ([MS-RPCE] 2.2.6).
	if data[0] != 0x01 || data[1] != 0x10 {
		t.Errorf("common header version/endianness = %#x %#x, want 0x01 0x10", data[0], data[1])
	}
	if binary.LittleEndian.Uint16(data[2:]) != 8 {
		t.Errorf("CommonHeaderLength = %d, want 8", binary.LittleEndian.Uint16(data[2:]))
	}
	if binary.LittleEndian.Uint32(data[4:]) != 0xCCCCCCCC {
		t.Errorf("common header filler = %#x, want 0xCCCCCCCC", binary.LittleEndian.Uint32(data[4:]))
	}
	objLen := binary.LittleEndian.Uint32(data[8:])
	if int(objLen) != len(data)-16 {
		t.Errorf("ObjectBufferLength = %d, want %d", objLen, len(data)-16)
	}
	if len(data)%8 != 0 {
		t.Errorf("serialized logon info not 8-octet aligned: %d bytes", len(data))
	}

	got, err := UnmarshalKerbValidationInfo(data)
	if err != nil {
		t.Fatalf("UnmarshalKerbValidationInfo: %v", err)
	}
	if got.UserId != 500 || got.PrimaryGroupId != 513 || got.GroupCount != 3 {
		t.Errorf("scalars mismatch: UserId=%d PrimaryGroupId=%d GroupCount=%d", got.UserId, got.PrimaryGroupId, got.GroupCount)
	}
	if got.EffectiveName.Length != uint16(2*len("Administrator")) {
		t.Errorf("EffectiveName.Length = %d, want %d", got.EffectiveName.Length, 2*len("Administrator"))
	}
	if s := decodeUTF16(got.EffectiveName.Buffer); s != "Administrator" {
		t.Errorf("EffectiveName = %q, want Administrator", s)
	}
	if len(got.GroupIds) != 3 || got.GroupIds[1].RelativeId != 512 {
		t.Errorf("GroupIds mismatch: %+v", got.GroupIds)
	}
	if got.SidCount != 1 || len(got.ExtraSids) != 1 {
		t.Fatalf("ExtraSids mismatch: SidCount=%d len=%d", got.SidCount, len(got.ExtraSids))
	}
	if got.ExtraSids[0].Sid == nil || got.ExtraSids[0].Sid.String() != "S-1-5-21-1111111111-2222222222-3333333333-519" {
		t.Errorf("ExtraSids[0] = %v", got.ExtraSids[0].Sid)
	}
	if got.LogonDomainId == nil || got.LogonDomainId.String() != "S-1-5-21-1111111111-2222222222-3333333333" {
		t.Errorf("LogonDomainId = %v", got.LogonDomainId)
	}
}

func TestForgeAndSignPAC(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 16) // RC4 NT hash
	info := sampleInfo(t)

	p, err := Forge(info, "Administrator", time.Unix(1700000000, 0), 23)
	if err != nil {
		t.Fatalf("Forge: %v", err)
	}
	signed, err := p.Sign(key, key)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	parsed, err := Parse(signed)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, ok := parsed.Buffer(BufferLogonInfo); !ok {
		t.Error("no logon info buffer")
	}
	if _, ok := parsed.Buffer(BufferClientInfo); !ok {
		t.Error("no client info buffer")
	}
	if err := parsed.VerifyServerSignature(key); err != nil {
		t.Errorf("server signature: %v", err)
	}
	if err := parsed.VerifyKDCSignature(key); err != nil {
		t.Errorf("KDC signature: %v", err)
	}

	// The logon info recovered from the signed PAC must still decode.
	lb, _ := parsed.Buffer(BufferLogonInfo)
	if _, err := UnmarshalKerbValidationInfo(lb.Data); err != nil {
		t.Errorf("logon info in signed PAC did not decode: %v", err)
	}
}

// TestForgeAndSignPACAESSHA2 forges, signs and verifies a PAC with the AES-SHA2
// enctypes (19 and 20), exercising the per-etype signature-type selection and the
// RFC 8009 keyed checksum through the PAC Sign/Verify path end to end. The
// underlying checksum is validated against RFC 8009 Appendix A vectors in the
// crypto package (TestAES2ChecksumVectors); this test confirms the PAC layer
// routes the right signature type and verifies what it signed. Covers #1933.
func TestForgeAndSignPACAESSHA2(t *testing.T) {
	info := sampleInfo(t)
	for _, c := range []struct {
		name    string
		etype   int
		keyLen  int
		sigType uint32
		sigSize int
	}{
		{"AES128-SHA256", 19, 16, 0x00000013, 16},
		{"AES256-SHA384", 20, 32, 0x00000014, 24},
	} {
		t.Run(c.name, func(t *testing.T) {
			key := bytes.Repeat([]byte{0x42}, c.keyLen)

			p, err := Forge(info, "Administrator", time.Unix(1700000000, 0), c.etype)
			if err != nil {
				t.Fatalf("Forge: %v", err)
			}
			srv, ok := p.Buffer(BufferServerChecksum)
			if !ok || binary.LittleEndian.Uint32(srv.Data) != c.sigType {
				t.Fatalf("server checksum SignatureType = %#x, want %#x", binary.LittleEndian.Uint32(srv.Data), c.sigType)
			}
			if len(srv.Data) != 4+c.sigSize {
				t.Fatalf("server checksum buffer = %d bytes, want %d", len(srv.Data), 4+c.sigSize)
			}

			signed, err := p.Sign(key, key)
			if err != nil {
				t.Fatalf("Sign: %v", err)
			}
			parsed, err := Parse(signed)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if err := parsed.VerifyServerSignature(key); err != nil {
				t.Errorf("server signature: %v", err)
			}
			if err := parsed.VerifyKDCSignature(key); err != nil {
				t.Errorf("KDC signature: %v", err)
			}
			// A different key must not verify: proves the AES-SHA2 checksum is
			// actually keyed, not a fixed or empty value.
			wrong := bytes.Repeat([]byte{0x24}, c.keyLen)
			if err := parsed.VerifyServerSignature(wrong); err == nil {
				t.Error("server signature verified under the wrong key")
			}
		})
	}
}

func TestClientInfoLayout(t *testing.T) {
	ci := MarshalClientInfo(time.Unix(1700000000, 0), "Administrator")
	nameLen := binary.LittleEndian.Uint16(ci[8:])
	if int(nameLen) != 2*len("Administrator") {
		t.Errorf("NameLength = %d, want %d", nameLen, 2*len("Administrator"))
	}
	if len(ci) != 10+int(nameLen) {
		t.Errorf("client info length = %d, want %d", len(ci), 10+int(nameLen))
	}
	if decodeUTF16LEBytes(ci[10:]) != "Administrator" {
		t.Errorf("client name = %q", decodeUTF16LEBytes(ci[10:]))
	}
}

func TestSignatureTypeForEType(t *testing.T) {
	cases := []struct {
		etype   int
		sigType uint32
		size    int
	}{
		{23, 0xFFFFFF76, 16},
		{17, 0x0000000F, 12},
		{18, 0x00000010, 12},
		{19, 0x00000013, 16}, // aes128-cts-hmac-sha256-128 (RFC 8009)
		{20, 0x00000014, 24}, // aes256-cts-hmac-sha384-192 (RFC 8009)
	}
	for _, c := range cases {
		st, sz, ok := SignatureTypeForEType(c.etype)
		if !ok || st != c.sigType || sz != c.size {
			t.Errorf("etype %d: got (%#x,%d,%v), want (%#x,%d,true)", c.etype, st, sz, ok, c.sigType, c.size)
		}
	}
	if _, _, ok := SignatureTypeForEType(999); ok {
		t.Error("expected unsupported etype to report false")
	}
}

func decodeUTF16(units []uint16) string {
	var b []rune
	for _, u := range units {
		b = append(b, rune(u))
	}
	return string(b)
}

func decodeUTF16LEBytes(b []byte) string {
	var r []rune
	for i := 0; i+1 < len(b); i += 2 {
		r = append(r, rune(binary.LittleEndian.Uint16(b[i:])))
	}
	return string(r)
}
