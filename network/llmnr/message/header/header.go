package header

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/llmnr/errors"
)

// Size of LLMNR message header in bytes
const HeaderSize = 12

// Header represents the LLMNR message header.
//
// The header contains essential information about the LLMNR message, including the message ID, flags, and counts of
// various sections such as questions, answers, authority records, and additional records.
//
// Fields:
//   - Identifier: A 16-bit identifier assigned by the program that generates any kind of query. This identifier is copied
//     to the corresponding reply and can be used by the requester to match up replies to outstanding queries.
//   - Flags: A 16-bit field containing various flags that control the message flow and interpretation. These flags
//     include the Query/Response flag (QR), Operation code (OP), Conflict flag (C), Truncation flag (TC), and Tentative flag (T).
//   - QDCount: An unsigned 16-bit integer specifying the number of entries in the question section of the message.
//   - ANCount: An unsigned 16-bit integer specifying the number of resource records in the answer section of the message.
//   - NSCount: An unsigned 16-bit integer specifying the number of name server resource records in the authority records section of the message.
//   - ARCount: An unsigned 16-bit integer specifying the number of resource records in the additional records section of the message.
//
// Usage example:
//
//	header := Header{
//	    Identifier: 12345,
//	    Flags:   FlagQR,
//	    QDCount: 1,
//	    ANCount: 0,
//	    NSCount: 0,
//	    ARCount: 0,
//	}
type Header struct {
	Identifier uint16 `json:"identifier"`
	Flags      Flags  `json:"flags"`
	QDCount    uint16 `json:"qd_count"` // Question count
	ANCount    uint16 `json:"an_count"` // Answer count
	NSCount    uint16 `json:"ns_count"` // Authority count
	ARCount    uint16 `json:"ar_count"` // Additional count
}

// Marshal encodes the Header into a 12-byte big-endian representation.
func (h *Header) Marshal() ([]byte, error) {
	marshalledData := make([]byte, 12)

	binary.BigEndian.PutUint16(marshalledData[0:2], h.Identifier)

	flags, err := h.Flags.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flags: %w", err)
	}
	copy(marshalledData[2:4], flags)

	binary.BigEndian.PutUint16(marshalledData[4:6], h.QDCount)

	binary.BigEndian.PutUint16(marshalledData[6:8], h.ANCount)

	binary.BigEndian.PutUint16(marshalledData[8:10], h.NSCount)

	binary.BigEndian.PutUint16(marshalledData[10:12], h.ARCount)

	return marshalledData, nil
}

// Unmarshal decodes a 12-byte big-endian representation into the Header receiver.
// It returns an error if the input slice is not exactly 12 bytes.
func (h *Header) Unmarshal(data []byte) (int, error) {
	if len(data) != 12 {
		return 0, fmt.Errorf("invalid length: got %d bytes, want 12 bytes: %w", len(data), errors.ErrInvalidHeader)
	}

	bytesRead := 0
	h.Identifier = binary.BigEndian.Uint16(data[bytesRead : bytesRead+2])
	bytesRead += 2

	bytesReadFlags, err := h.Flags.Unmarshal(data[bytesRead : bytesRead+2])
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal flags: %w: %w", err, errors.ErrInvalidHeader)
	}
	bytesRead += bytesReadFlags

	h.QDCount = binary.BigEndian.Uint16(data[bytesRead : bytesRead+2])
	bytesRead += 2

	h.ANCount = binary.BigEndian.Uint16(data[bytesRead : bytesRead+2])
	bytesRead += 2

	h.NSCount = binary.BigEndian.Uint16(data[bytesRead : bytesRead+2])
	bytesRead += 2

	h.ARCount = binary.BigEndian.Uint16(data[bytesRead : bytesRead+2])
	bytesRead += 2

	return bytesRead, nil
}

// Describe prints a detailed description of the Header struct.
// Parameters:
// - indent: An integer value specifying the indentation level for the output.
func (h *Header) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)

	fmt.Printf("%s<Header>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mIdentifier\x1b[0m: 0x%04x (%d)\n", indentPrompt, h.Identifier, h.Identifier)
	fmt.Printf("%s │ \x1b[93mFlags\x1b[0m: 0x%04x (%s)\n", indentPrompt, uint16(h.Flags), h.Flags.String())
	fmt.Printf("%s │ \x1b[93mQDCount\x1b[0m: %d\n", indentPrompt, h.QDCount)
	fmt.Printf("%s │ \x1b[93mANCount\x1b[0m: %d\n", indentPrompt, h.ANCount)
	fmt.Printf("%s │ \x1b[93mNSCount\x1b[0m: %d\n", indentPrompt, h.NSCount)
	fmt.Printf("%s │ \x1b[93mARCount\x1b[0m: %d\n", indentPrompt, h.ARCount)
	fmt.Printf("%s └───\n", indentPrompt)
}
