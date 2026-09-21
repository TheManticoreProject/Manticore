package client

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/crypto/cmac"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
)

// Offsets of the Flags and Signature fields within the 64-byte SMB2 header.
const (
	signFlagsOffset     = 16
	signSignatureOffset = 48
	signSignatureLength = 16
)

// signMessage signs a marshalled SMB2 message in place using HMAC-SHA256, the
// signing algorithm for the SMB 2.0.2 and 2.1 dialects (MS-SMB2 3.1.4.1):
//
//  1. Set SMB2_FLAGS_SIGNED in the header Flags.
//  2. Zero the 16-byte Signature field.
//  3. Compute HMAC-SHA256 over the entire message with the signing key.
//  4. Copy the first 16 bytes of the digest into the Signature field.
//
// The flag is set before hashing because it is covered by the signature.
func signMessage(key, message []byte) {
	if len(message) < header.SMB2_HEADER_SIZE {
		return
	}

	f := binary.LittleEndian.Uint32(message[signFlagsOffset : signFlagsOffset+4])
	f |= uint32(flags.SMB2_FLAGS_SIGNED)
	binary.LittleEndian.PutUint32(message[signFlagsOffset:signFlagsOffset+4], f)

	for i := 0; i < signSignatureLength; i++ {
		message[signSignatureOffset+i] = 0
	}

	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	digest := mac.Sum(nil)
	copy(message[signSignatureOffset:signSignatureOffset+signSignatureLength], digest[:signSignatureLength])
}

// verifySignature recomputes the HMAC-SHA256 signature of a received SMB2 message
// and compares it (in constant time) to the Signature the message carries. The
// computation is done on a copy so the caller's buffer is left intact.
func verifySignature(key, message []byte) bool {
	if len(message) < header.SMB2_HEADER_SIZE {
		return false
	}

	received := make([]byte, signSignatureLength)
	copy(received, message[signSignatureOffset:signSignatureOffset+signSignatureLength])

	work := make([]byte, len(message))
	copy(work, message)
	for i := 0; i < signSignatureLength; i++ {
		work[signSignatureOffset+i] = 0
	}

	mac := hmac.New(sha256.New, key)
	mac.Write(work)
	digest := mac.Sum(nil)

	return hmac.Equal(received, digest[:signSignatureLength])
}

// signMessageCMAC signs a marshalled SMB2 message in place using AES-128-CMAC,
// the signing algorithm for the SMB 3.0, 3.0.2, and 3.1.1 dialects when no
// alternative signing algorithm is negotiated (MS-SMB2 3.1.4.1). The procedure
// mirrors the 2.x path — set SMB2_FLAGS_SIGNED, zero the Signature field, then
// compute the MAC over the whole message — but uses AES-128-CMAC (RFC 4493)
// keyed with the 16-byte SigningKey and takes the full 16-byte tag.
func signMessageCMAC(key, message []byte) {
	if len(message) < header.SMB2_HEADER_SIZE {
		return
	}

	f := binary.LittleEndian.Uint32(message[signFlagsOffset : signFlagsOffset+4])
	f |= uint32(flags.SMB2_FLAGS_SIGNED)
	binary.LittleEndian.PutUint32(message[signFlagsOffset:signFlagsOffset+4], f)

	for i := 0; i < signSignatureLength; i++ {
		message[signSignatureOffset+i] = 0
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	mac := cmac.New(block)
	mac.Write(message)
	digest := mac.Sum(nil)
	copy(message[signSignatureOffset:signSignatureOffset+signSignatureLength], digest[:signSignatureLength])
}

// verifySignatureCMAC recomputes the AES-128-CMAC signature of a received SMB2
// message and compares it in constant time to the Signature it carries. The
// computation is done on a copy so the caller's buffer is left intact.
func verifySignatureCMAC(key, message []byte) bool {
	if len(message) < header.SMB2_HEADER_SIZE {
		return false
	}

	received := make([]byte, signSignatureLength)
	copy(received, message[signSignatureOffset:signSignatureOffset+signSignatureLength])

	work := make([]byte, len(message))
	copy(work, message)
	for i := 0; i < signSignatureLength; i++ {
		work[signSignatureOffset+i] = 0
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return false
	}
	mac := cmac.New(block)
	mac.Write(work)
	digest := mac.Sum(nil)

	return hmac.Equal(received, digest[:signSignatureLength])
}

// gmacNonce builds the 12-byte AES-GMAC nonce for signing (MS-SMB2 3.1.4.1):
// bytes 0–7 are the message's MessageId in little-endian, bytes 8–9 are zero,
// and bytes 10–11 are the direction (0x0001 for client-to-server requests,
// 0x0002 for server-to-client responses). The SMB2_FLAGS_SERVER_TO_REDIR flag
// in the header Flags distinguishes the two.
func gmacNonce(message []byte) []byte {
	const messageIdOffset = 24 // within the 64-byte SMB2 header
	nonce := make([]byte, 12)
	copy(nonce[0:8], message[messageIdOffset:messageIdOffset+8])
	f := binary.LittleEndian.Uint32(message[signFlagsOffset : signFlagsOffset+4])
	if f&uint32(flags.SMB2_FLAGS_SERVER_TO_REDIR) != 0 {
		binary.LittleEndian.PutUint16(nonce[10:12], 0x0002)
	} else {
		binary.LittleEndian.PutUint16(nonce[10:12], 0x0001)
	}
	return nonce
}

// signMessageGMAC signs a marshalled SMB2 message in place using AES-GMAC, the
// signing algorithm negotiated via SMB2_SIGNING_CAPABILITIES when AES-GMAC is
// selected (MS-SMB2 3.1.4.1). GMAC is GCM with an empty plaintext: the entire
// message is the additional authenticated data (AAD) and the 16-byte
// authentication tag is the signature.
func signMessageGMAC(key, message []byte) {
	if len(message) < header.SMB2_HEADER_SIZE {
		return
	}

	f := binary.LittleEndian.Uint32(message[signFlagsOffset : signFlagsOffset+4])
	f |= uint32(flags.SMB2_FLAGS_SIGNED)
	binary.LittleEndian.PutUint32(message[signFlagsOffset:signFlagsOffset+4], f)

	for i := 0; i < signSignatureLength; i++ {
		message[signSignatureOffset+i] = 0
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, 12)
	if err != nil {
		return
	}
	nonce := gmacNonce(message)
	tag := gcm.Seal(nil, nonce, nil, message)
	copy(message[signSignatureOffset:signSignatureOffset+signSignatureLength], tag[:signSignatureLength])
}

// verifySignatureGMAC recomputes the AES-GMAC signature of a received SMB2
// message and compares it in constant time to the Signature it carries.
func verifySignatureGMAC(key, message []byte) bool {
	if len(message) < header.SMB2_HEADER_SIZE {
		return false
	}

	received := make([]byte, signSignatureLength)
	copy(received, message[signSignatureOffset:signSignatureOffset+signSignatureLength])

	work := make([]byte, len(message))
	copy(work, message)
	for i := 0; i < signSignatureLength; i++ {
		work[signSignatureOffset+i] = 0
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return false
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, 12)
	if err != nil {
		return false
	}
	nonce := gmacNonce(work)
	tag := gcm.Seal(nil, nonce, nil, work)

	return hmac.Equal(received, tag[:signSignatureLength])
}

// signMessageForDialect signs a message in place with the algorithm appropriate
// for the negotiated dialect and signing algorithm. When the SMB 3.1.1
// SMB2_SIGNING_CAPABILITIES context negotiated a specific algorithm, that
// algorithm is used; otherwise the dialect default applies (AES-128-CMAC for
// SMB 3.x, HMAC-SHA256 for 2.x).
func signMessageForDialect(dialect dialects.Dialect, signingAlg int, key, message []byte) {
	if signingAlg == commands.SMB2_SIGNING_ALG_AES_GMAC {
		signMessageGMAC(key, message)
		return
	}
	if isSMB3Dialect(dialect) {
		signMessageCMAC(key, message)
		return
	}
	signMessage(key, message)
}

// verifySignatureForDialect verifies a message signature with the algorithm
// appropriate for the negotiated dialect and signing algorithm.
func verifySignatureForDialect(dialect dialects.Dialect, signingAlg int, key, message []byte) bool {
	if signingAlg == commands.SMB2_SIGNING_ALG_AES_GMAC {
		return verifySignatureGMAC(key, message)
	}
	if isSMB3Dialect(dialect) {
		return verifySignatureCMAC(key, message)
	}
	return verifySignature(key, message)
}
