package llmnr

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// Common errors
var (
	ErrNameTooLong  = errors.New("domain name too long")
	ErrLabelTooLong = errors.New("label too long")
)

// ValidateDomainName validates a domain name according to the rules specified by the LLMNR protocol.
// The function checks if the domain name exceeds the maximum allowed length and if any label within the domain name
// exceeds the maximum allowed label length. A domain name is composed of labels separated by dots.
//
// Parameters:
// - name: The domain name to be validated.
//
// Returns:
//   - An error if the domain name is invalid, such as if it is too long or if any label is too long.
//   - nil if the domain name is valid.
//
// Usage example:
//
//	err := ValidateDomainName("example.local")
//	if err != nil {
//	    log.Fatalf("Invalid domain name: %v", err)
//	}
//
// The function will return an error in the following cases:
// - If the domain name exceeds MaxDomainLength.
// - If any label within the domain name exceeds MaxLabelLength.
func ValidateDomainName(name string) error {
	if len(name) > MaxDomainLength {
		return ErrNameTooLong
	}

	labels := strings.Split(name, ".")
	for _, label := range labels {
		if len(label) > MaxLabelLength {
			return ErrLabelTooLong
		}
	}
	return nil
}

// EncodeDomainName encodes a domain name into a byte slice in the wire format as specified by the LLMNR protocol.
// The function converts the domain name into a sequence of labels, each prefixed by its length, and ends with a zero byte.
//
// Parameters:
// - name: The domain name to be encoded.
//
// Returns:
//   - A byte slice containing the encoded domain name.
//   - An error if the encoding fails, such as if a label is too long.
//
// Usage example:
//
//	encodedName, err := EncodeDomainName("example.local")
//	if err != nil {
//	    log.Fatalf("Failed to encode domain name: %v", err)
//	}
//
// The resulting byte slice for "example.local" would be:
//
//	[]byte{7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 5, 'l', 'o', 'c', 'a', 'l', 0}
func EncodeDomainName(name string) ([]byte, error) {
	if name == "" {
		return []byte{0}, nil
	}

	var buf []byte
	for _, label := range strings.Split(name, ".") {
		if len(label) > MaxLabelLength {
			return nil, ErrLabelTooLong
		}
		buf = append(buf, byte(len(label)))
		buf = append(buf, label...)
	}

	return append(buf, 0), nil
}

// DecodeDomainName decodes a byte slice into a domain name string in the wire format as specified by the LLMNR protocol.
// The function processes the byte slice to extract the sequence of labels, handling pointers if present, and reconstructs
// the original domain name.
//
// Parameters:
// - data: A byte slice containing the encoded domain name in wire format.
// - offset: The starting position in the byte slice from which to begin decoding.
//
// Returns:
//   - A string containing the decoded domain name.
//   - An integer representing the new offset position after decoding.
//   - An error if the decoding fails at any point, such as if the data is too short, if there is an invalid pointer, or if a label is too long.
//
// Usage example:
//
//	data := []byte{7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 5, 'l', 'o', 'c', 'a', 'l', 0}
//	offset := 0
//	domainName, newOffset, err := DecodeDomainName(data, offset)
//	if err != nil {
//	    log.Fatalf("Failed to decode domain name: %v", err)
//	}
//	fmt.Printf("Decoded domain name: %s, New offset: %d\n", domainName, newOffset)
//
// The resulting domain name for the given byte slice would be "example.local" and the new offset would be 14.
func DecodeDomainName(data []byte, offset int) (string, int, error) {
	if offset >= len(data) {
		return "", offset, fmt.Errorf("offset out of bounds")
	}

	var (
		labels []string
		start  = offset
		curr   = offset
	)

	for {
		if curr >= len(data) {
			return "", start, fmt.Errorf("truncated name")
		}

		length := int(data[curr])
		if length == 0 {
			curr++
			break
		}

		if length&labelPointer == labelPointer {
			if curr+1 >= len(data) {
				return "", start, fmt.Errorf("truncated pointer")
			}
			pointer := int(binary.BigEndian.Uint16(data[curr:]) & 0x3FFF)
			if pointer >= start {
				return "", start, fmt.Errorf("invalid pointer")
			}
			suffix, _, err := DecodeDomainName(data, pointer)
			if err != nil {
				return "", start, err
			}
			if len(labels) > 0 {
				return strings.Join(labels, ".") + "." + suffix, curr + 2, nil
			}
			return suffix, curr + 2, nil
		}

		curr++
		if curr+length > len(data) {
			return "", start, ErrLabelTooLong
		}

		labels = append(labels, string(data[curr:curr+length]))
		curr += length
	}

	if len(labels) == 0 {
		return ".", curr, nil
	}

	return strings.Join(labels, "."), curr, nil
}
