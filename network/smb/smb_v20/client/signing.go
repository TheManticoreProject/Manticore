package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"

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
