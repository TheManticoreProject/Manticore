package headers_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/material/fek/headers"
)

func TestFEK_KEY_MATERIAL_HEADER_Unmarshal_Marshal(t *testing.T) {
	tests := []struct {
		name  string
		input headers.FEK_KEY_MATERIAL_HEADER
	}{
		{
			name: "Nominal RSA-2048 + AES-256",
			input: headers.FEK_KEY_MATERIAL_HEADER{
				BitLength:   2048,
				CbPublicExp: 3,
				CbModulus:   256,
				CbAESKDFKey: 32,
			},
		},
		{
			name: "Larger public exponent",
			input: headers.FEK_KEY_MATERIAL_HEADER{
				BitLength:   2048,
				CbPublicExp: 8,
				CbModulus:   256,
				CbAESKDFKey: 32,
			},
		},
		{
			name: "Zero-sized",
			input: headers.FEK_KEY_MATERIAL_HEADER{
				BitLength:   0,
				CbPublicExp: 0,
				CbModulus:   0,
				CbAESKDFKey: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.input.Marshal()
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if len(data) != 16 {
				t.Errorf("Marshal() produced %d bytes, want 16", len(data))
			}

			parsed := headers.FEK_KEY_MATERIAL_HEADER{}
			n, err := parsed.Unmarshal(data)
			if err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if n != 16 {
				t.Errorf("Unmarshal() consumed %d bytes, want 16", n)
			}
			if !parsed.Equal(&tt.input) {
				t.Errorf("Unmarshal() = %v, want %v", parsed, tt.input)
			}
		})
	}
}

func TestFEK_KEY_MATERIAL_HEADER_Unmarshal_ShortInput(t *testing.T) {
	h := headers.FEK_KEY_MATERIAL_HEADER{}
	_, err := h.Unmarshal([]byte{0x00, 0x08, 0x00, 0x00})
	if err == nil {
		t.Fatal("Expected error on short buffer, got nil")
	}
}
