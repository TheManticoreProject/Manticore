package domain_name_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/llmnr/domain_name"
	"github.com/TheManticoreProject/Manticore/network/llmnr/errors"
)

func TestValidateDomainName(t *testing.T) {
	tests := []struct {
		name     string
		expected error
	}{
		{"example.com", nil},
		{"a-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-long-domain-name-that-exceeds-the-maximum-allowed-length-for-a-domain-name.com", errors.ErrNameTooLong},
		{"valid-domain.local", nil},
		{"another.valid-domain.local", nil},
		{"invalid_domain_with_underscores.com", nil}, // Assuming underscores are allowed in this context
		{"", nil}, // Empty domain name should be valid
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := domain_name.ValidateDomainName(test.name)
			if err != test.expected {
				t.Errorf("ValidateDomainName = %v; want %v", err, test.expected)
			}
		})
	}
}

func TestEncodeDomainName(t *testing.T) {
	tests := []struct {
		name     string
		expected []byte
	}{
		{"example.com", []byte{7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}},
	}

	for _, test := range tests {
		t.Run("EncodeDomainName", func(t *testing.T) {
			encoded, err := domain_name.EncodeDomainName(test.name)
			if err != nil {
				t.Fatalf("failed to encode domain name: %v", err)
			}
			if !bytes.Equal(encoded, test.expected) {
				t.Errorf("EncodeDomainName = %v; want %v", encoded, test.expected)
			}
		})
	}
}

func TestDecodeDomainName(t *testing.T) {
	tests := []struct {
		data     []byte
		expected string
	}{
		{
			data:     []byte{7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0},
			expected: "example.com",
		},
	}

	for _, test := range tests {
		t.Run("DecodeDomainName", func(t *testing.T) {
			var offset int
			decoded, _, err := domain_name.DecodeDomainName(test.data, offset)
			if err != nil {
				t.Fatalf("failed to decode domain name: %v", err)
			}
			if decoded != test.expected {
				t.Errorf("DecodeDomainName = %v; want %v", decoded, test.expected)
			}
		})
	}
}

// TestDecodeDomainNameCompressionPointer verifies that a 0xC0 compression
// pointer is resolved relative to the start of the message (RFC 1035 §4.1.4)
// and that it consumes exactly two bytes at the point it appears, regardless of
// the length of the name it references.
func TestDecodeDomainNameCompressionPointer(t *testing.T) {
	// The full name "example.com" lives at offset 2, preceded by two filler
	// bytes so the pointer offset is non-zero. A second name at offset 15 is a
	// single label "www" followed by a pointer (0xC0 0x02) back to the name at
	// offset 2, yielding "www.example.com".
	msg := []byte{
		0xAA, 0xBB, // filler (e.g. would be part of the header in a real message)
		7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0, // offset 2..14
		3, 'w', 'w', 'w', 0xC0, 0x02, // offset 15: "www" + pointer to offset 2
	}

	// Decode the compressed name at offset 15.
	name, newOffset, err := domain_name.DecodeDomainName(msg, 15)
	if err != nil {
		t.Fatalf("failed to decode compressed name: %v", err)
	}
	if name != "www.example.com" {
		t.Errorf("compressed name = %q; want %q", name, "www.example.com")
	}
	// "www" label (1+3 bytes) + pointer (2 bytes) = 6 bytes consumed in the
	// original stream: offset 15 -> 21.
	if newOffset != 21 {
		t.Errorf("new offset = %d; want %d", newOffset, 21)
	}

	// Decoding the referenced name directly must yield the suffix.
	suffix, _, err := domain_name.DecodeDomainName(msg, 2)
	if err != nil {
		t.Fatalf("failed to decode referenced name: %v", err)
	}
	if suffix != "example.com" {
		t.Errorf("referenced name = %q; want %q", suffix, "example.com")
	}
}

// TestDecodeDomainNamePointerLoopRejected ensures malformed packets with
// compression pointer loops (self, forward, or cyclic references) are rejected
// with an error instead of looping forever or panicking.
func TestDecodeDomainNamePointerLoopRejected(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		// offset to start decoding at
		offset int
	}{
		{
			name:   "self pointer at offset 0",
			data:   []byte{0xC0, 0x00},
			offset: 0,
		},
		{
			name: "forward pointer",
			// offset 0: pointer to offset 2; offset 2: pointer to offset 0.
			// The first pointer references a later offset and is rejected.
			data:   []byte{0xC0, 0x02, 0xC0, 0x00},
			offset: 0,
		},
		{
			name: "label then self pointer",
			// "a" label followed by a pointer that jumps back onto itself.
			data:   []byte{1, 'a', 0xC0, 0x02},
			offset: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			done := make(chan struct{})
			var err error
			go func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("DecodeDomainName panicked: %v", r)
					}
					close(done)
				}()
				_, _, err = domain_name.DecodeDomainName(test.data, test.offset)
			}()

			select {
			case <-done:
				if err == nil {
					t.Errorf("expected error for pointer loop, got nil")
				}
			case <-time.After(2 * time.Second):
				t.Fatalf("DecodeDomainName did not terminate (pointer loop)")
			}
		})
	}
}
