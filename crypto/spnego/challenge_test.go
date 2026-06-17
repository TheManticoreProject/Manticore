package spnego

import "testing"

func TestDecodeNTHash(t *testing.T) {
	// A valid 32-hex NT hash decodes to its 16 raw bytes.
	got, err := decodeNTHash("520126a03f5d5a8d836f1c4f34ede7ce")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := [16]byte{0x52, 0x01, 0x26, 0xa0, 0x3f, 0x5d, 0x5a, 0x8d, 0x83, 0x6f, 0x1c, 0x4f, 0x34, 0xed, 0xe7, 0xce}
	if got != want {
		t.Fatalf("decodeNTHash = %x, want %x", got, want)
	}
}

func TestDecodeNTHashErrors(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"too short", "520126a0"},
		{"too long", "520126a03f5d5a8d836f1c4f34ede7ce00"},
		{"non-hex", "zz0126a03f5d5a8d836f1c4f34ede7ce"},
		{"empty", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := decodeNTHash(tc.in); err == nil {
				t.Fatalf("decodeNTHash(%q) = nil error, want error", tc.in)
			}
		})
	}
}
