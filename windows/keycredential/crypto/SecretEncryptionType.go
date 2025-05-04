package crypto

import "encoding/binary"

type SecretEncryptionType struct {
	Value int

	// Internal
	RawBytes     []byte
	RawBytesSize uint32
}

const (
	// TODO: Add support for SAM encryption types

	// Database secret encryption using PEK without salt.
	// <remarks>Used until Windows Server 2000 Beta 2</remarks>
	SecretEncryptionType_DatabaseRC4 int = 0x10

	// Database secret encryption using PEK with salt.
	// <remarks>Used in Windows Server 2000 - Windows Server 2012 R2.</remarks>
	SecretEncryptionType_DatabaseRC4WithSalt int = 0x11

	// Replicated secret encryption using Session Key with salt.
	SecretEncryptionType_ReplicationRC4WithSalt int = 0x12

	// Database secret encryption using PEK and AES.
	// <remarks>Used since Windows Server 2016 TP4.</remarks>
	SecretEncryptionType_DatabaseAES int = 0x13
)

// FromBytes parses the SecretEncryptionType from a byte array.
//
// Parameters:
// - value: A byte array containing the SecretEncryptionType to be parsed.
func (set *SecretEncryptionType) FromBytes(value []byte) {
	set.RawBytes = value[:4]
	set.RawBytesSize = 4

	set.Value = int(binary.LittleEndian.Uint32(value[:4]))
}

// ToBytes returns the raw bytes of the SecretEncryptionType.
//
// Returns:
// - A byte array containing the raw bytes of the SecretEncryptionType.
func (set *SecretEncryptionType) ToBytes() []byte {
	blob := make([]byte, 4)
	binary.LittleEndian.PutUint32(blob, uint32(set.Value))
	return blob
}

// String returns the string representation of the SecretEncryptionType.
func (set *SecretEncryptionType) String() string {
	switch set.Value {
	case SecretEncryptionType_DatabaseRC4:
		return "DatabaseRC4"
	case SecretEncryptionType_DatabaseRC4WithSalt:
		return "DatabaseRC4WithSalt"
	case SecretEncryptionType_ReplicationRC4WithSalt:
		return "ReplicationRC4WithSalt"
	case SecretEncryptionType_DatabaseAES:
		return "DatabaseAES"
	default:
		return "Unknown"
	}
}
