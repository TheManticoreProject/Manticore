package ldap

import (
	"bytes"
	"testing"
)

// dnBinaryValue builds a DN-with-binary value string for the given binary blob
// and DN, using the DNWithBinary marshaller (the exact form AD returns).
func dnBinaryValue(t *testing.T, blob []byte, dn string) string {
	t.Helper()
	dnb := DNWithBinary{DistinguishedName: dn, BinaryData: blob}
	return dnb.String()
}

func TestFilterKeyCredentialLinks(t *testing.T) {
	blobA := []byte{0x01, 0x02, 0x03, 0xaa}
	blobB := []byte{0xde, 0xad, 0xbe, 0xef}
	valA := dnBinaryValue(t, blobA, "CN=a,DC=x")
	valB := dnBinaryValue(t, blobB, "CN=b,DC=x")
	junk := "not-a-dn-binary-value"

	// keep everything but blobA (simulating a KeyID predicate that rejects A).
	keep := func(b []byte) bool { return !bytes.Equal(b, blobA) }

	survivors, removed := FilterKeyCredentialLinks([]string{valA, valB, junk}, keep)
	if !removed {
		t.Fatal("expected removed=true when a value is dropped")
	}
	if len(survivors) != 2 {
		t.Fatalf("expected 2 survivors, got %d: %v", len(survivors), survivors)
	}
	if survivors[0] != valB {
		t.Fatalf("expected valB retained, got %q", survivors[0])
	}
	// The unparseable value must be conservatively kept.
	if survivors[1] != junk {
		t.Fatalf("expected junk value retained, got %q", survivors[1])
	}
}

func TestFilterKeyCredentialLinksNoMatch(t *testing.T) {
	valA := dnBinaryValue(t, []byte{0x01}, "CN=a,DC=x")
	// keep everything.
	survivors, removed := FilterKeyCredentialLinks([]string{valA}, func([]byte) bool { return true })
	if removed {
		t.Fatal("expected removed=false when nothing matches")
	}
	if len(survivors) != 1 || survivors[0] != valA {
		t.Fatalf("expected the single value retained unchanged, got %v", survivors)
	}
}

func TestFilterKeyCredentialLinksDropAll(t *testing.T) {
	valA := dnBinaryValue(t, []byte{0x01}, "CN=a,DC=x")
	valB := dnBinaryValue(t, []byte{0x02}, "CN=b,DC=x")
	// keep nothing -> both dropped, empty survivor set signals a flush.
	survivors, removed := FilterKeyCredentialLinks([]string{valA, valB}, func([]byte) bool { return false })
	if !removed {
		t.Fatal("expected removed=true")
	}
	if len(survivors) != 0 {
		t.Fatalf("expected no survivors, got %v", survivors)
	}
}
