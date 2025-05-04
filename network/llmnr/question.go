package llmnr

import (
	"encoding/binary"
	"fmt"
)

// Question represents a question in an LLMNR message.
//
// An LLMNR (Link-Local Multicast Name Resolution) question consists of a domain name, a type, and a class.
// The domain name specifies the name being queried, while the type and class specify the type of the query
// (e.g., TypeA, TypeAAAA) and the class of the query (e.g., ClassIN), respectively.
//
// Fields:
// - Name: The domain name being queried.
// - Type: The type of the query (e.g., TypeA, TypeAAAA).
// - Class: The class of the query (e.g., ClassIN).
//
// The Question struct is used in the Questions section of an LLMNR message to represent individual questions
// being asked in the message. Each question is encoded and decoded using the EncodeQuestion and DecodeQuestion
// functions, respectively.
type Question struct {
	Name  string `json:"name"`
	Type  uint16 `json:"type"`
	Class uint16 `json:"class"`
}

// EncodeQuestion encodes a Question struct into a byte slice.
//
// This function takes a Question struct and encodes its fields into a byte slice in the wire format
// as specified by the LLMNR protocol. The domain name is encoded first, followed by the type and class
// fields. The encoded byte slice can then be included in an LLMNR message.
//
// Parameters:
// - buf: A byte slice to which the encoded question will be appended.
// - q: The Question struct to be encoded.
//
// Returns:
// - A byte slice containing the encoded question.
// - An error if the domain name encoding fails.

// Usage:
//
//	buf, err := EncodeQuestion(buf, question)
//	if err != nil {
//	    // handle error
//	}
//
// The function returns the updated byte slice with the encoded question appended to it, or an error
// if the domain name encoding fails.
func EncodeQuestion(q Question) ([]byte, error) {
	buf := []byte{}

	nameBuf, err := EncodeDomainName(q.Name)
	if err != nil {
		return nil, err
	}
	buf = append(buf, nameBuf...)

	bufferUint16 := make([]byte, 2)
	binary.BigEndian.PutUint16(bufferUint16, q.Type)
	buf = append(buf, bufferUint16...)

	binary.BigEndian.PutUint16(bufferUint16, q.Class)
	buf = append(buf, bufferUint16...)

	return buf, nil
}

// DecodeQuestion decodes a byte slice into a Question struct.
//
// This function takes a byte slice and an offset, and decodes the data starting from the offset
// into a Question struct. The domain name is decoded first, followed by the type and class fields.
// The function ensures that the data is not truncated and returns an error if any part of the
// decoding process fails.
//
// Parameters:
// - data: A byte slice containing the encoded question in wire format.
// - offset: The starting position in the byte slice from which to begin decoding.
//
// Returns:
//   - A Question struct containing the decoded data.
//   - An integer representing the new offset after decoding.
//   - An error if the decoding fails at any point, such as if the data is truncated or if there is an error
//     decoding the domain name.
//
// Usage:
//
//	question, newOffset, err := DecodeQuestion(data, offset)
//	if err != nil {
//	    // handle error
//	}
//
// The function returns the decoded Question struct, the new offset, and an error if any.
func DecodeQuestion(data []byte, offset int) (Question, int, error) {
	var q Question
	var err error

	q.Name, offset, err = DecodeDomainName(data, offset)
	if err != nil {
		return Question{}, offset, err
	}

	if offset+4 > len(data) {
		return Question{}, offset, fmt.Errorf("truncated question")
	}

	q.Type = binary.BigEndian.Uint16(data[offset:])
	offset += 2
	q.Class = binary.BigEndian.Uint16(data[offset:])
	offset += 2

	return q, offset, nil
}
