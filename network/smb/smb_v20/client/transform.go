package client

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
	transformHeaderSize = 52

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

// isTransformHeader reports whether a received frame begins with the SMB2
// TRANSFORM_HEADER protocol identifier and is thus an encrypted message.
func isTransformHeader(data []byte) bool {
	return len(data) >= 4 &&
		data[0] == transformProtocolID[0] &&
		data[1] == transformProtocolID[1] &&
		data[2] == transformProtocolID[2] &&
		data[3] == transformProtocolID[3]
}

// aeadForCipher returns the authenticated cipher and nonce length for the
// negotiated encryption algorithm, keyed with the supplied 16-byte key. For the
// 3.0/3.0.2 dialects the cipher is implicitly AES-128-CCM (cipher == 0).
func aeadForCipher(cipherID uint16, key []byte) (seal, open func(nonce, msg, aad []byte) ([]byte, error), nonceLen int, algo uint16, err error) {
	block, berr := aes.NewCipher(key)
	if berr != nil {
		return nil, nil, 0, 0, fmt.Errorf("transform: %w", berr)
	}

	switch cipherID {
	case commands.SMB2_ENCRYPTION_AES128_GCM, commands.SMB2_ENCRYPTION_AES256_GCM:
		gcm, gerr := cipher.NewGCMWithNonceSize(block, gcmNonceLength)
		if gerr != nil {
			return nil, nil, 0, 0, fmt.Errorf("transform: %w", gerr)
		}
		seal = func(nonce, msg, aad []byte) ([]byte, error) { return gcm.Seal(nil, nonce, msg, aad), nil }
		open = func(nonce, msg, aad []byte) ([]byte, error) { return gcm.Open(nil, nonce, msg, aad) }
		return seal, open, gcmNonceLength, cipherID, nil
	case commands.SMB2_ENCRYPTION_AES128_CCM, commands.SMB2_ENCRYPTION_AES256_CCM, 0:
		seal = func(nonce, msg, aad []byte) ([]byte, error) { return ccmSeal(block, nonce, msg, aad) }
		open = func(nonce, msg, aad []byte) ([]byte, error) { return ccmOpen(block, nonce, msg, aad) }
		id := cipherID
		if id == 0 {
			id = commands.SMB2_ENCRYPTION_AES128_CCM
		}
		return seal, open, ccmNonceLength, id, nil
	default:
		return nil, nil, 0, 0, fmt.Errorf("transform: unsupported cipher 0x%04x", cipherID)
	}
}

// encryptMessage wraps a marshalled SMB2 message in an SMB2 TRANSFORM_HEADER and
// encrypts it with the session's EncryptionKey and the connection's negotiated
// cipher (MS-SMB2 3.1.4.3). The 16-byte Nonce field is filled from a
// monotonically increasing per-session counter so a nonce is never reused; the
// AEAD tag is placed in the Signature field.
func (c *Client) encryptMessage(plaintext []byte) ([]byte, error) {
	if c.Session == nil || len(c.Session.EncryptionKey) == 0 {
		return nil, fmt.Errorf("transform: no session encryption key")
	}

	seal, _, nonceLen, algo, err := aeadForCipher(c.Connection.Cipher, c.Session.EncryptionKey)
	if err != nil {
		return nil, err
	}

	hdr := make([]byte, transformHeaderSize)
	copy(hdr[transformProtocolIDOffset:], transformProtocolID[:])

	// Unique nonce: a per-session counter encoded little-endian in the leading
	// bytes of the (16-byte) Nonce field. The unused trailing bytes stay zero, as
	// the spec requires for both the AES-CCM and AES-GCM nonce structures.
	c.Session.nonceCounter++
	nonce := make([]byte, nonceLen)
	binary.LittleEndian.PutUint64(nonce, c.Session.nonceCounter)
	copy(hdr[transformNonceOffset:transformNonceOffset+transformNonceLength], nonce)

	binary.LittleEndian.PutUint32(hdr[transformOrigMsgSizeOffset:], uint32(len(plaintext)))
	binary.LittleEndian.PutUint16(hdr[transformFlagsOffset:], transformEncryptedFlagValue)
	_ = algo // Flags == EncryptionAlgorithm value (0x0001) for the ciphers we use.
	binary.LittleEndian.PutUint64(hdr[transformSessionIDOffset:], c.Session.SessionId)

	aad := hdr[transformAADStart:transformHeaderSize]
	sealed, err := seal(nonce, plaintext, aad)
	if err != nil {
		return nil, err
	}

	// Seal returns ciphertext || tag; split the trailing tag into the Signature
	// field and keep the ciphertext as the transformed message body.
	if len(sealed) < transformSignatureLength {
		return nil, fmt.Errorf("transform: sealed output too short")
	}
	ctLen := len(sealed) - transformSignatureLength
	copy(hdr[transformSignatureOffset:transformSignatureOffset+transformSignatureLength], sealed[ctLen:])

	out := make([]byte, 0, transformHeaderSize+ctLen)
	out = append(out, hdr...)
	out = append(out, sealed[:ctLen]...)
	return out, nil
}

// decryptMessage unwraps and decrypts an SMB2 TRANSFORM_HEADER frame with the
// session's DecryptionKey, returning the plaintext SMB2 message. The AEAD tag is
// taken from the Signature field and reattached to the ciphertext for
// verification, so a tampered message fails to decrypt.
func (c *Client) decryptMessage(data []byte) ([]byte, error) {
	if c.Session == nil || len(c.Session.DecryptionKey) == 0 {
		return nil, fmt.Errorf("transform: no session decryption key")
	}
	if len(data) < transformHeaderSize {
		return nil, fmt.Errorf("transform: frame shorter than TRANSFORM_HEADER")
	}

	_, open, nonceLen, _, err := aeadForCipher(c.Connection.Cipher, c.Session.DecryptionKey)
	if err != nil {
		return nil, err
	}

	origSize := binary.LittleEndian.Uint32(data[transformOrigMsgSizeOffset:])
	nonce := make([]byte, nonceLen)
	copy(nonce, data[transformNonceOffset:transformNonceOffset+nonceLen])
	aad := data[transformAADStart:transformHeaderSize]
	tag := data[transformSignatureOffset : transformSignatureOffset+transformSignatureLength]
	ciphertext := data[transformHeaderSize:]

	sealed := make([]byte, 0, len(ciphertext)+transformSignatureLength)
	sealed = append(sealed, ciphertext...)
	sealed = append(sealed, tag...)

	plaintext, err := open(nonce, sealed, aad)
	if err != nil {
		return nil, fmt.Errorf("transform: decryption failed: %w", err)
	}
	if uint32(len(plaintext)) != origSize {
		return nil, fmt.Errorf("transform: decrypted size %d does not match OriginalMessageSize %d", len(plaintext), origSize)
	}
	return plaintext, nil
}
