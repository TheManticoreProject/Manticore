package question

import (
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/llmnr/class"
	"github.com/TheManticoreProject/Manticore/network/llmnr/domain_name"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
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
	Name  domain_name.DomainName `json:"name"`
	Type  llmnr_type.Type        `json:"type"`
	Class class.Class            `json:"class"`
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
func (q *Question) Marshal() ([]byte, error) {
	buf := []byte{}

	// Marshal domain name
	nameBuf, err := q.Name.Marshal()
	if err != nil {
		return nil, fmt.Errorf("error marshalling domain name: %w", err)
	}
	buf = append(buf, nameBuf...)

	// Marshal type
	typeBuf, err := q.Type.Marshal()
	if err != nil {
		return nil, fmt.Errorf("error marshalling type: %w", err)
	}
	buf = append(buf, typeBuf...)

	// Marshal class
	classBuf, err := q.Class.Marshal()
	if err != nil {
		return nil, fmt.Errorf("error marshalling class: %w", err)
	}
	buf = append(buf, classBuf...)

	return buf, nil
}

// Unmarshal decodes a byte slice into the receiver Question struct.
//
// This method takes a byte slice (starting at offset 0) and decodes its wire format
// representation into the Question struct fields. It decodes the domain name first,
// followed by the type and class fields. It returns the number of bytes read and an error
// if any part of the decoding process fails.
//
// Parameters:
// - data: A byte slice containing the encoded question in wire format.
//
// Returns:
//   - An integer representing the number of bytes read from data.
//   - An error if the decoding fails at any point, such as if the data is truncated or if there is an error
//     decoding the domain name.
//
// Usage:
//
//	var q Question
//	n, err := q.Unmarshal(data)
//	if err != nil {
//	    // handle error
//	}
func (q *Question) Unmarshal(data []byte) (int, error) {
	return q.UnmarshalFromMessage(data, 0)
}

// UnmarshalFromMessage decodes a Question located at offset inside the full
// message buffer. It is used instead of Unmarshal when the question's name may
// contain 0xC0 compression pointers, so those pointers resolve against the
// start of the message (RFC 1035 §4.1.4) rather than against a sub-slice.
//
// Parameters:
// - message: The full LLMNR message in wire format.
// - offset: The absolute offset of the question inside message.
//
// Returns:
//   - The number of bytes consumed from message at offset.
//   - An error if the decoding fails at any point.
func (q *Question) UnmarshalFromMessage(message []byte, offset int) (int, error) {
	start := offset

	// Unmarshal domain name
	bytesReadDomainName, err := q.Name.UnmarshalFromMessage(message, offset)
	if err != nil {
		return 0, fmt.Errorf("error unmarshalling domain name: %w", err)
	}
	offset += bytesReadDomainName

	// Unmarshal type
	if offset+2 > len(message) {
		return 0, fmt.Errorf("truncated question, missing type")
	}
	bytesReadType, err := q.Type.Unmarshal(message[offset : offset+2])
	if err != nil {
		return 0, fmt.Errorf("error unmarshalling type: %w", err)
	}
	offset += bytesReadType

	// Unmarshal class
	if offset+2 > len(message) {
		return 0, fmt.Errorf("truncated question, missing class")
	}
	bytesReadClass, err := q.Class.Unmarshal(message[offset : offset+2])
	if err != nil {
		return 0, fmt.Errorf("error unmarshalling class: %w", err)
	}
	offset += bytesReadClass

	return offset - start, nil
}

// Describe prints a detailed description of the Question.
//
// Parameters:
// - indent: An integer value specifying the indentation level for the output.
func (q *Question) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<Question>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mName\x1b[0m: %s\n", indentPrompt, q.Name)
	fmt.Printf("%s │ \x1b[93mType\x1b[0m: %s (0x%04x)\n", indentPrompt, q.Type.String(), q.Type)
	fmt.Printf("%s │ \x1b[93mClass\x1b[0m: %s (0x%04x)\n", indentPrompt, q.Class.String(), q.Class)
	fmt.Printf("%s └───\n", indentPrompt)
}
