package types

import (
	"encoding/binary"
	"fmt"
	"math"
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
		BufferFormat: 0x00,
		Length:       USHORT(len(buffer)),
		Buffer:       buffer,
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
func (s *SMB_STRING) SetString(str string) error {
	if len(str) > math.MaxUint16 {
		return fmt.Errorf("string too long")
	}

	s.Buffer = []UCHAR([]byte(str))
	s.Length = USHORT(len(str))

	return nil
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
	case SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT:
		// A two-byte USHORT value indicating the length of the data buffer. The data buffer follows immediately after the length field.
		buffer = append(buffer, s.BufferFormat)

		// Length of the data buffer
		buf2 := make([]byte, 2)
		if len(s.Buffer) > math.MaxUint16 {
			return nil, fmt.Errorf("string too long (length: %d)", len(s.Buffer))
		}
		s.Length = USHORT(len(s.Buffer))
		binary.LittleEndian.PutUint16(buf2, uint16(s.Length))
		buffer = append(buffer, buf2...)

		// Data buffer
		buffer = append(buffer, s.Buffer...)

	case SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING:
		// A null-terminated OEM_STRING.
		// This format code is used only in the SMB_COM_NEGOTIATE (section 2.2.4.52) command to identify SMB dialect strings.
		buffer = append(buffer, s.BufferFormat)

		// Data buffer
		buffer = append(buffer, s.Buffer...)
		buffer = append(buffer, 0x00)

	case SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING_16BIT:
		// A null-terminated OEM_STRING in UTF-16.
		buffer = append(buffer, s.BufferFormat)

		// Length of the data buffer
		buf2 := make([]byte, 2)
		if len(s.Buffer) > math.MaxUint16 {
			return nil, fmt.Errorf("string too long (length: %d)", len(s.Buffer))
		}
		s.Length = USHORT(len(s.Buffer))
		binary.LittleEndian.PutUint16(buf2, uint16(s.Length))
		buffer = append(buffer, buf2...)

		// Data buffer
		buffer = append(buffer, s.Buffer...)
		buffer = append(buffer, 0x00)

	case SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING:
		// This field MUST be 0x04, which indicates that a null-terminated ASCII string follows.
		// In the NT LAN Manager dialect, the string is of type SMB_STRING unless otherwise specified.
		buffer = append(buffer, s.BufferFormat)

		// Data buffer
		buffer = append(buffer, s.Buffer...)
		buffer = append(buffer, 0x00)

	case SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK:
		// This field MUST be 0x05, which indicates that a variable block follows.
		buffer = append(buffer, s.BufferFormat)

		// Length of the data buffer
		buf2 := make([]byte, 2)
		if len(s.Buffer) > math.MaxUint16 {
			return nil, fmt.Errorf("string too long (length: %d)", len(s.Buffer))
		}
		s.Length = USHORT(len(s.Buffer))
		binary.LittleEndian.PutUint16(buf2, uint16(s.Length))
		buffer = append(buffer, buf2...)

		// Data buffer
		buffer = append(buffer, s.Buffer...)

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
	case SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT:
		// Variable block with 16-bit length
		if len(buffer) < 3 {
			return 0, fmt.Errorf("buffer too short for format 0x%02x", s.BufferFormat)
		}

		// Length of the data buffer
		s.Length = USHORT(binary.LittleEndian.Uint16(buffer[1:3]))
		if len(buffer) < int(s.Length)+3 {
			return 0, fmt.Errorf("buffer too short for specified length")
		}

		// Data buffer
		s.Buffer = make([]UCHAR, s.Length)
		copy(s.Buffer, buffer[3:3+s.Length])

		return int(s.Length) + 3, nil

	case SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING:
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

		// Data buffer
		s.Buffer = make([]UCHAR, nullPos-1)
		copy(s.Buffer, buffer[1:nullPos])

		// Length of the data buffer
		s.Length = USHORT(len(s.Buffer))

		return nullPos + 1, nil

	case SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING_16BIT:
		// Variable block with 16-bit length
		if len(buffer) < 3 {
			return 0, fmt.Errorf("buffer too short for format 0x%02x", s.BufferFormat)
		}

		s.Length = USHORT(binary.LittleEndian.Uint16(buffer[1:3]))
		if len(buffer) < int(s.Length)+3 {
			return 0, fmt.Errorf("buffer too short for specified length")
		}

		// Data buffer
		// We ignore the null terminator here
		s.Buffer = make([]UCHAR, s.Length)
		copy(s.Buffer, buffer[3:3+s.Length])

		// We count the null terminator here because it is consumed
		bytesRead := int(s.Length) + 3 + 1

		return bytesRead, nil

	case SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING:
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

		// Data buffer
		s.Buffer = make([]UCHAR, nullPos-1)
		copy(s.Buffer, buffer[1:nullPos])

		// Length of the data buffer
		s.Length = USHORT(len(s.Buffer))

		return nullPos + 1, nil

	case SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK:
		// This field MUST be 0x05, which indicates that a variable block follows.
		if len(buffer) < 3 {
			return 0, fmt.Errorf("buffer too short for format 0x%02x", s.BufferFormat)
		}

		s.Length = USHORT(binary.LittleEndian.Uint16(buffer[1:3]))
		if len(buffer) < int(s.Length)+3 {
			return 0, fmt.Errorf("buffer too short for specified length")
		}

		// Data buffer
		s.Buffer = make([]UCHAR, s.Length)
		copy(s.Buffer, buffer[3:3+s.Length])

		return int(s.Length) + 3, nil

	default:
		return 0, fmt.Errorf("invalid buffer format: 0x%02x", s.BufferFormat)
	}
}

// UnmarshalWithEncoding deserializes an SMB_STRING whose characters are as wide as
// the enclosing message declared.
//
// The null-terminated formats end at their first null CHARACTER, not their first
// null byte. In a Unicode message a character is two bytes, so "\newdir" begins
// 5C 00 and a scan for a single null byte ends the string after one byte — leaving
// a buffer that is half a character long. Every caller then either truncates the
// name to nothing or, decoding an odd number of bytes as UTF-16, reads past the
// end of it.
//
// The formats that carry an explicit length are unaffected: their length already
// says where the string ends, whatever its characters are.
//
// Parameters:
//   - buffer: the serialized SMB_STRING, starting at its BufferFormat byte
//   - unicode: whether the enclosing message set SMB_FLAGS2_UNICODE
//
// Returns:
//   - The number of bytes consumed
//   - An error if the buffer cannot be parsed
func (s *SMB_STRING) UnmarshalWithEncoding(buffer []byte, unicode bool) (int, error) {
	// An aligned string is the common case: the first string in a command's data
	// block lands on an even offset from the header without help, because the
	// block's own offset and the format byte together make it so.
	return s.UnmarshalWithEncodingAt(buffer, unicode, 0)
}

// UnmarshalWithEncodingAt is UnmarshalWithEncoding told where the string sits.
//
// stringStart is the offset, from the start of the SMB header, of the first
// character — that is, of the byte after the BufferFormat. [MS-CIFS] requires a
// Unicode string to begin on a 2-byte boundary measured from the header, so when
// that offset is odd a padding byte stands between the format byte and the
// characters.
//
// This matters for the second string of a command that carries two. The first one
// is aligned by construction, but the first string's length then decides where the
// second begins, so half the time a pad byte appears there — and a decoder that
// does not expect it reads every character of the second name straddling two
// others, which yields a plausible-looking name made of the wrong bytes rather
// than an error.
//
// Passing 0 means "already aligned", which is what the single-string commands are.
//
// Parameters:
//   - buffer: the serialized SMB_STRING, starting at its BufferFormat byte
//   - unicode: whether the enclosing message set SMB_FLAGS2_UNICODE
//   - stringStart: the header-relative offset of the byte after the format byte
//
// Returns:
//   - The number of bytes consumed, the padding byte included
//   - An error if the buffer cannot be parsed
func (s *SMB_STRING) UnmarshalWithEncodingAt(buffer []byte, unicode bool, stringStart int) (int, error) {
	if !unicode {
		return s.Unmarshal(buffer)
	}
	if len(buffer) < 1 {
		return 0, fmt.Errorf("buffer too short to unmarshal SMB_STRING")
	}

	switch buffer[0] {
	case SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING,
		SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING:
		s.BufferFormat = buffer[0]

		// Where the characters actually begin, once any alignment byte is passed.
		start := 1
		if stringStart%2 != 0 {
			start++
		}
		if start > len(buffer) {
			return 0, fmt.Errorf("buffer too short for an aligned SMB_STRING")
		}

		// The characters are two bytes wide, so the pairs to test step two at a
		// time from there. An odd offset cannot end the string: the high half of
		// one character and the low half of the next can both be zero without the
		// string having ended.
		for index := start; index+1 < len(buffer); index += 2 {
			if buffer[index] == 0x00 && buffer[index+1] == 0x00 {
				s.Buffer = make([]UCHAR, index-start)
				copy(s.Buffer, buffer[start:index])
				s.Length = USHORT(len(s.Buffer))
				return index + 2, nil
			}
		}
		return 0, fmt.Errorf("no null terminator found for format 0x%02x", buffer[0])
	}

	// Every other format states its own length, so the byte-oriented reader is
	// already correct for it.
	return s.Unmarshal(buffer)
}

// MarshalWithEncoding serializes an SMB_STRING whose characters are as wide as the
// enclosing message declared.
//
// It is the counterpart of UnmarshalWithEncoding, and exists for the same reason:
// the null-terminated formats end at a null CHARACTER. A two-byte character needs
// a two-byte terminator, and emitting a single null byte after a Unicode string
// leaves the field with no terminator at all — the reader consumes the low half of
// the null as the high half of the last character and runs on into whatever
// follows.
//
// Parameters:
//   - unicode: whether the enclosing message set SMB_FLAGS2_UNICODE
//
// Returns:
//   - The serialized SMB_STRING
//   - An error if it cannot be serialized
func (s *SMB_STRING) MarshalWithEncoding(unicode bool) ([]byte, error) {
	return s.MarshalWithEncodingAt(unicode, 0)
}

// MarshalWithEncodingAt is MarshalWithEncoding told where the string will sit, so
// it can emit the alignment byte that a Unicode string beginning on an odd
// header-relative offset requires.
//
// stringStart is the offset, from the start of the SMB header, that the first
// character would occupy without padding — that is, of the byte after the
// BufferFormat. Passing 0 means "already aligned".
//
// Parameters:
//   - unicode: whether the enclosing message set SMB_FLAGS2_UNICODE
//   - stringStart: the header-relative offset of the byte after the format byte
//
// Returns:
//   - The serialized SMB_STRING, alignment byte included
//   - An error if it cannot be serialized
func (s *SMB_STRING) MarshalWithEncodingAt(unicode bool, stringStart int) ([]byte, error) {
	if !unicode {
		return s.Marshal()
	}

	switch s.BufferFormat {
	case SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING,
		SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING:
		buffer := []byte{s.BufferFormat}
		if stringStart%2 != 0 {
			buffer = append(buffer, 0x00)
		}
		buffer = append(buffer, s.Buffer...)
		// The two-byte terminator that a two-byte character needs.
		buffer = append(buffer, 0x00, 0x00)
		return buffer, nil
	}

	// Every other format states its own length, so the byte-oriented writer is
	// already correct for it.
	return s.Marshal()
}
