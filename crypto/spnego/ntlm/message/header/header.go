package header

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/types"
)

// NTLM signature
var NTLM_SIGNATURE = [8]byte{'N', 'T', 'L', 'M', 'S', 'S', 'P', 0}

// Header is the header of an NTLM message
type Header struct {
	// Signature (8 bytes): An 8-byte character array that MUST contain the ASCII string ('N', 'T', 'L', 'M', 'S', 'S', 'P', '\0').
	Signature [8]byte

	// MessageType (4 bytes): A 32-bit unsigned integer that indicates the message type.
	MessageType types.MessageType
}

// Marshal serializes the Header into a byte slice
func (m *Header) Marshal() ([]byte, error) {
	marshalledData := []byte{}

	// Write signature
	marshalledData = append(marshalledData, m.Signature[:]...)

	// Write message type
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(m.MessageType))
	marshalledData = append(marshalledData, buf...)

	return marshalledData, nil
}

// Unmarshal deserializes a byte slice into a Header
//
// The signature is checked rather than merely copied. Everything that follows it
// is read at fixed offsets and at offsets the message itself declares, so a buffer
// that is not an NTLM message at all is otherwise parsed as one: the fields come
// out as whatever the bytes happened to say, and the caller has no way to tell.
// Refusing here means a parse failure is reported where it happens.
func (m *Header) Unmarshal(data []byte) (int, error) {
	if len(data) < 12 {
		return 0, fmt.Errorf("data too short to be a valid Header")
	}

	copy(m.Signature[:], data[:8])
	if m.Signature != NTLM_SIGNATURE {
		return 0, fmt.Errorf("data does not begin with the NTLMSSP signature (got % x)", m.Signature)
	}

	m.MessageType = types.MessageType(binary.LittleEndian.Uint32(data[8:12]))

	return 12, nil
}

// Expect reports whether the header names the message type the caller is parsing.
//
// The three NTLM messages have different layouts and no shared discriminator
// beyond this field, so parsing one as another produces a message populated from
// the wrong offsets rather than an error.
func (m *Header) Expect(messageType types.MessageType) error {
	if m.MessageType != messageType {
		return fmt.Errorf("message is a %s, not a %s", m.MessageType, messageType)
	}
	return nil
}

// GetType returns the message type
func (m *Header) GetType() types.MessageType {
	return m.MessageType
}

// SetType sets the message type
func (m *Header) SetType(messageType types.MessageType) {
	m.MessageType = messageType
}

// GetSignature returns the signature
func (m *Header) GetSignature() [8]byte {
	return m.Signature
}

// SetSignature sets the signature
func (m *Header) SetSignature(signature [8]byte) {
	m.Signature = signature
}
