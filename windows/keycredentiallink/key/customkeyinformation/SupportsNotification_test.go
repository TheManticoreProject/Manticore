package customkeyinformation

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"
)

func TestSupportsNotification_Unmarshal_Valid(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"True", []byte{1}},
		{"False", []byte{0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sn SupportsNotification
			n, err := sn.Unmarshal(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if n != 1 {
				t.Fatalf("bytes read = %d, want 1", n)
			}
		})
	}
}

func TestSupportsNotification_Unmarshal_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"TooShort", []byte{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sn SupportsNotification
			if _, err := sn.Unmarshal(tt.input); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestSupportsNotification_Marshal_Valid(t *testing.T) {
	tests := []struct {
		name     string
		value    SupportsNotification
		expected []byte
	}{
		{"True", SupportsNotification_TRUE, []byte{1}},
		{"False", SupportsNotification_FALSE, []byte{0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := tt.value.Marshal()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(out) != 1 || out[0] != tt.expected[0] {
				t.Fatalf("Marshal() = %v, want %v", out, tt.expected)
			}
		})
	}
}

// TestSupportsNotificationKeepsByte checks every byte value survives Unmarshal and
// Marshal: the field used to be a bool, so 2-255 came back as 1.
func TestSupportsNotificationKeepsByte(t *testing.T) {
	for v := 0; v < 256; v++ {
		var sn SupportsNotification
		if _, err := sn.Unmarshal([]byte{byte(v)}); err != nil {
			t.Fatal(err)
		}
		out, _ := sn.Marshal()
		if len(out) != 1 || out[0] != byte(v) {
			t.Fatalf("0x%02x re-marshals as %x", v, out)
		}
		if sn.Bool() != (v != 0) {
			t.Errorf("0x%02x: Bool() = %v", v, sn.Bool())
		}
	}
	if s := SupportsNotification(0xd1).String(); s != "True (0xd1)" {
		t.Errorf("String() = %q", s)
	}
}

// TestCustomKeyInformationSupportsNotificationRoundTrip checks the review's
// counterexample: a CustomKeyInformation whose SupportsNotification byte is 0xd1.
func TestCustomKeyInformationSupportsNotificationRoundTrip(t *testing.T) {
	data := []byte{0x01, 0x00, 0x01, 0xd1}
	var cki CustomKeyInformation
	if _, err := cki.Unmarshal(data, version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}); err != nil {
		t.Fatal(err)
	}
	out, err := cki.Marshal()
	if err != nil || string(out) != string(data) {
		t.Errorf("round trip = %x (err %v), want %x", out, err, data)
	}
}
