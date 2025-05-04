package crypto

import (
	"bytes"
	"testing"
)

func TestRSAKeyMaterial_ToBytes_FromBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   RSAKeyMaterial
		wantErr bool
	}{
		{
			name: "Valid RSA key material",
			input: RSAKeyMaterial{
				KeySize:  8,
				Exponent: 0b10000000000000001,
				Modulus:  []byte{0x11, 0x11, 0x11, 0x11},
				Prime1:   []byte{0x22, 0x22, 0x22, 0x22},
				Prime2:   []byte{0x33, 0x33, 0x33, 0x33},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rk := &RSAKeyMaterial{}
			err := rk.FromBytes(tt.input.ToBytes())
			if (err != nil) != tt.wantErr {
				t.Errorf("FromBytes() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.input.KeySize != rk.KeySize {
				t.Errorf("KeySize = %v, want %v", rk.KeySize, tt.input.KeySize)
			}
			if tt.input.Exponent != rk.Exponent {
				t.Errorf("Exponent = %v, want %v", rk.Exponent, tt.input.Exponent)
			}
			if !bytes.Equal(tt.input.Modulus, rk.Modulus) {
				t.Errorf("Modulus = %v, want %v", rk.Modulus, tt.input.Modulus)
			}
			if !bytes.Equal(tt.input.Prime1, rk.Prime1) {
				t.Errorf("Prime1 = %v, want %v", rk.Prime1, tt.input.Prime1)
			}
			if !bytes.Equal(tt.input.Prime2, rk.Prime2) {
				t.Errorf("Prime2 = %v, want %v", rk.Prime2, tt.input.Prime2)
			}
		})
	}
}
