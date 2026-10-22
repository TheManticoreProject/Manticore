package security

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// SMB2 TRANSFORM_HEADER layout (MS-SMB2 2.2.41). The header is 52 bytes and
// precedes the encrypted SMB2 message. The bytes from the Nonce field through
// the end of the header (offsets 20..52) form the additional authenticated data
// (AAD); the ProtocolId and Signature fields are excluded. The Signature field
// carries the AEAD authentication tag.
const (
	TransformHeaderSize = 52

	transformProtocolIDOffset   = 0  // 4 bytes: 0xFD 'S' 'M' 'B'
	transformSignatureOffset    = 4  // 16 bytes: AEAD tag
	transformNonceOffset        = 20 // 16 bytes
	transformOrigMsgSizeOffset  = 36 // 4 bytes
	transformReservedOffset     = 40 // 2 bytes
	transformFlagsOffset        = 42 // 2 bytes: Flags (3.1.1) / EncryptionAlgorithm (3.0.x)
	transformSessionIDOffset    = 44 // 8 bytes
	transformAADStart           = transformNonceOffset
	transformSignatureLength    = 16
	transformNonceLength        = 16
	gcmNonceLength              = 12
	ccmNonceLength              = 11
	transformEncryptedFlagValue = 0x0001 // Flags: Encrypted / EncryptionAlgorithm: AES-128-CCM
)

// transformProtocolID is the 4-byte marker prefixing an encrypted SMB2 message:
// 0xFD 'S' 'M' 'B'.
var transformProtocolID = [4]byte{0xFD, 'S', 'M', 'B'}

// IsTransformHeader reports whether a received frame begins with the SMB2
// TRANSFORM_HEADER protocol identifier and is thus an encrypted message.
func IsTransformHeader(data []byte) bool {
	return len(data) >= 4 &&
		data[0] == transformProtocolID[0] &&
		data[1] == transformProtocolID[1] &&
		data[2] == transformProtocolID[2] &&
		data[3] == transformProtocolID[3]
}

// aeadForCipher returns the authenticated cipher and nonce length for the
// negotiated encryption algorithm, keyed with the supplied 16-byte key. For the
// 3.0/3.0.2 dialects the cipher is implicitly AES-128-CCM (cipher == 0).
func aeadForCipher(cipherID uint16, key []byte) (seal, open func(nonce, msg, aad []byte) ([]byte, error), nonceLen int, err error) {
	block, berr := aes.NewCipher(key)
	if berr != nil {
		return nil, nil, 0, fmt.Errorf("transform: %w", berr)
	}

	switch cipherID {
	case commands.SMB2_ENCRYPTION_AES128_GCM, commands.SMB2_ENCRYPTION_AES256_GCM:
		gcm, gerr := cipher.NewGCMWithNonceSize(block, gcmNonceLength)
		if gerr != nil {
			return nil, nil, 0, fmt.Errorf("transform: %w", gerr)
		}
		seal = func(nonce, msg, aad []byte) ([]byte, error) { return gcm.Seal(nil, nonce, msg, aad), nil }
		open = func(nonce, msg, aad []byte) ([]byte, error) { return gcm.Open(nil, nonce, msg, aad) }
		return seal, open, gcmNonceLength, nil
	case commands.SMB2_ENCRYPTION_AES128_CCM, commands.SMB2_ENCRYPTION_AES256_CCM, 0:
		seal = func(nonce, msg, aad []byte) ([]byte, error) { return ccmSeal(block, nonce, msg, aad) }
		open = func(nonce, msg, aad []byte) ([]byte, error) { return ccmOpen(block, nonce, msg, aad) }
		return seal, open, ccmNonceLength, nil
	default:
		return nil, nil, 0, fmt.Errorf("transform: unsupported cipher 0x%04x", cipherID)
	}
}

// TransformSessionID returns the SessionId of a TRANSFORM_HEADER frame, which
// selects the key to decrypt it with. It returns 0 for a frame too short to hold one.
func TransformSessionID(data []byte) uint64 {
	if len(data) < TransformHeaderSize {
		return 0
	}
	return binary.LittleEndian.Uint64(data[transformSessionIDOffset:])
}

// Encrypt wraps a marshalled SMB2 message in a TRANSFORM_HEADER and seals it
// with key under cipherID (MS-SMB2 3.1.4.3). nonce MUST never repeat for a key;
// it fills the leading bytes of the Nonce field.
func Encrypt(cipherID uint16, key []byte, sessionID, nonce uint64, plaintext []byte) ([]byte, error) {
	seal, _, nonceLen, err := aeadForCipher(cipherID, key)
	if err != nil {
		return nil, err
	}

	hdr := make([]byte, TransformHeaderSize)
	copy(hdr[transformProtocolIDOffset:], transformProtocolID[:])
	nonceBytes := make([]byte, nonceLen)
	binary.LittleEndian.PutUint64(nonceBytes, nonce)
	copy(hdr[transformNonceOffset:], nonceBytes)
	binary.LittleEndian.PutUint32(hdr[transformOrigMsgSizeOffset:], uint32(len(plaintext)))
	binary.LittleEndian.PutUint16(hdr[transformFlagsOffset:], transformEncryptedFlagValue)
	binary.LittleEndian.PutUint64(hdr[transformSessionIDOffset:], sessionID)

	sealed, err := seal(nonceBytes, plaintext, hdr[transformAADStart:])
	if err != nil {
		return nil, err
	}
	// Seal returns ciphertext || tag; the tag goes in the Signature field.
	ctLen := len(sealed) - transformSignatureLength
	copy(hdr[transformSignatureOffset:], sealed[ctLen:])
	return append(hdr, sealed[:ctLen]...), nil
}

// Decrypt opens a TRANSFORM_HEADER frame with key under cipherID and returns the
// plaintext SMB2 message. A tampered frame fails the AEAD tag check.
func Decrypt(cipherID uint16, key, data []byte) ([]byte, error) {
	if len(data) < TransformHeaderSize {
		return nil, fmt.Errorf("transform: frame shorter than TRANSFORM_HEADER")
	}
	_, open, nonceLen, err := aeadForCipher(cipherID, key)
	if err != nil {
		return nil, err
	}

	nonce := data[transformNonceOffset : transformNonceOffset+nonceLen]
	tag := data[transformSignatureOffset : transformSignatureOffset+transformSignatureLength]
	sealed := append(append([]byte{}, data[TransformHeaderSize:]...), tag...)

	plaintext, err := open(nonce, sealed, data[transformAADStart:TransformHeaderSize])
	if err != nil {
		return nil, fmt.Errorf("transform: decryption failed: %w", err)
	}
	if origSize := binary.LittleEndian.Uint32(data[transformOrigMsgSizeOffset:]); uint32(len(plaintext)) != origSize {
		return nil, fmt.Errorf("transform: decrypted size %d does not match OriginalMessageSize %d", len(plaintext), origSize)
	}
	return plaintext, nil
}
