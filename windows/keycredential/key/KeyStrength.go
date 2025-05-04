package key

import "encoding/binary"

// KeyStrength specifies the strength of the NGC key.
// See: https://msdn.microsoft.com/en-us/library/mt220496.aspx
type KeyStrength struct {
	Name  string
	Value uint32

	// Internal
	RawBytes     []byte
	RawBytesSize uint32
}

const (
	// Key strength is unknown.
	KeyStrength_Unknown uint32 = 0x00

	// Key strength is weak.
	KeyStrength_Weak uint32 = 0x01

	// Key strength is normal.
	KeyStrength_Normal uint32 = 0x02
)

func (ks *KeyStrength) FromBytes(value []byte) {
	ks.RawBytes = value[:4]
	ks.RawBytesSize = 4

	ks.Value = binary.LittleEndian.Uint32(value[:4])

	switch ks.Value {
	case KeyStrength_Unknown:
		ks.Name = "Unknown"
	case KeyStrength_Weak:
		ks.Name = "Weak"
	case KeyStrength_Normal:
		ks.Name = "Normal"
	}
}
