package domain_name

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/llmnr/constants"
	"github.com/TheManticoreProject/Manticore/network/llmnr/errors"
)

type DomainName string

func (d *DomainName) Validate() error {
	if len(*d) > constants.MaxDomainLength {
		return errors.ErrNameTooLong
	}

	labels := strings.Split(string(*d), ".")
	for _, label := range labels {
		if len(label) > constants.MaxLabelLength {
			return errors.ErrLabelTooLong
		}
	}
	return nil
}

func (d *DomainName) Marshal() ([]byte, error) {
	return EncodeDomainName(string(*d))
}

// Unmarshal decodes a domain name from data starting at offset 0 and sets the receiver.
// It returns the number of bytes consumed from data.
//
// This entry point treats data as if the name started at the very beginning of
// the message, so it can only resolve compression pointers that reference bytes
// present in data. When the name lives inside a larger message and may contain
// 0xC0 compression pointers that reference earlier bytes, callers must use
// UnmarshalFromMessage instead so the pointers resolve against the message origin.
func (d *DomainName) Unmarshal(data []byte) (int, error) {
	return d.UnmarshalFromMessage(data, 0)
}

// UnmarshalFromMessage decodes a domain name located at offset inside the full
// message buffer and sets the receiver. Compression pointers (0xC0) are resolved
// relative to the start of message, as mandated by RFC 1035 §4.1.4.
//
// It returns the number of bytes consumed in the original stream at offset. A
// compression pointer consumes exactly two bytes at the point it appears,
// regardless of the length of the name it references.
func (d *DomainName) UnmarshalFromMessage(message []byte, offset int) (int, error) {
	name, newOffset, err := DecodeDomainName(message, offset)
	if err != nil {
		return 0, err
	}
	*d = DomainName(name)
	return newOffset - offset, nil
}

// ValidateDomainName validates a domain name according to LLMNR/DNS label rules.
func ValidateDomainName(name string) error {
	d := DomainName(name)
	return d.Validate()
}

// EncodeDomainName serializes a domain name into LLMNR/DNS wire format.
// Labels are length-prefixed and the sequence is terminated by a zero-length label.
func EncodeDomainName(name string) ([]byte, error) {
	// Root or empty name encodes to 0x00
	if name == "" || name == "." {
		return []byte{0}, nil
	}

	trimmed := strings.TrimSuffix(name, ".")
	if len(trimmed) > constants.MaxDomainLength {
		return nil, errors.ErrNameTooLong
	}

	var buf []byte
	for _, label := range strings.Split(trimmed, ".") {
		if len(label) > constants.MaxLabelLength {
			return nil, errors.ErrLabelTooLong
		}
		buf = append(buf, byte(len(label)))
		buf = append(buf, []byte(label)...)
	}
	buf = append(buf, 0) // terminating root label
	return buf, nil
}

// DecodeDomainName parses a domain name from LLMNR/DNS wire format located at
// offset inside the full message buffer data. Because DNS/LLMNR compression
// pointers are offsets from the start of the message (RFC 1035 §4.1.4), data
// MUST be the whole message (starting at the header ID field) rather than a
// sub-slice, otherwise 0xC0 pointers resolve against the wrong origin.
//
// It returns the decoded name, the new offset (past the encoded name in the
// original stream — a compression pointer counts as exactly two bytes at the
// point it appears), and an error if decoding fails.
func DecodeDomainName(data []byte, offset int) (string, int, error) {
	if offset < 0 || offset >= len(data) {
		return "", offset, fmt.Errorf("offset out of bounds: %w", errors.ErrInvalidDomainName)
	}

	originalOffset := offset
	consumed := 0
	jumped := false
	labels := []string{}

	// visited records every offset at which we begin reading a label or a
	// pointer. Revisiting an offset means the message contains a compression
	// pointer cycle (or an overlapping name), which we reject instead of
	// looping forever. Combined with the strictly-backward pointer check
	// below, this guarantees termination on malformed data.
	visited := make(map[int]struct{})

	for {
		if offset >= len(data) {
			return "", originalOffset, fmt.Errorf("truncated name: %w", errors.ErrInvalidDomainName)
		}

		length := int(data[offset])

		// Pointer (11xxxxxx): a two-octet compression pointer whose low 14
		// bits are an offset from the start of the message (RFC 1035 §4.1.4).
		if length&constants.LabelPointer == constants.LabelPointer {
			// Need one more byte for the pointer
			if offset+1 >= len(data) {
				return "", originalOffset, fmt.Errorf("truncated pointer: %w", errors.ErrInvalidDomainName)
			}
			ptr := int(binary.BigEndian.Uint16(data[offset:]) & 0x3FFF)
			if !jumped {
				// A pointer consumes exactly two bytes from the original
				// stream, regardless of the length of the name it references.
				consumed += 2
			}
			// A pointer may only reference a prior occurrence of a name, i.e.
			// it must point strictly backwards. Forward and self references are
			// rejected per RFC guidance and to guarantee decoding terminates.
			if ptr >= offset {
				return "", originalOffset, fmt.Errorf("invalid pointer: not a backward reference: %w", errors.ErrInvalidDomainName)
			}
			if _, seen := visited[ptr]; seen {
				return "", originalOffset, fmt.Errorf("invalid pointer: cycle detected: %w", errors.ErrInvalidDomainName)
			}
			visited[offset] = struct{}{}
			// Follow the pointer
			offset = ptr
			jumped = true
			continue
		}

		if _, seen := visited[offset]; seen {
			return "", originalOffset, fmt.Errorf("invalid name: cycle detected: %w", errors.ErrInvalidDomainName)
		}
		visited[offset] = struct{}{}

		// End of name
		if length == 0 {
			if !jumped {
				consumed++ // account for the zero byte only when not following a pointer
			}
			break
		}

		// Regular label
		offset++
		if offset+length > len(data) {
			return "", originalOffset, fmt.Errorf("truncated label: %w", errors.ErrInvalidDomainName)
		}
		label := string(data[offset : offset+length])
		labels = append(labels, label)
		offset += length
		if !jumped {
			consumed += 1 + length
		}
	}

	if len(labels) == 0 {
		return ".", originalOffset + consumed, nil
	}
	return strings.Join(labels, "."), originalOffset + consumed, nil
}
