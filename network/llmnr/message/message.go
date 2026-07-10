package message

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/llmnr/class"
	"github.com/TheManticoreProject/Manticore/network/llmnr/constants"
	"github.com/TheManticoreProject/Manticore/network/llmnr/domain_name"
	"github.com/TheManticoreProject/Manticore/network/llmnr/errors"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message/header"
	"github.com/TheManticoreProject/Manticore/network/llmnr/question"
	"github.com/TheManticoreProject/Manticore/network/llmnr/resourcerecord"
)

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
	// The header of the LLMNR message, containing metadata such as the transaction ID and flags.
	Header header.Header
	// A slice of Question structs representing the questions in the message.
	Questions []question.Question `json:"questions"`

	// A slice of ResourceRecord structs representing the answers in the message.
	Answers []resourcerecord.ResourceRecord `json:"answers"`

	// A slice of ResourceRecord structs representing the authority records in the message.
	Authority []resourcerecord.ResourceRecord `json:"authority"`

	// A slice of ResourceRecord structs representing the additional records in the message.
	Additional []resourcerecord.ResourceRecord `json:"additional"`
}

// NewMessage creates a new LLMNR message with a randomly generated transaction ID and initializes
// the Questions, Answers, Authority, and Additional sections as empty slices. The Header of the
// message is also initialized with the generated transaction ID.
//
// Returns:
// - A pointer to the newly created Message instance.
func NewMessage() *Message {
	return &Message{
		Header: header.Header{
			Identifier: uint16(rand.Uint32()), // Generate random transaction ID
			Flags:      header.Flags(0),
			QDCount:    0,
			ANCount:    0,
			NSCount:    0,
			ARCount:    0,
		},
		Questions:  make([]question.Question, 0),
		Answers:    make([]resourcerecord.ResourceRecord, 0),
		Authority:  make([]resourcerecord.ResourceRecord, 0),
		Additional: make([]resourcerecord.ResourceRecord, 0),
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

	response.Header.Identifier = msg.Header.Identifier
	response.Header.Flags = msg.Header.Flags | header.FlagQR

	response.Questions = []question.Question{}

	response.Answers = []resourcerecord.ResourceRecord{}

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
func (m *Message) AddQuestion(name string, qtype llmnr_type.Type, qclass class.Class) error {
	if err := domain_name.ValidateDomainName(name); err != nil {
		return err
	}

	m.Questions = append(m.Questions, question.Question{
		Name:  domain_name.DomainName(name),
		Type:  qtype,
		Class: qclass,
	})

	m.Header.QDCount = uint16(len(m.Questions))

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
func (m *Message) AddAnswer(rr resourcerecord.ResourceRecord) error {
	if err := rr.Name.Validate(); err != nil {
		return err
	}

	m.Answers = append(m.Answers, rr)

	m.Header.ANCount = uint16(len(m.Answers))

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
	rdata := resourcerecord.IPv4ToRData(ip)

	if rdata == nil {
		return fmt.Errorf("invalid IPv4 address")
	}

	rr := resourcerecord.ResourceRecord{
		Name:  domain_name.DomainName(name),
		Type:  llmnr_type.TypeA,
		Class: class.ClassIN,
		TTL:   30,
		RData: rdata,
	}
	rr.RDLength = uint16(len(rr.RData))

	// Check if the name is already in Questions
	found := false
	for _, question := range m.Questions {
		if question.Name == domain_name.DomainName(name) {
			found = true
			break
		}
	}

	// If the name is not found in Questions, add it
	if !found {
		m.Questions = append(m.Questions, question.Question{
			Name:  domain_name.DomainName(name),
			Type:  llmnr_type.TypeA,
			Class: class.ClassIN,
		})
		m.Header.QDCount = uint16(len(m.Questions))
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
	rdata := resourcerecord.IPv6ToRData(ip)

	if rdata == nil {
		return fmt.Errorf("invalid IPv6 address")
	}

	rr := resourcerecord.ResourceRecord{
		Name:  domain_name.DomainName(name),
		Type:  llmnr_type.TypeAAAA,
		Class: class.ClassIN,
		TTL:   30,
		RData: rdata,
	}
	rr.RDLength = uint16(len(rr.RData))

	// Check if the name is already in Questions
	found := false
	for _, question := range m.Questions {
		if question.Name == domain_name.DomainName(name) {
			found = true
			break
		}
	}

	// If the name is not found in Questions, add it
	if !found {
		m.Questions = append(m.Questions, question.Question{
			Name:  domain_name.DomainName(name),
			Type:  llmnr_type.TypeAAAA,
			Class: class.ClassIN,
		})
		m.Header.QDCount = uint16(len(m.Questions))
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
	if len(m.Questions) != int(m.Header.QDCount) {
		return errors.ErrInvalidMessage
	}
	if len(m.Answers) != int(m.Header.ANCount) {
		return errors.ErrInvalidMessage
	}
	if len(m.Authority) != int(m.Header.NSCount) {
		return errors.ErrInvalidMessage
	}
	if len(m.Additional) != int(m.Header.ARCount) {
		return errors.ErrInvalidMessage
	}

	// Validate all names in the message
	for _, q := range m.Questions {
		if err := q.Name.Validate(); err != nil {
			return err
		}
	}

	for _, rr := range m.Answers {
		if err := rr.Name.Validate(); err != nil {
			return err
		}
	}

	for _, rr := range m.Authority {
		if err := rr.Name.Validate(); err != nil {
			return err
		}
	}

	for _, rr := range m.Additional {
		if err := rr.Name.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Marshal serializes the Message struct into a byte slice according to the LLMNR wire format.
// It encodes the header, questions, and answers sections of the message.
//
// Returns:
// - A byte slice containing the encoded message.
// - An error if encoding fails at any point, such as if there is an error encoding the questions or answers.
func (m *Message) Marshal() ([]byte, error) {
	marshalledData := make([]byte, 0, constants.MaxPacketSize)

	bufferUint16 := make([]byte, 2)

	// Encode header
	// ID - A 16-bit identifier assigned by the program that generates any kind of query. This identifier is copied
	// to the corresponding reply and can be used by the requester to match up replies to outstanding queries.
	binary.BigEndian.PutUint16(bufferUint16, m.Header.Identifier)
	marshalledData = append(marshalledData, bufferUint16...)

	// Flags - A 16-bit field containing various flags that control the message flow and interpretation. These flags
	// include the Query/Response flag (QR), Operation code (OP), Conflict flag (C), Truncation flag (TC), and Tentative flag (T).
	binary.BigEndian.PutUint16(bufferUint16, uint16(m.Header.Flags))
	marshalledData = append(marshalledData, bufferUint16...)

	// QDCOUNT - An unsigned 16-bit integer specifying the number of entries in the question section.
	m.Header.QDCount = uint16(len(m.Questions))
	binary.BigEndian.PutUint16(bufferUint16, m.Header.QDCount)
	marshalledData = append(marshalledData, bufferUint16...)

	// ANCOUNT - An unsigned 16-bit integer specifying the number of resource records in the answer section.
	m.Header.ANCount = uint16(len(m.Answers))
	binary.BigEndian.PutUint16(bufferUint16, m.Header.ANCount)
	marshalledData = append(marshalledData, bufferUint16...)

	// NSCOUNT - An unsigned 16-bit integer specifying the number of name server resource records in the authority records section.
	m.Header.NSCount = uint16(len(m.Authority))
	binary.BigEndian.PutUint16(bufferUint16, m.Header.NSCount)
	marshalledData = append(marshalledData, bufferUint16...)

	// ARCOUNT - An unsigned 16-bit integer specifying the number of resource records in the additional records section.
	m.Header.ARCount = uint16(len(m.Additional))
	binary.BigEndian.PutUint16(bufferUint16, m.Header.ARCount)
	marshalledData = append(marshalledData, bufferUint16...)

	// Encode questions
	for _, q := range m.Questions {
		questionBuf, err := q.Marshal()
		if err != nil {
			return nil, fmt.Errorf("encoding question: %w", err)
		}
		marshalledData = append(marshalledData, questionBuf...)
	}

	// Encode answers
	for _, a := range m.Answers {
		answerBuf, err := a.Marshal()
		if err != nil {
			return nil, fmt.Errorf("encoding answer: %w", err)
		}
		marshalledData = append(marshalledData, answerBuf...)
	}

	// Encode authority
	for _, a := range m.Authority {
		authorityBuf, err := a.Marshal()
		if err != nil {
			return nil, fmt.Errorf("encoding authority: %w", err)
		}
		marshalledData = append(marshalledData, authorityBuf...)
	}

	// Encode additional
	for _, a := range m.Additional {
		additionalBuf, err := a.Marshal()
		if err != nil {
			return nil, fmt.Errorf("encoding additional: %w", err)
		}
		marshalledData = append(marshalledData, additionalBuf...)
	}

	return marshalledData, nil
}

// Unmarshal decodes a byte slice into the Message receiver. It expects the byte slice to be in the wire format
// as specified by the LLMNR protocol. The function first checks if the provided data is at least as long as the
// LLMNR header. It then proceeds to decode the header fields, followed by the question and answer sections.
//
// Parameters:
// - data: A byte slice containing the LLMNR message in wire format.
//
// Returns:
//   - The number of bytes read from data.
//   - An error if the decoding fails at any point, such as if the data is too short or if there is an error
//     decoding the question or answer sections.
func (m *Message) Unmarshal(data []byte) (int, error) {
	if len(data) < header.HeaderSize {
		return 0, fmt.Errorf("message too short")
	}

	// Unmarshal header. From here on we track an absolute offset into data so
	// that 0xC0 compression pointers, which are relative to the start of the
	// message (RFC 1035 §4.1.4), resolve against the message origin rather than
	// against a sub-slice.
	offset := 0
	bytesReadHeader, err := m.Header.Unmarshal(data[offset : offset+header.HeaderSize])
	if err != nil {
		return 0, fmt.Errorf("error unmarshalling header: %w", err)
	}
	offset += bytesReadHeader

	// Decode questions
	for i := uint16(0); i < m.Header.QDCount; i++ {
		q := question.Question{}
		n, err := q.UnmarshalFromMessage(data, offset)
		if err != nil {
			return 0, fmt.Errorf("error unmarshalling question: %w", err)
		}
		offset += n
		m.Questions = append(m.Questions, q)
	}

	// Decode answers
	for i := uint16(0); i < m.Header.ANCount; i++ {
		rr := resourcerecord.ResourceRecord{}
		n, err := rr.UnmarshalFromMessage(data, offset)
		if err != nil {
			return 0, fmt.Errorf("error unmarshalling answer: %w", err)
		}
		offset += n
		m.Answers = append(m.Answers, rr)
	}

	// Decode authority
	for i := uint16(0); i < m.Header.NSCount; i++ {
		rr := resourcerecord.ResourceRecord{}
		n, err := rr.UnmarshalFromMessage(data, offset)
		if err != nil {
			return 0, fmt.Errorf("error unmarshalling authority: %w", err)
		}
		offset += n
		m.Authority = append(m.Authority, rr)
	}

	// Decode additional
	for i := uint16(0); i < m.Header.ARCount; i++ {
		rr := resourcerecord.ResourceRecord{}
		n, err := rr.UnmarshalFromMessage(data, offset)
		if err != nil {
			return 0, fmt.Errorf("error unmarshalling additional: %w", err)
		}
		offset += n
		m.Additional = append(m.Additional, rr)
	}

	return offset, nil
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
	return m.Header.Flags.IsQuery()
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
	return m.Header.Flags.IsResponse()
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
	m.Header.Flags &^= header.FlagQR
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
	m.Header.Flags |= header.FlagQR
}

// Describe prints a detailed description of the Message structure.
//
// Parameters:
// - indent: An integer value specifying the indentation level for the output.
func (m *Message) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)

	fmt.Printf("%s<Message>\n", indentPrompt)
	m.Header.Describe(indent + 1)

	fmt.Printf("%s │ \x1b[93mQuestions\x1b[0m (%d):\n", indentPrompt, len(m.Questions))
	for _, q := range m.Questions {
		q.Describe(indent + 1)
	}
	if len(m.Questions) > 0 {
		fmt.Printf("%s │  └───\n", indentPrompt)
	}

	fmt.Printf("%s │ \x1b[93mAnswers\x1b[0m (%d):\n", indentPrompt, len(m.Answers))
	for _, a := range m.Answers {
		a.Describe(indent + 1)
	}
	if len(m.Answers) > 0 {
		fmt.Printf("%s │  └───\n", indentPrompt)
	}

	fmt.Printf("%s │ \x1b[93mAuthority\x1b[0m (%d):\n", indentPrompt, len(m.Authority))
	for _, a := range m.Authority {
		a.Describe(indent + 1)
	}
	if len(m.Authority) > 0 {
		fmt.Printf("%s │  └───\n", indentPrompt)
	}

	fmt.Printf("%s │ \x1b[93mAdditional\x1b[0m (%d):\n", indentPrompt, len(m.Additional))
	for _, a := range m.Additional {
		a.Describe(indent + 1)
	}
	if len(m.Additional) > 0 {
		fmt.Printf("%s │  └───\n", indentPrompt)
	}

	fmt.Printf("%s └───\n", indentPrompt)
}
