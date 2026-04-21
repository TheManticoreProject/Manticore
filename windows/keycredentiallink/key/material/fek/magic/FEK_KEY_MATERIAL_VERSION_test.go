package magic_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/material/fek/magic"
)

func TestFEK_KEY_MATERIAL_VERSION_UnmarshalMarshal(t *testing.T) {
	tests := []struct {
		name  string
		input magic.FEK_KEY_MATERIAL_VERSION
	}{
		{
			name: "Version 1",
			input: magic.FEK_KEY_MATERIAL_VERSION{
				Version: magic.FEK_KEY_VERSION_1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshalledData, err := tt.input.Marshal()
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if len(marshalledData) != 1 {
				t.Errorf("Marshal() produced %d bytes, want 1", len(marshalledData))
			}

			parsed := magic.FEK_KEY_MATERIAL_VERSION{}
			n, err := parsed.Unmarshal(marshalledData)
			if err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if n != 1 {
				t.Errorf("Unmarshal() consumed %d bytes, want 1", n)
			}
			if !parsed.Equal(&tt.input) {
				t.Errorf("Unmarshal() = %v, want %v", parsed, tt.input)
			}
		})
	}
}

func TestFEK_KEY_MATERIAL_VERSION_Unmarshal_ShortInput(t *testing.T) {
	v := magic.FEK_KEY_MATERIAL_VERSION{}
	_, err := v.Unmarshal([]byte{})
	if err == nil {
		t.Fatal("Expected error on empty buffer, got nil")
	}
}
