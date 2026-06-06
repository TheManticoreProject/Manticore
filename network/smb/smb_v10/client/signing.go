package client

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"fmt"
)

// Server signing policy states, recorded in Server.SigningState after negotiation.
const (
	SigningStateDisabled = "disabled"
	SigningStateEnabled  = "enabled"
	SigningStateRequired = "required"
)

// securitySignatureOffset is the byte offset, from the start of the SMB header, of
// the 8-byte SecurityFeatures field that carries the message signature:
// Protocol(4) + Command(1) + Status(4) + Flags(1) + Flags2(2) + PIDHigh(2) = 14.
const securitySignatureOffset = 14

// securitySignatureSize is the size, in bytes, of the SecuritySignature field.
const securitySignatureSize = 8

// computeSMBSignature computes the 8-byte SMB v1 message signature for a fully
// marshalled SMB message, per [MS-CIFS] 3.1.5.1. The signature is the first 8 bytes
// of MD5(MACKey || SMBMessage), where the message's SecuritySignature field has
// first been set to the 32-bit sequence number (little-endian) followed by four
// zero bytes. The input slice is not modified.
func computeSMBSignature(macKey, message []byte, sequenceNumber uint32) [securitySignatureSize]byte {
	// Work on a copy so the caller's buffer is untouched while we substitute the
	// sequence number into the signature field for the digest computation.
	buf := make([]byte, len(message))
	copy(buf, message)

	if len(buf) >= securitySignatureOffset+securitySignatureSize {
		binary.LittleEndian.PutUint32(buf[securitySignatureOffset:securitySignatureOffset+4], sequenceNumber)
		for i := securitySignatureOffset + 4; i < securitySignatureOffset+securitySignatureSize; i++ {
			buf[i] = 0x00
		}
	}

	h := md5.New()
	h.Write(macKey)
	h.Write(buf)
	digest := h.Sum(nil)

	var signature [securitySignatureSize]byte
	copy(signature[:], digest[:securitySignatureSize])
	return signature
}

// signSMBMessage computes the SMB v1 signature for the marshalled message and
// writes it into the message's SecuritySignature field in place.
func signSMBMessage(macKey, message []byte, sequenceNumber uint32) {
	if len(message) < securitySignatureOffset+securitySignatureSize {
		return
	}
	signature := computeSMBSignature(macKey, message, sequenceNumber)
	copy(message[securitySignatureOffset:securitySignatureOffset+securitySignatureSize], signature[:])
}

// verifySMBSignature reports whether the marshalled message carries a valid SMB v1
// signature for the given MAC key and expected sequence number. The comparison is
// constant-time.
func verifySMBSignature(macKey, message []byte, sequenceNumber uint32) bool {
	if len(message) < securitySignatureOffset+securitySignatureSize {
		return false
	}
	received := make([]byte, securitySignatureSize)
	copy(received, message[securitySignatureOffset:securitySignatureOffset+securitySignatureSize])

	expected := computeSMBSignature(macKey, message, sequenceNumber)
	return hmac.Equal(received, expected[:])
}

// signOutgoing signs a marshalled request in place when signing is active for the
// connection, consuming the next send sequence number. It returns the sequence
// number expected on the matching response and whether the message was signed. When
// signing is inactive it leaves the message untouched and returns (0, false).
func (c *Client) signOutgoing(marshalled []byte) (uint32, bool) {
	if c.Connection == nil || !c.Connection.IsSigningActive {
		return 0, false
	}
	sequenceNumber := c.Connection.ClientNextSendSequenceNumber
	signSMBMessage(c.Connection.SigningSessionKey, marshalled, sequenceNumber)
	// The response carries the next sequence number; advance past both the request
	// and its response for the following exchange.
	c.Connection.ClientNextSendSequenceNumber += 2
	return sequenceNumber + 1, true
}

// verifyIncoming validates the signature of a received response against the given
// expected sequence number when signing is active. It is a no-op when signing is
// inactive.
func (c *Client) verifyIncoming(raw []byte, responseSequenceNumber uint32) error {
	if c.Connection == nil || !c.Connection.IsSigningActive {
		return nil
	}
	if !verifySMBSignature(c.Connection.SigningSessionKey, raw, responseSequenceNumber) {
		return fmt.Errorf("invalid SMB message signature on response (expected sequence %d)", responseSequenceNumber)
	}
	return nil
}
