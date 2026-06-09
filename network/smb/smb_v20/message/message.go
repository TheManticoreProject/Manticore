package message

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
)

// Message is a single SMB2 message: a 64-byte header followed by one command
// body. Compounded messages (several header+body segments chained via the
// header NextCommand field) are handled by MarshalCompound / UnmarshalCompound;
// each segment is itself a Message.
type Message struct {
	Header  *header.Header
	Command command_interface.CommandInterface
}

// NewMessage creates an empty SMB2 message with a default header and no command.
func NewMessage() *Message {
	return &Message{
		Header:  header.NewHeader(),
		Command: nil,
	}
}

// SetCommand attaches a command to the message and copies its command code into
// the header.
func (m *Message) SetCommand(command command_interface.CommandInterface) {
	m.Command = command
	m.Header.Command = command.GetCommandCode()
}

// newCommandForHeader builds the appropriate command body for the message's
// command code, choosing the request or response variant from the header's
// SMB2_FLAGS_SERVER_TO_REDIR flag.
func newCommandForHeader(h *header.Header) (command_interface.CommandInterface, error) {
	if h.IsResponse() {
		return commands.CreateResponseCommand(h.Command)
	}
	return commands.CreateRequestCommand(h.Command)
}

// Marshal serializes a single (non-compounded) SMB2 message: the header followed
// by the command body. NextCommand is forced to 0. Use MarshalCompound to chain
// several messages.
func (m *Message) Marshal() ([]byte, error) {
	if m.Command == nil {
		return nil, fmt.Errorf("no command set on message")
	}

	m.Header.Command = m.Command.GetCommandCode()
	m.Header.NextCommand = 0

	marshalledHeader, err := m.Header.Marshal()
	if err != nil {
		return nil, err
	}

	marshalledCommand, err := m.Command.Marshal()
	if err != nil {
		return nil, err
	}

	return append(marshalledHeader, marshalledCommand...), nil
}

// Unmarshal deserializes a single SMB2 message from the start of data. The
// command body runs from the end of the header to the header's NextCommand
// offset when set (a compounded segment), or to the end of data otherwise.
//
// Returns the number of bytes consumed: the NextCommand offset when set, or the
// full length of data otherwise.
func (m *Message) Unmarshal(data []byte) (int, error) {
	if len(data) < header.SMB2_HEADER_SIZE {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 message: have %d bytes, need at least %d", len(data), header.SMB2_HEADER_SIZE)
	}

	headerBytes, err := m.Header.Unmarshal(data)
	if err != nil {
		return 0, err
	}

	// Determine where this command's region ends.
	regionEnd := len(data)
	consumed := len(data)
	if next := int(m.Header.NextCommand); next != 0 {
		if next < header.SMB2_HEADER_SIZE || next > len(data) {
			return 0, fmt.Errorf("NextCommand offset %d out of bounds (message is %d bytes)", next, len(data))
		}
		regionEnd = next
		consumed = next
	}

	command, err := newCommandForHeader(m.Header)
	if err != nil {
		return 0, err
	}
	if _, err := command.Unmarshal(data[headerBytes:regionEnd]); err != nil {
		return 0, err
	}
	m.Command = command

	return consumed, nil
}
