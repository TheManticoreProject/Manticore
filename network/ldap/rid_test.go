package ldap

import (
	"bytes"
	"testing"

	ber "github.com/go-asn1-ber/asn1-ber"
	ldapv3 "github.com/go-ldap/ldap/v3"
)

func TestEncodeLDAPBinaryFilterValue(t *testing.T) {
	got := encodeLDAPBinaryFilterValue([]byte{0x00, 'A', 'z', '0', '9', '*', 0xff})
	want := `\00Az09\2a\ff`
	if got != want {
		t.Fatalf("encodeLDAPBinaryFilterValue() = %q, want %q", got, want)
	}
}

func TestObjectSIDFilterCarriesPackedSID(t *testing.T) {
	filter, err := objectSIDFilter("S-1-5-32-544")
	if err != nil {
		t.Fatalf("objectSIDFilter returned an error: %s", err)
	}

	wantFilter := `(objectSid=\01\02\00\00\00\00\00\05\20\00\00\00\20\02\00\00)`
	if filter != wantFilter {
		t.Fatalf("objectSIDFilter() = %q, want %q", filter, wantFilter)
	}

	packet, err := ldapv3.CompileFilter(filter)
	if err != nil {
		t.Fatalf("compiled filter is invalid: %s", err)
	}
	if packet.Tag != ldapv3.FilterEqualityMatch || len(packet.Children) != 2 {
		t.Fatalf("compiled filter is not an equality match: tag=%d children=%d", packet.Tag, len(packet.Children))
	}

	wantSID := []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05, 0x20, 0x00, 0x00, 0x00, 0x20, 0x02, 0x00, 0x00}
	gotSID, ok := packet.Children[1].Value.(string)
	if !ok {
		t.Fatalf("assertion value has type %T, want string", packet.Children[1].Value)
	}
	if !bytes.Equal([]byte(gotSID), wantSID) {
		t.Fatalf("assertion value = %x, want packed SID %x", []byte(gotSID), wantSID)
	}
	if packet.Children[1].ClassType != ber.ClassUniversal || packet.Children[1].Tag != ber.TagOctetString {
		t.Fatalf("assertion value is not an LDAP octet string: class=%d tag=%d", packet.Children[1].ClassType, packet.Children[1].Tag)
	}
}

func TestObjectSIDFilterRejectsInvalidSID(t *testing.T) {
	if _, err := objectSIDFilter("not-a-sid"); err == nil {
		t.Fatal("objectSIDFilter accepted an invalid SID")
	}
}
