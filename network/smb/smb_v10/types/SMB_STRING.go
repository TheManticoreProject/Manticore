package types

import (
	"encoding/binary"
	"fmt"
)

const (
	SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT             = 0x01
	SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING       = 0x02
	SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING_16BIT = 0x03
	SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING     = 0x04
	SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK                   = 0x05
)

// SMB_STRING
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/9189a82f-c1c0-4af9-818c-85050f7e5e66
type SMB_STRING struct {
	BufferFormat UCHAR
	Length       USHORT
	Buffer       []UCHAR
}

// NewSMB_STRING creates a new SMB_STRING structure
//
// Parameters:
// - buffer: A byte slice containing the serialized SMB_STRING structure
//
// Returns:
// - A pointer to the new SMB_STRING structure
func NewSMB_STRING(buffer []byte) *SMB_STRING {
	return &SMB_STRING{
		Length: USHORT(len(buffer)),
		Buffer: buffer,
	}
}

// SetBufferFormat sets the buffer format of the SMB_STRING structure
//
// Parameters:
// - bufferFormat: A byte to set the buffer format of the SMB_STRING structure to
func (s *SMB_STRING) SetBufferFormat(bufferFormat UCHAR) {
	s.BufferFormat = bufferFormat
}

// SetString sets the string of the SMB_STRING structure
//
// Parameters:
// - str: A string to set the SMB_STRING structure to
func (s *SMB_STRING) SetString(str string) {
	s.Buffer = []UCHAR(str)
	s.Length = USHORT(len(str))
}

// Marshal serializes the SMB_STRING structure into a byte slice.
// This method converts the SMB_STRING structure into its binary representation
// according to the SMB protocol format. It first writes the BufferFormat byte,
// followed by the Length and Buffer fields.
//
// Returns:
// - A byte slice containing the marshalled SMB_STRING structure
// - An error if marshalling fails, or nil if successful
func (s *SMB_STRING) Marshal() ([]byte, error) {
	buffer := []byte{}

	switch s.BufferFormat {
	case 0x01:
		// A two-byte USHORT value indicating the length of the data buffer. The data buffer follows immediately after the length field.
		buffer = append(buffer, s.BufferFormat)

		buf2 := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf2, uint16(s.Length+1))
		buffer = append(buffer, buf2...)

		buffer = append(buffer, s.Buffer...)
		buffer = append(buffer, 0x00)

	case 0x02:
		// A null-terminated OEM_STRING.
		// This format code is used only in the SMB_COM_NEGOTIATE (section 2.2.4.52) command to identify SMB dialect strings.
		buffer = append(buffer, s.BufferFormat)

		buffer = append(buffer, s.Buffer...)
		buffer = append(buffer, 0x00)

	case 0x03:
		// A null-terminated OEM_STRING.
		buffer = append(buffer, s.BufferFormat)

		buf2 := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf2, uint16(s.Length+1))
		buffer = append(buffer, buf2...)

		buffer = append(buffer, s.Buffer...)
		buffer = append(buffer, 0x00)

	case 0x04:
		// This field MUST be 0x04, which indicates that a null-terminated ASCII string follows.
		// In the NT LAN Manager dialect, the string is of type SMB_STRING unless otherwise specified.
		buffer = append(buffer, s.BufferFormat)

		buffer = append(buffer, s.Buffer...)
		buffer = append(buffer, 0x00)

	case 0x05:
		// This field MUST be 0x05, which indicates that a variable block follows.
		buffer = append(buffer, s.BufferFormat)

		buf2 := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf2, uint16(s.Length+1))
		buffer = append(buffer, buf2...)

		buffer = append(buffer, s.Buffer...)
		buffer = append(buffer, 0x00)

	default:
		return nil, fmt.Errorf("invalid buffer format: %d", s.BufferFormat)
	}

	return buffer, nil
}

// Unmarshal deserializes the SMB_STRING structure from a byte slice.
// This method extracts the BufferFormat, Length, and Buffer fields from the input byte slice
// and populates the SMB_STRING structure with the extracted values.
//
// Parameters:
// - buffer: A byte slice containing the serialized SMB_STRING structure
//
// Returns:
// - An error if the unmarshalling process fails, or nil if successful
func (s *SMB_STRING) Unmarshal(buffer []byte) (int, error) {
	if len(buffer) < 1 {
		return 0, fmt.Errorf("buffer too short to unmarshal SMB_STRING")
	}

	s.BufferFormat = buffer[0]

	// Handle different buffer formats
	switch s.BufferFormat {
	case 0x01:
		// Variable block with 16-bit length
		if len(buffer) < 3 {
			return 0, fmt.Errorf("buffer too short for format 0x%02x", s.BufferFormat)
		}

		s.Length = USHORT(binary.LittleEndian.Uint16(buffer[1:3]))
		if len(buffer) < int(s.Length)+3 {
			return 0, fmt.Errorf("buffer too short for specified length")
		}

		s.Buffer = make([]UCHAR, s.Length)
		copy(s.Buffer, buffer[3:3+s.Length])

		return int(s.Length) + 3, nil

	case 0x02:
		// Null-terminated string for dialect negotiation
		// Find the null terminator
		nullPos := -1
		for i := 1; i < len(buffer); i++ {
			if buffer[i] == 0x00 {
				nullPos = i
				break
			}
		}
		if nullPos == -1 {
			return 0, fmt.Errorf("no null terminator found for format 0x%02x", s.BufferFormat)
		}

		s.Buffer = make([]UCHAR, nullPos-1)
		copy(s.Buffer, buffer[1:nullPos])

		s.Length = USHORT(len(s.Buffer))

		return nullPos + 1, nil

	case 0x03:
		// Variable block with 16-bit length
		if len(buffer) < 3 {
			return 0, fmt.Errorf("buffer too short for format 0x%02x", s.BufferFormat)
		}

		s.Length = USHORT(binary.LittleEndian.Uint16(buffer[1:3]))
		if len(buffer) < int(s.Length)+3 {
			return 0, fmt.Errorf("buffer too short for specified length")
		}

		s.Buffer = make([]UCHAR, s.Length)
		copy(s.Buffer, buffer[3:3+s.Length])

		return int(s.Length) + 3, nil

	case 0x04:
		// This field MUST be 0x04, which indicates that a null-terminated ASCII string follows.
		nullPos := -1
		for i := 1; i < len(buffer); i++ {
			if buffer[i] == 0x00 {
				nullPos = i
				break
			}
		}
		if nullPos == -1 {
			return 0, fmt.Errorf("no null terminator found for format 0x%02x", s.BufferFormat)
		}

		s.Buffer = make([]UCHAR, nullPos-1)
		copy(s.Buffer, buffer[1:nullPos])

		s.Length = USHORT(len(s.Buffer))

		return nullPos + 1, nil

	case 0x05:
		// This field MUST be 0x05, which indicates that a variable block follows.
		s.Length = USHORT(binary.LittleEndian.Uint16(buffer[1:3]))

		s.Buffer = make([]UCHAR, s.Length)
		copy(s.Buffer, buffer[3:3+s.Length])

		return int(s.Length) + 3, nil

	default:
		return 0, fmt.Errorf("invalid buffer format: 0x%02x", s.BufferFormat)
	}
}
