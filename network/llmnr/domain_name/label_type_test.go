package domain_name_test

import (
	"bytes"
	stderrors "errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/llmnr/domain_name"
	"github.com/TheManticoreProject/Manticore/network/llmnr/errors"
)

// TestDecodeDomainNameReservedLabelTypeRejected verifies that a length octet
// whose top two bits are 01 or 10 is rejected. Only 00 (a label of at most 63
// octets) and 11 (a compression pointer) are defined by RFC 1035 §4.1.4; the
// other two combinations must not be read as 64..191-octet labels.
func TestDecodeDomainNameReservedLabelTypeRejected(t *testing.T) {
	for _, lengthOctet := range []byte{0x40, 0x7F, 0x80, 0xBF} {
		data := append([]byte{lengthOctet}, bytes.Repeat([]byte{'a'}, int(lengthOctet))...)
		data = append(data, 0)

		name, _, err := domain_name.DecodeDomainName(data, 0)
		if err == nil {
			t.Errorf("length octet 0x%02X: decoded %d-byte name, want error", lengthOctet, len(name))
			continue
		}
		if !stderrors.Is(err, errors.ErrInvalidDomainName) {
			t.Errorf("length octet 0x%02X: error %v does not wrap ErrInvalidDomainName", lengthOctet, err)
		}
	}
}

// TestDecodeDomainNameTooLongRejected verifies that a name whose wire encoding
// exceeds 255 octets (RFC 1035 §3.1) is rejected, including when the length is
// reached by following compression pointers.
func TestDecodeDomainNameTooLongRejected(t *testing.T) {
	label := append([]byte{63}, bytes.Repeat([]byte{'a'}, 63)...)

	// Four 63-octet labels plus the root octet is 4*64+1 = 257 octets.
	flat := append(bytes.Repeat(label, 4), 0)
	if _, _, err := domain_name.DecodeDomainName(flat, 0); !stderrors.Is(err, errors.ErrNameTooLong) {
		t.Errorf("257-octet name: error = %v, want ErrNameTooLong", err)
	}

	// Three 63-octet labels plus the root octet (193 octets) followed by a
	// name made of one more 63-octet label and a pointer back to offset 0:
	// the expanded name is again 257 octets.
	msg := append(bytes.Repeat(label, 3), 0)
	start := len(msg)
	msg = append(msg, label...)
	msg = append(msg, 0xC0, 0x00)
	if _, _, err := domain_name.DecodeDomainName(msg, start); !stderrors.Is(err, errors.ErrNameTooLong) {
		t.Errorf("257-octet compressed name: error = %v, want ErrNameTooLong", err)
	}

	// A name of exactly 255 octets is still accepted: three 63-octet labels,
	// one 61-octet label and the root octet.
	exact := append(bytes.Repeat(label, 3), 61)
	exact = append(exact, bytes.Repeat([]byte{'b'}, 61)...)
	exact = append(exact, 0)
	if len(exact) != 255 {
		t.Fatalf("test setup: name is %d octets, want 255", len(exact))
	}
	if _, _, err := domain_name.DecodeDomainName(exact, 0); err != nil {
		t.Errorf("255-octet name: unexpected error %v", err)
	}
}
