package llmnr

import (
	"encoding/binary"
	"fmt"
	"math/rand"
)

// Header represents the LLMNR message header.
//
// The header contains essential information about the LLMNR message, including the message ID, flags, and counts of
// various sections such as questions, answers, authority records, and additional records.
//
// Fields:
//   - ID: A 16-bit identifier assigned by the program that generates any kind of query. This identifier is copied
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
//	    ID:      12345,
//	    Flags:   FlagQR,
//	    QDCount: 1,
//	    ANCount: 0,
//	    NSCount: 0,
//	    ARCount: 0,
//	}
type Header struct {
	ID      uint16 `json:"id"`
	Flags   uint16 `json:"flags"`
	QDCount uint16 `json:"qd_count"` // Question count
	ANCount uint16 `json:"an_count"` // Answer count
	NSCount uint16 `json:"ns_count"` // Authority count
	ARCount uint16 `json:"ar_count"` // Additional count
}

// Message represents an LLMNR message.
//
// An LLMNR (Link-Local Multicast Name Resolution) message consists of a header and four sections:
// Questions, Answers, Authority, and Additional. The header contains metadata about the message,
// such as the transaction ID and various flags. The Questions section contains the questions being
// asked in the message, while the Answers, Authority, and Additional sections contain resource
// records that provide answers, authority information, and additional information, respectively.
//
// Fields:
// - Header: The header of the LLMNR message, containing metadata such as the transaction ID and flags.
// - Questions: A slice of Question structs representing the questions in the message.
// - Answers: A slice of ResourceRecord structs representing the answers in the message.
// - Authority: A slice of ResourceRecord structs representing the authority records in the message.
// - Additional: A slice of ResourceRecord structs representing the additional records in the message.
type Message struct {
	Header
	Questions  []Question       `json:"questions"`
	Answers    []ResourceRecord `json:"answers"`
	Authority  []ResourceRecord `json:"authority"`
	Additional []ResourceRecord `json:"additional"`
}

// NewMessage creates a new LLMNR message with a randomly generated transaction ID and initializes
// the Questions, Answers, Authority, and Additional sections as empty slices. The Header of the
// message is also initialized with the generated transaction ID.
//
// Returns:
// - A pointer to the newly created Message instance.
func NewMessage() *Message {
	return &Message{
		Header: Header{
			ID: uint16(rand.Uint32()), // Generate random transaction ID
		},
		Questions:  make([]Question, 0),
		Answers:    make([]ResourceRecord, 0),
		Authority:  make([]ResourceRecord, 0),
		Additional: make([]ResourceRecord, 0),
	}
}

// CreateResponseFromMessage creates a new LLMNR message that is a response to the given message.
//
// Parameters:
// - msg: The message to create a response for.
//
// Returns:
// - A pointer to the newly created Message instance.
func CreateResponseFromMessage(msg *Message) *Message {
	response := NewMessage()

	response.Header.ID = msg.Header.ID
	response.Header.Flags = msg.Header.Flags | FlagQR

	response.Questions = []Question{}

	response.Answers = []ResourceRecord{}

	return response
}

// AddQuestion adds a question to the Questions section of the LLMNR message and updates the
// question count in the header. It validates the domain name of the question before adding it.
//
// Parameters:
// - name: The domain name for the question.
// - qtype: The type of the question (e.g., TypeA, TypeAAAA).
// - qclass: The class of the question (e.g., ClassIN).
//
// Returns:
// - An error if the domain name is invalid.
// - nil if the question is successfully added.
func (m *Message) AddQuestion(name string, qtype, qclass uint16) error {
	if err := ValidateDomainName(name); err != nil {
		return err
	}

	m.Questions = append(m.Questions, Question{
		Name:  name,
		Type:  qtype,
		Class: qclass,
	})

	m.QDCount = uint16(len(m.Questions))

	return nil
}

// AddAnswer adds a resource record to the Answers section of the LLMNR message and updates the
// answer count in the header. It validates the domain name of the resource record before adding it.
//
// Parameters:
// - rr: The resource record to be added to the Answers section.
//
// Returns:
// - An error if the domain name of the resource record is invalid.
// - nil if the resource record is successfully added.
func (m *Message) AddAnswer(rr ResourceRecord) error {
	if err := ValidateDomainName(rr.Name); err != nil {
		return err
	}

	m.Answers = append(m.Answers, rr)

	m.ANCount = uint16(len(m.Answers))

	return nil
}

// AddAnswerClassINTypeA adds a resource record with Class IN and Type A to the Answers section
// of the LLMNR message and updates the answer count in the header. It validates the domain name
// of the resource record before adding it.
//
// Parameters:
// - name: The domain name for the resource record.
// - rdata: The resource data for the Type A record (e.g., an IPv4 address).
//
// Returns:
// - An error if the domain name of the resource record is invalid.
// - nil if the resource record is successfully added.
func (m *Message) AddAnswerClassINTypeA(name, ip string) error {
	rdata := IPv4ToRData(ip)

	if rdata == nil {
		return fmt.Errorf("invalid IPv4 address")
	}

	rr := ResourceRecord{
		Name:  name,
		Type:  TypeA,
		Class: ClassIN,
		TTL:   30,
		RData: rdata,
	}
	rr.RDLength = uint16(len(rr.RData))

	// Check if the name is already in Questions
	found := false
	for _, question := range m.Questions {
		if question.Name == name {
			found = true
			break
		}
	}

	// If the name is not found in Questions, add it
	if !found {
		m.Questions = append(m.Questions, Question{
			Name:  name,
			Type:  TypeA,
			Class: ClassIN,
		})
		m.QDCount = uint16(len(m.Questions))
	}

	return m.AddAnswer(rr)
}

// AddAnswerClassINTypeAAAA adds a resource record with Class IN and Type AAAA to the Answers section
// of the LLMNR message and updates the answer count in the header. It validates the domain name
// of the resource record before adding it.
//
// Parameters:
// - name: The domain name for the resource record.
// - rdata: The resource data for the Type AAAA record (e.g., an IPv6 address).
//
// Returns:
// - An error if the domain name of the resource record is invalid.
// - nil if the resource record is successfully added.
func (m *Message) AddAnswerClassINTypeAAAA(name, ip string) error {
	rdata := IPv6ToRData(ip)

	if rdata == nil {
		return fmt.Errorf("invalid IPv6 address")
	}

	rr := ResourceRecord{
		Name:  name,
		Type:  TypeAAAA,
		Class: ClassIN,
		TTL:   30,
		RData: rdata,
	}
	rr.RDLength = uint16(len(rr.RData))

	// Check if the name is already in Questions
	found := false
	for _, question := range m.Questions {
		if question.Name == name {
			found = true
			break
		}
	}

	// If the name is not found in Questions, add it
	if !found {
		m.Questions = append(m.Questions, Question{
			Name:  name,
			Type:  TypeAAAA,
			Class: ClassIN,
		})
		m.QDCount = uint16(len(m.Questions))
	}

	return m.AddAnswer(rr)
}

// Validate checks the integrity of the LLMNR message by ensuring that the counts in the header
// match the actual number of questions, answers, authority, and additional records. It also
// validates the domain names in the questions and answers sections.
//
// Returns:
// - An error if any of the counts do not match or if any domain name is invalid.
// - nil if the message is valid.
func (m *Message) Validate() error {
	// Check counts match actual data
	if len(m.Questions) != int(m.QDCount) {
		return ErrInvalidMessage
	}
	if len(m.Answers) != int(m.ANCount) {
		return ErrInvalidMessage
	}
	if len(m.Authority) != int(m.NSCount) {
		return ErrInvalidMessage
	}
	if len(m.Additional) != int(m.ARCount) {
		return ErrInvalidMessage
	}

	// Validate all names in the message
	for _, q := range m.Questions {
		if err := ValidateDomainName(q.Name); err != nil {
			return err
		}
	}

	for _, rr := range m.Answers {
		if err := ValidateDomainName(rr.Name); err != nil {
			return err
		}
	}

	return nil
}

// Encode serializes the Message struct into a byte slice according to the LLMNR wire format.
// It encodes the header, questions, and answers sections of the message.
//
// Returns:
// - A byte slice containing the encoded message.
// - An error if encoding fails at any point, such as if there is an error encoding the questions or answers.
func (m *Message) Encode() ([]byte, error) {
	packet := make([]byte, 0, MaxPacketSize)

	bufferUint16 := make([]byte, 2)

	// Encode header
	// ID - A 16-bit identifier assigned by the program that generates any kind of query. This identifier is copied
	// to the corresponding reply and can be used by the requester to match up replies to outstanding queries.
	binary.BigEndian.PutUint16(bufferUint16, m.ID)
	packet = append(packet, bufferUint16...)

	// Flags - A 16-bit field containing various flags that control the message flow and interpretation. These flags
	// include the Query/Response flag (QR), Operation code (OP), Conflict flag (C), Truncation flag (TC), and Tentative flag (T).
	binary.BigEndian.PutUint16(bufferUint16, m.Flags)
	packet = append(packet, bufferUint16...)

	// QDCOUNT - An unsigned 16-bit integer specifying the number of entries in the question section.
	m.QDCount = uint16(len(m.Questions))
	binary.BigEndian.PutUint16(bufferUint16, m.QDCount)
	packet = append(packet, bufferUint16...)

	// ANCOUNT - An unsigned 16-bit integer specifying the number of resource records in the answer section.
	m.ANCount = uint16(len(m.Answers))
	binary.BigEndian.PutUint16(bufferUint16, m.ANCount)
	packet = append(packet, bufferUint16...)

	// NSCOUNT - An unsigned 16-bit integer specifying the number of name server resource records in the authority records section.
	m.NSCount = uint16(len(m.Authority))
	binary.BigEndian.PutUint16(bufferUint16, m.NSCount)
	packet = append(packet, bufferUint16...)

	// ARCOUNT - An unsigned 16-bit integer specifying the number of resource records in the additional records section.
	m.ARCount = uint16(len(m.Additional))
	binary.BigEndian.PutUint16(bufferUint16, m.ARCount)
	packet = append(packet, bufferUint16...)

	// Encode questions
	for _, q := range m.Questions {
		questionBuf, err := EncodeQuestion(q)
		if err != nil {
			return nil, fmt.Errorf("encoding question: %w", err)
		}
		packet = append(packet, questionBuf...)
	}

	// Encode answers
	for _, a := range m.Answers {
		answerBuf, err := EncodeResourceRecord(a)
		if err != nil {
			return nil, fmt.Errorf("encoding answer: %w", err)
		}
		packet = append(packet, answerBuf...)
	}

	return packet, nil
}

// DecodeMessage decodes a byte slice into a Message struct. It expects the byte slice to be in the wire format
// as specified by the LLMNR protocol. The function first checks if the provided data is at least as long as the
// LLMNR header. It then proceeds to decode the header fields, followed by the question and answer sections.
//
// Parameters:
// - data: A byte slice containing the LLMNR message in wire format.
//
// Returns:
//   - A pointer to a Message struct containing the decoded data.
//   - An error if the decoding fails at any point, such as if the data is too short or if there is an error
//     decoding the question or answer sections.
func DecodeMessage(data []byte) (*Message, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("message too short")
	}

	msg := &Message{}

	// Decode header
	msg.ID = binary.BigEndian.Uint16(data[0:])

	msg.Flags = binary.BigEndian.Uint16(data[2:])

	msg.QDCount = binary.BigEndian.Uint16(data[4:])

	msg.ANCount = binary.BigEndian.Uint16(data[6:])

	msg.NSCount = binary.BigEndian.Uint16(data[8:])

	msg.ARCount = binary.BigEndian.Uint16(data[10:])

	offset := HeaderSize

	// Decode questions
	var err error
	for i := uint16(0); i < msg.QDCount; i++ {
		var q Question
		q, offset, err = DecodeQuestion(data, offset)
		if err != nil {
			return nil, fmt.Errorf("decoding question: %w", err)
		}
		msg.Questions = append(msg.Questions, q)
	}

	// Decode answers
	for i := uint16(0); i < msg.ANCount; i++ {
		var rr ResourceRecord
		rr, offset, err = DecodeResourceRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("decoding answer: %w", err)
		}
		msg.Answers = append(msg.Answers, rr)
	}

	return msg, nil
}

// IsQuery returns true if the message is a query.
//
// This function checks the QR (Query/Response) flag in the message's Flags field.
// If the QR flag is not set, the message is considered a query and the function returns true.
// If the QR flag is set, the message is considered a response and the function returns false.
//
// Returns:
//   - A boolean value indicating whether the message is a query (true) or not (false).
func (m *Message) IsQuery() bool {
	return (m.Flags & FlagQR) == 0
}

// IsResponse returns true if the message is a response.
//
// This function checks the QR (Query/Response) flag in the message's Flags field.
// If the QR flag is set, the message is considered a response and the function returns true.
// If the QR flag is not set, the message is considered a query and the function returns false.
//
// Returns:
//   - A boolean value indicating whether the message is a response (true) or not (false).
func (m *Message) IsResponse() bool {
	return (m.Flags & FlagQR) != 0
}

// SetQuery marks the message as a query.
//
// This function clears the QR (Query/Response) flag in the message's Flags field.
// By clearing the QR flag, the message is considered a query.
//
// Usage:
//
//	msg.SetQuery()
//
// After calling this function, the message will be marked as a query.
//
// Returns:
//   - Nothing. This function modifies the message in place.
func (m *Message) SetQuery() {
	m.Flags &^= FlagQR
}

// SetResponse marks the message as a response.
//
// This function sets the QR (Query/Response) flag in the message's Flags field.
// By setting the QR flag, the message is considered a response.
//
// Usage:
//
//	msg.SetResponse()
//
// After calling this function, the message will be marked as a response.
//
// Returns:
//   - Nothing. This function modifies the message in place.
func (m *Message) SetResponse() {
	m.Flags |= FlagQR
}
