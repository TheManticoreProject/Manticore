package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
	"github.com/TheManticoreProject/Manticore/network/llmnr/resourcerecord"
)

// HandlerDescribePacket logs detailed information about an LLMNR packet received by the
//
// Parameters:
// - server: A pointer to the Server that received the packet.
// - remoteAddr: The address of the remote client that sent the packet.
// - writer: A ResponseWriter to send responses back to the client.
// - message: The LLMNR message received from the client.
//
// The function logs the following details about the received LLMNR packet:
// - The remote address of the client that sent the packet.
// - The number of questions in the packet, and for each question:
//   - The class of the question.
//   - The type of the question.
//   - The name in the question.
//
// - The number of answers in the packet, and for each answer:
//   - The class of the answer.
//   - The type of the answer.
//   - The name in the answer.
//   - The TTL (Time to Live) of the answer.
//   - The RDLENGTH (length of the RDATA field) of the answer.
//   - The RDATA (resource data) of the answer.
//
// - The number of authority records in the packet, and for each authority record:
//   - The class of the authority record.
//   - The type of the authority record.
//   - The name in the authority record.
//   - The TTL (Time to Live) of the authority record.
//   - The RDLENGTH (length of the RDATA field) of the authority record.
//   - The RDATA (resource data) of the authority record.
//
// The function uses a logger to output the information in a structured format, with indentation to
// represent the hierarchy of the packet's contents. The logger is locked during the function execution
// to ensure thread-safe logging.
func HandlerDescribePacket(server *Server, remoteAddr net.Addr, writer ResponseWriter, message *message.Message) bool {

	logger.Infof("Received LLMNR packet from [%s]", remoteAddr.String())

	if len(message.Questions) > 0 {
		logger.Infof(" ├─ Questions: (%d)", len(message.Questions))
		stringLen := len(fmt.Sprintf("%d", len(message.Questions)))
		formatString := fmt.Sprintf(" │  ├─ Question [%%0%dd/%%0%dd]", stringLen, stringLen)
		for i, q := range message.Questions {
			logger.Infof(formatString, i+1, len(message.Questions))
			logger.Infof(" │  │  ├─ Class : 0x%04x (%s)", q.Class, q.Class.String())
			logger.Infof(" │  │  ├─ Type  : 0x%04x (%s)", q.Type, q.Type.String())
			logger.Infof(" │  │  └─ Name  : \"%s\"", q.Name)
		}
		logger.Info(" │  └─ ")
	}

	if len(message.Answers) > 0 {
		logger.Infof(" ├─ Answers: (%d)", len(message.Answers))
		stringLen := len(fmt.Sprintf("%d", len(message.Answers)))
		formatString := fmt.Sprintf(" │  ├─ Answer [%%0%dd/%%0%dd]", stringLen, stringLen)
		for i, r := range message.Answers {
			logger.Infof(formatString, i+1, len(message.Answers))
			logger.Infof(" │  │  ├─ Class    : 0x%04x (%s)", r.Class, r.Class.String())
			logger.Infof(" │  │  ├─ Type     : 0x%04x (%s)", r.Type, r.Type.String())
			logger.Infof(" │  │  ├─ Name     : \"%s\"", r.Name)
			logger.Infof(" │  │  ├─ TTL      : %d", r.TTL)
			logger.Infof(" │  │  ├─ RDLENGTH : %d", r.RDLength)
			logger.Infof(" │  │  └─ RDATA    : %s", r.RData)
		}
		logger.Info(" │  └─ ")
	}

	if len(message.Authority) > 0 {
		logger.Infof(" ├─ Authority: (%d)", len(message.Authority))
		stringLen := len(fmt.Sprintf("%d", len(message.Authority)))
		formatString := fmt.Sprintf(" │  ├─ Authority [%%0%dd/%%0%dd]", stringLen, stringLen)
		for i, r := range message.Authority {
			logger.Infof(formatString, i+1, len(message.Authority))
			logger.Infof(" │  │  ├─ Class    : 0x%04x (%s)", r.Class, r.Class.String())
			logger.Infof(" │  │  ├─ Type     : 0x%04x (%s)", r.Type, r.Type.String())
			logger.Infof(" │  │  ├─ Name     : \"%s\"", r.Name)
			logger.Infof(" │  │  ├─ TTL      : %d", r.TTL)
			logger.Infof(" │  │  ├─ RDLENGTH : %d", r.RDLength)
			logger.Infof(" │  │  └─ RDATA    : %s", r.RData)
		}
		logger.Info(" │  └─ ")
	}

	if len(message.Additional) > 0 {
		logger.Infof(" ├─ Additional: (%d)", len(message.Additional))
		stringLen := len(fmt.Sprintf("%d", len(message.Additional)))
		formatString := fmt.Sprintf(" │  ├─ Additional [%%0%dd/%%0%dd]", stringLen, stringLen)
		for i, r := range message.Additional {
			logger.Infof(formatString, i+1, len(message.Additional))
			logger.Infof(" │  │  ├─ Class    : 0x%04x (%s)", r.Class, r.Class.String())
			logger.Infof(" │  │  ├─ Type     : 0x%04x (%s)", r.Type, r.Type.String())
			logger.Infof(" │  │  ├─ Name     : \"%s\"", r.Name)
			logger.Infof(" │  │  ├─ TTL      : %d", r.TTL)
			logger.Infof(" │  │  ├─ RDLENGTH : %d", r.RDLength)
			logger.Infof(" │  │  └─ RDATA    : %s", r.RData)
		}
		logger.Info(" │  └─ ")
	}

	logger.Info(" └─ ")

	logger.Info("")

	return false
}

// HandlerDescribePacketJson logs the details of the LLMNR message in JSON format.
//
// Parameters:
// - server: A pointer to the Server that received the packet.
// - remoteAddr: The address of the remote client that sent the packet.
// - writer: A ResponseWriter to send responses back to the client.
// - message: The LLMNR message received from the client.
//
// The function marshals the decoded message into an indented JSON document (see
// MessageToJson) and logs it via the same logger used by HandlerDescribePacket.
// Like its sibling it always returns false so that packet processing continues
// with the next handler.
func HandlerDescribePacketJson(server *Server, remoteAddr net.Addr, writer ResponseWriter, message *message.Message) bool {

	data, err := MessageToJson(remoteAddr, message)
	if err != nil {
		logger.Infof("Failed to render LLMNR packet from [%s] as JSON: %s", remoteAddr.String(), err)
		return false
	}

	logger.Info(string(data))

	return false
}

// jsonPacket is the JSON-shaped view of an LLMNR message as received from a
// remote peer. It intentionally exposes only the decoded protocol fields (never
// the wire-internal bookkeeping carried by the underlying structs) so the
// rendered document stays stable and human-readable.
type jsonPacket struct {
	RemoteAddr string               `json:"remote_addr"`
	Header     jsonHeader           `json:"header"`
	Questions  []jsonQuestion       `json:"questions"`
	Answers    []jsonResourceRecord `json:"answers"`
	Authority  []jsonResourceRecord `json:"authority"`
	Additional []jsonResourceRecord `json:"additional"`
}

// jsonHeader is the JSON-shaped view of the LLMNR header, exposing both the raw
// section counts and the individually decoded flag bits (QR/Opcode/C/TC/T/Z/RCODE
// per RFC 4795 §2.1.1).
type jsonHeader struct {
	ID      uint16    `json:"id"`
	Flags   jsonFlags `json:"flags"`
	QDCount uint16    `json:"qd_count"`
	ANCount uint16    `json:"an_count"`
	NSCount uint16    `json:"ns_count"`
	ARCount uint16    `json:"ar_count"`
}

// jsonFlags is the JSON-shaped view of the 16-bit LLMNR flags word, exposing the
// raw value alongside each decoded field.
type jsonFlags struct {
	Raw    uint16 `json:"raw"`
	QR     bool   `json:"qr"`
	Opcode uint8  `json:"opcode"`
	C      bool   `json:"c"`
	TC     bool   `json:"tc"`
	T      bool   `json:"t"`
	Z      uint8  `json:"z"`
	RCODE  uint8  `json:"rcode"`
}

// jsonQuestion is the JSON-shaped view of a question section entry.
type jsonQuestion struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	TypeCode  uint16 `json:"type_code"`
	Class     string `json:"class"`
	ClassCode uint16 `json:"class_code"`
}

// jsonResourceRecord is the JSON-shaped view of a resource record. RData holds a
// typed representation of the RDATA when the record type is recognized (via the
// As* accessors), and RDataHex always carries the raw RDATA bytes as a hex
// string so that unmodeled types (or records whose typed decode fails) remain
// fully represented.
type jsonResourceRecord struct {
	Name      string      `json:"name"`
	Type      string      `json:"type"`
	TypeCode  uint16      `json:"type_code"`
	Class     string      `json:"class"`
	ClassCode uint16      `json:"class_code"`
	TTL       uint32      `json:"ttl"`
	RDLength  uint16      `json:"rdlength"`
	RData     interface{} `json:"rdata,omitempty"`
	RDataHex  string      `json:"rdata_hex"`
}

// jsonSRV is the JSON-shaped view of an SRV record's typed RDATA (RFC 2782).
type jsonSRV struct {
	Priority uint16 `json:"priority"`
	Weight   uint16 `json:"weight"`
	Port     uint16 `json:"port"`
	Target   string `json:"target"`
}

// MessageToJson builds an indented JSON representation of a decoded LLMNR
// message. It renders the header (including the individually decoded flag bits
// and the section counts), the questions (name/type/class), and the answer,
// authority and additional resource records (name/type/class/ttl plus typed
// RDATA where the record type is recognized). The raw RDATA bytes are always
// included as a hex string so that unmodeled record types stay fully
// represented. The build is robust: a record whose typed decode errors simply
// falls back to its raw bytes and never panics.
//
// Parameters:
// - remoteAddr: The address of the remote peer that sent the message.
// - msg: The decoded LLMNR message to render.
//
// Returns:
// - The indented JSON document.
// - An error if JSON marshalling fails.
func MessageToJson(remoteAddr net.Addr, msg *message.Message) ([]byte, error) {
	packet := jsonPacket{
		Header: jsonHeader{
			ID: msg.Header.Identifier,
			Flags: jsonFlags{
				Raw:    uint16(msg.Header.Flags),
				QR:     msg.Header.Flags.IsResponse(),
				Opcode: msg.Header.Flags.Opcode(),
				C:      msg.Header.Flags.IsConflict(),
				TC:     msg.Header.Flags.IsTruncation(),
				T:      msg.Header.Flags.IsTentative(),
				Z:      msg.Header.Flags.Z(),
				RCODE:  msg.Header.Flags.RCODE(),
			},
			QDCount: msg.Header.QDCount,
			ANCount: msg.Header.ANCount,
			NSCount: msg.Header.NSCount,
			ARCount: msg.Header.ARCount,
		},
		Questions:  make([]jsonQuestion, 0, len(msg.Questions)),
		Answers:    make([]jsonResourceRecord, 0, len(msg.Answers)),
		Authority:  make([]jsonResourceRecord, 0, len(msg.Authority)),
		Additional: make([]jsonResourceRecord, 0, len(msg.Additional)),
	}

	if remoteAddr != nil {
		packet.RemoteAddr = remoteAddr.String()
	}

	for _, q := range msg.Questions {
		packet.Questions = append(packet.Questions, jsonQuestion{
			Name:      string(q.Name),
			Type:      q.Type.String(),
			TypeCode:  uint16(q.Type),
			Class:     q.Class.String(),
			ClassCode: uint16(q.Class),
		})
	}

	for i := range msg.Answers {
		packet.Answers = append(packet.Answers, resourceRecordToJson(&msg.Answers[i]))
	}
	for i := range msg.Authority {
		packet.Authority = append(packet.Authority, resourceRecordToJson(&msg.Authority[i]))
	}
	for i := range msg.Additional {
		packet.Additional = append(packet.Additional, resourceRecordToJson(&msg.Additional[i]))
	}

	return json.MarshalIndent(packet, "", "  ")
}

// resourceRecordToJson builds the JSON-shaped view of a single resource record,
// decoding a typed RDATA value for the record types recognized by the As*
// accessors and always retaining the raw RDATA bytes as a hex string. A typed
// decode that errors is not fatal: the typed value is simply left out and only
// the raw bytes are reported.
func resourceRecordToJson(rr *resourcerecord.ResourceRecord) jsonResourceRecord {
	record := jsonResourceRecord{
		Name:      string(rr.Name),
		Type:      rr.Type.String(),
		TypeCode:  uint16(rr.Type),
		Class:     rr.Class.String(),
		ClassCode: uint16(rr.Class),
		TTL:       rr.TTL,
		RDLength:  rr.RDLength,
		RDataHex:  hex.EncodeToString(rr.RData),
	}

	switch rr.Type {
	case llmnr_type.TypeA:
		if ip, err := rr.AsA(); err == nil {
			record.RData = ip.String()
		}
	case llmnr_type.TypeAAAA:
		if ip, err := rr.AsAAAA(); err == nil {
			record.RData = ip.String()
		}
	case llmnr_type.TypePTR:
		if name, err := rr.AsPTR(); err == nil {
			record.RData = name
		}
	case llmnr_type.TypeCNAME:
		if name, err := rr.AsCNAME(); err == nil {
			record.RData = name
		}
	case llmnr_type.TypeNS:
		if name, err := rr.AsNS(); err == nil {
			record.RData = name
		}
	case llmnr_type.TypeTXT:
		if strs, err := rr.AsTXT(); err == nil {
			record.RData = strs
		}
	case llmnr_type.TypeSRV:
		if priority, weight, port, target, err := rr.AsSRV(); err == nil {
			record.RData = jsonSRV{
				Priority: priority,
				Weight:   weight,
				Port:     port,
				Target:   target,
			}
		}
	}

	return record
}
