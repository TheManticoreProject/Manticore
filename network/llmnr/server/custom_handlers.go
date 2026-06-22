package server

import (
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
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
// The function logs the details of the LLMNR message in JSON format.
func HandlerDescribePacketJson(server *Server, remoteAddr net.Addr, writer ResponseWriter, message *message.Message) bool {

	return false
}
