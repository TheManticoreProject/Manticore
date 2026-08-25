// Package signing implements SMB 1.0 message signing.
//
// The primitives are direction-agnostic, because signing is symmetric: each side
// signs what it sends and verifies what it receives with the same key and the
// same digest. Only the sequence numbers differ, and which of them a party uses
// is the one thing that distinguishes a client from a server here.
//
// # The MAC key
//
// The key is the ExportedSessionKey derived from the authentication exchange,
// used as-is. [MS-CIFS] 3.1.5.1 also defines a MACKey of SessionKey ||
// NTLMResponse for the paths that do not derive a session key, but neither side
// of this implementation takes such a path: extended session security is required
// precisely so a key exists to sign with.
//
// # Sequence numbers
//
// Signing is bootstrapped by the authentication exchange and then advances in
// pairs. A request carries an even sequence number and the response to it carries
// the odd number above, so the AUTHENTICATE request is signed at 0, its response
// at 1, and the next request at 2.
//
// A client therefore consumes a number when it sends and expects the next one
// back; a server verifies at the number it expects and signs its reply at the one
// above. ResponseSequenceNumber and NextRequestSequenceNumber name the two steps
// so both sides derive them the same way rather than each spelling out an
// increment.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/59ec6b57-7e1b-4a0e-bbdf-1a4f4e0e1d75
package signing

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
)

// SecuritySignatureOffset is the byte offset, from the start of the SMB header, of
// the 8-byte SecurityFeatures field that carries the signature:
// Protocol(4) + Command(1) + Status(4) + Flags(1) + Flags2(2) + PIDHigh(2) = 14.
const SecuritySignatureOffset = 14

// SecuritySignatureSize is the size, in bytes, of the SecuritySignature field.
const SecuritySignatureSize = 8

// Signature is an SMB 1.0 message signature.
type Signature = [SecuritySignatureSize]byte

// Compute returns the signature for a fully marshalled SMB message.
//
// The signature is the first 8 bytes of MD5(MACKey || message), where the
// message's SecuritySignature field has first been replaced by the 32-bit
// sequence number in little-endian order followed by four zero bytes. The input
// slice is not modified.
//
// Parameters:
//   - macKey: the MAC key, which is the exported session key
//   - message: the marshalled SMB message, header included
//   - sequenceNumber: the sequence number this message is signed at
//
// Returns:
//   - The signature
func Compute(macKey, message []byte, sequenceNumber uint32) Signature {
	// Work on a copy so the caller's buffer is untouched while the sequence
	// number stands in for the signature during the digest.
	buffer := make([]byte, len(message))
	copy(buffer, message)

	if len(buffer) >= SecuritySignatureOffset+SecuritySignatureSize {
		binary.LittleEndian.PutUint32(buffer[SecuritySignatureOffset:SecuritySignatureOffset+4], sequenceNumber)
		for i := SecuritySignatureOffset + 4; i < SecuritySignatureOffset+SecuritySignatureSize; i++ {
			buffer[i] = 0x00
		}
	}

	digest := md5.New()
	digest.Write(macKey)
	digest.Write(buffer)
	sum := digest.Sum(nil)

	var signature Signature
	copy(signature[:], sum[:SecuritySignatureSize])
	return signature
}

// Sign computes the signature for a marshalled message and writes it into the
// message's SecuritySignature field in place.
//
// A message too short to carry a signature is left alone: there is no field to
// write into, and the caller's own framing will have already rejected it.
//
// Parameters:
//   - macKey: the MAC key
//   - message: the marshalled SMB message, modified in place
//   - sequenceNumber: the sequence number to sign at
func Sign(macKey, message []byte, sequenceNumber uint32) {
	if len(message) < SecuritySignatureOffset+SecuritySignatureSize {
		return
	}
	signature := Compute(macKey, message, sequenceNumber)
	copy(message[SecuritySignatureOffset:SecuritySignatureOffset+SecuritySignatureSize], signature[:])
}

// Verify reports whether a marshalled message carries a valid signature for the
// given key and sequence number. The comparison is constant-time.
//
// Parameters:
//   - macKey: the MAC key
//   - message: the marshalled SMB message as received
//   - sequenceNumber: the sequence number the message is expected to be signed at
//
// Returns:
//   - true when the signature is valid
func Verify(macKey, message []byte, sequenceNumber uint32) bool {
	if len(message) < SecuritySignatureOffset+SecuritySignatureSize {
		return false
	}

	received := make([]byte, SecuritySignatureSize)
	copy(received, message[SecuritySignatureOffset:SecuritySignatureOffset+SecuritySignatureSize])

	expected := Compute(macKey, message, sequenceNumber)
	return hmac.Equal(received, expected[:])
}

// ResponseSequenceNumber returns the sequence number the response to a request
// must carry, given the number the request was signed at.
//
// Parameters:
//   - requestSequenceNumber: the number the request was signed at
//
// Returns:
//   - The number its response is signed at
func ResponseSequenceNumber(requestSequenceNumber uint32) uint32 {
	return requestSequenceNumber + 1
}

// NextRequestSequenceNumber returns the sequence number the following request
// must carry, given the number the current one was signed at. A request and its
// response consume one number each, so the next request skips both.
//
// Parameters:
//   - requestSequenceNumber: the number the current request was signed at
//
// Returns:
//   - The number the next request is signed at
func NextRequestSequenceNumber(requestSequenceNumber uint32) uint32 {
	return requestSequenceNumber + 2
}
