package parameters

import (
	"encoding/binary"
	"fmt"
)

// SMBParameters represents the parameters structure for SMB packets
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/c87a9a6e-e318-44d3-85e1-82398f8dc9f5
type Parameters struct {
	// WordCount (1 byte): The size, in two-byte words, of the Words field. This field can be zero, indicating that the Words
	// field is empty. Note that the size of this field is one byte and comes after the fixed 32-byte SMB Header (section 2.2.3.1),
	// which causes the Words field to be unaligned.
	WordCount uint8

	// Words (variable): The message-specific parameters structure. The size of this field MUST be (2 x WordCount) bytes.
	// If WordCount is 0x00, this field is not included.
	Words []uint16
}

// NewParameters creates a new SMB Parameters structure with default values.
// It initializes an empty Parameters structure with WordCount set to 0
// and an empty Words slice.
//
// Returns:
// - A pointer to the initialized Parameters structure
func NewParameters() *Parameters {
	return &Parameters{
		WordCount: 0,
		Words:     []uint16{},
	}
}

// AddWord appends a new 16-bit word to the Parameters structure.
// This method adds the provided word to the Words slice and
// updates the WordCount field to reflect the new count.
//
// Parameters:
// - word: The 16-bit word to add to the Parameters structure
func (p *Parameters) AddWord(word uint16) {
	p.Words = append(p.Words, word)

	// Update the WordCount field to reflect the new count
	p.WordCount = uint8(len(p.Words) * 2)
}

// GetBytesStream returns the binary representation of the Parameters structure.
// This method converts the Parameters structure into a byte slice according to
// the SMB protocol format. It first writes the WordCount byte, followed by each
// word in big-endian byte order (high byte first, then low byte).
//
// Unlike Marshal, this method does not return an error and is designed for
// direct inclusion in message construction.
//
// Returns:
// - A byte slice containing the binary representation of the Parameters structure
func (p *Parameters) GetBytesStream() []byte {
	bytes := []byte{}
	for _, word := range p.Words {
		bytes = append(bytes, uint8(word>>8), uint8(word&0xFF))
	}
	return bytes
}

// GetBytes returns the bytes of the Parameters structure
//
// This function returns the bytes of the Parameters structure. It returns the Bytes field.
func (p *Parameters) GetBytes() []byte {
	return p.GetBytesStream()
}

// Size returns the size of the Data structure
//
// This function returns the size of the Data structure. It returns the ByteCount field.
func (p *Parameters) Size() uint16 {
	return uint16(len(p.Words))
}

// AddWordsFromBytesStream converts a byte stream to a slice of uint16 parameters
// It reads the byte stream in chunks of 2 bytes and converts each chunk to a uint16 word
// If the byte stream length is odd, the last byte will be padded with zero
func (p *Parameters) AddWordsFromBytesStream(bytesStream []byte) {
	parameters := []uint16{}

	// Process bytes in chunks of 2 (16 bits per word)
	for i := 0; i < len(bytesStream); i += 2 {
		if i+1 < len(bytesStream) {
			// Convert two bytes to a uint16 (little-endian)
			word := uint16(bytesStream[i])<<8 | uint16(bytesStream[i+1])
			parameters = append(parameters, word)
		} else {
			// Handle odd length - pad the last byte with zero
			word := uint16(bytesStream[i])
			parameters = append(parameters, word)
		}
	}

	p.Words = append(p.Words, parameters...)
	p.WordCount = uint8(len(p.Words))
}

// Marshal serializes the Parameters structure into a byte slice.
// This method converts the Parameters structure into its binary representation
// according to the SMB protocol format. It first writes the WordCount byte,
// followed by each word in little-endian byte order.
//
// Returns:
// - A byte slice containing the marshalled Parameters structure
// - An error if marshalling fails, or nil if successful
func (p *Parameters) Marshal() ([]byte, error) {
	marshalled := []byte{}

	if p.WordCount != uint8(len(p.Words)) {
		return nil, fmt.Errorf("WordCount does not match the number of words in the Words slice")
	}

	// WordCount (1 byte): The size, in two-byte words, of the Words field. This field can be zero, indicating that the Words
	// field is empty. Note that the size of this field is one byte and comes after the fixed 32-byte SMB Header (section 2.2.3.1),
	// which causes the Words field to be unaligned.
	marshalled = append(marshalled, p.WordCount)

	// Words (variable): The message-specific parameters structure. The size of this field MUST be (2 x WordCount) bytes.
	// If WordCount is 0x00, this field is not included.
	if p.WordCount > 0 {
		for _, word := range p.Words {
			buf2 := make([]byte, 2)
			binary.BigEndian.PutUint16(buf2, word)
			marshalled = append(marshalled, buf2...)
		}
	}

	return marshalled, nil
}

// Unmarshal deserializes a byte slice into the Parameters structure.
// This method reads the binary representation of the Parameters structure
// from the input byte slice according to the SMB protocol format. It first
// reads the WordCount byte, then reads each word in little-endian byte order.
//
// Parameters:
// - data: The byte slice containing the serialized Parameters structure
//
// Returns:
// - int: The number of bytes read from the input byte slice
// - error: Any error encountered during deserialization, or nil if successful
func (p *Parameters) Unmarshal(data []byte) (int, error) {
	bytesRead := 0

	if len(data) == 0 {
		return bytesRead, fmt.Errorf("data is empty")
	}

	fmt.Printf("Parameters: %v\n", p)

	p.WordCount = uint8(data[0])
	bytesRead += 1

	data = data[bytesRead:]

	if p.WordCount > 0 {
		// Each word is 2 bytes
		if len(data) < int(p.WordCount)*2 {
			return bytesRead, fmt.Errorf("data too short to unmarshal SMB parameters")
		}

		p.Words = make([]uint16, p.WordCount)
		for i := 0; i < int(p.WordCount); i++ {
			p.Words[i] = binary.BigEndian.Uint16(data[i*2 : 2+i*2])
		}
		bytesRead += int(p.WordCount) * 2
	} else {
		p.Words = []uint16{}
	}

	return bytesRead, nil
}
