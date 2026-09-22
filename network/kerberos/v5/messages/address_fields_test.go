package messages

import (
	"bytes"
	"testing"
	"time"
)

func TestEncTicketPartAddressRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	want := []HostAddress{
		{AddrType: 2, Address: []byte{198, 51, 100, 8}},
		{AddrType: 24, Address: bytes.Repeat([]byte{0x20}, 16)},
	}
	part := EncTicketPart{
		Flags:  NewKerberosFlags(),
		Key:    EncryptionKey{KeyType: ETypeAES256CTSHMACSHA196, KeyValue: bytes.Repeat([]byte{1}, 32)},
		CRealm: "CORP.LOCAL", CName: PrincipalName{NameType: NameTypePrincipal, NameString: []string{"alice"}},
		Transited: TransitedEncoding{TRType: 0, Contents: []byte{}},
		AuthTime:  now, EndTime: now.Add(time.Hour), CAddr: want,
		AuthorizationData: []AuthorizationData{{ADType: 42, ADData: []byte{1, 2, 3}}},
	}
	wire, err := part.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var got EncTicketPart
	if _, err := got.Unmarshal(wire); err != nil {
		t.Fatal(err)
	}
	if len(got.CAddr) != len(want) {
		t.Fatalf("caddr length: got %d want %d", len(got.CAddr), len(want))
	}
	for i := range want {
		if got.CAddr[i].AddrType != want[i].AddrType || !bytes.Equal(got.CAddr[i].Address, want[i].Address) {
			t.Errorf("caddr[%d]: got %+v want %+v", i, got.CAddr[i], want[i])
		}
	}
	if len(got.AuthorizationData) != 1 || got.AuthorizationData[0].ADType != 42 {
		t.Errorf("authorization data following caddr was not preserved: %+v", got.AuthorizationData)
	}
	if got.Flags.BitLength != 32 {
		t.Errorf("ticket flags not preserved")
	}
}
