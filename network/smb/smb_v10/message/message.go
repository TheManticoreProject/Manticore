package message

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
)

// SMBMessage represents an SMB message
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/4d330f4c-151c-4d79-b207-40bd4f754da9
type Message struct {
	Header  *header.Header
	Command command_interface.CommandInterface
}

// NewMessage creates a new SMB message
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/4d330f4c-151c-4d79-b207-40bd4f754da9
func NewMessage() *Message {
	return &Message{
		Header:  header.NewHeader(),
		Command: nil,
	}
}

// AddCommand adds a command to the message
// If the command is an AndX, it will add the command and the next command
// If the command is not an AndX, it will add the command
func (m *Message) AddCommand(command command_interface.CommandInterface) {
	if m.Command == nil {
		// First command added
		m.Header.Command = command.GetCommandCode()
		m.Command = command
	} else {
		// Add command to the end of the chain
		m.Command.AddCommandToChain(command)
	}
}

// Marshal serializes the Message into a byte slice.
//
// This method marshals the message's components in the following order:
//  1. Header - Contains protocol identifier, command code, status, flags, etc.
//  2. Parameters and Data blocks - For each ParameterDataBlock in the message:
//     a. Parameters - Contains command-specific parameters
//     b. Data - Contains command-specific data
//
// The marshalled data follows the SMB protocol format as specified in MS-CIFS.
// For AndX messages, multiple parameter-data blocks will be marshalled sequentially.
//
// Returns:
// - A byte slice containing the marshalled message
// - An error if marshalling any component fails
func (m *Message) Marshal() ([]byte, error) {
	// Marshal the header
	marshalled_header, err := m.Header.Marshal()
	if err != nil {
		return nil, err
	}

	// Check if there is a command to marshal
	if m.Command == nil {
		return nil, fmt.Errorf("no command added to message")
	}

	// Marshal every command in the chain, in order.
	//
	// Each one is marshalled exactly once: a command's Marshal appends to its own
	// parameter and data blocks, so marshalling it twice would emit those fields
	// twice. The AndX offsets are therefore patched into the marshalled bytes
	// below rather than filled in and re-marshalled.
	chain := []command_interface.CommandInterface{}
	blocks := [][]byte{}
	written := 0
	for command := m.Command; command != nil; command = command.GetNextCommand() {
		if command.GetNextCommand() != nil && !command.IsAndX() {
			return nil, fmt.Errorf(
				"command 0x%02X is followed in an AndX chain but is not an AndX command",
				uint8(command.GetCommandCode()))
		}

		// A command that describes its own fields by an offset from the SMB header
		// needs to know how far into the message it sits. Batched second or later,
		// a response that assumed the front of the message points its client at
		// the wrong bytes.
		if positioned, ok := command.(command_interface.ChainPositioned); ok {
			positioned.SetChainOffset(written)
		}

		block, err := command.Marshal()
		if err != nil {
			return nil, err
		}
		if command.IsAndX() && len(block) < andxBlockEnd {
			return nil, fmt.Errorf(
				"AndX command 0x%02X marshalled to %d bytes, too few to hold its AndX block",
				uint8(command.GetCommandCode()), len(block))
		}

		chain = append(chain, command)
		blocks = append(blocks, block)
		written += len(block)
	}

	// AndXOffset is measured from the start of the SMB header, so where a command
	// begins is known only once everything ahead of it has been marshalled. Walk
	// the blocks accumulating that position and patch each command's AndX block
	// with the command code and the offset of the one that follows it.
	position := len(marshalled_header)
	for index, block := range blocks {
		if index > 0 && chain[index-1].IsAndX() {
			previous := blocks[index-1]
			previous[andxCommandOffset] = uint8(chain[index].GetCommandCode())
			previous[andxReservedOffset] = 0x00
			binary.LittleEndian.PutUint16(previous[andxOffsetField:andxBlockEnd], uint16(position))
		}
		position += len(block)
	}

	// The chain is terminated explicitly. A command's own Marshal writes the
	// terminator when it was given no AndX block, but a command that was linked
	// and then turned out to be last would otherwise name a successor that is not
	// there, which a client would follow off the end of the message.
	if last := len(blocks) - 1; last >= 0 && chain[last].IsAndX() {
		blocks[last][andxCommandOffset] = uint8(codes.SMB_COM_NO_ANDX_COMMAND)
		blocks[last][andxReservedOffset] = 0x00
		binary.LittleEndian.PutUint16(blocks[last][andxOffsetField:andxBlockEnd], 0)
	}

	marshalled_message := append([]byte{}, marshalled_header...)
	for _, block := range blocks {
		marshalled_message = append(marshalled_message, block...)
	}

	return marshalled_message, nil
}

// The AndX block sits at the start of an AndX command's parameter words, which
// follow the one-byte WordCount: AndXCommand(1) AndXReserved(1) AndXOffset(2).
// Message.Unmarshal reads the same positions when it follows a chain.
const (
	andxCommandOffset  = 1
	andxReservedOffset = 2
	andxOffsetField    = 3
	andxBlockEnd       = 5
)

// Unmarshal deserializes a byte slice into the Message structure.
//
// This method reads the binary representation of the Message structure
// from the input byte slice according to the SMB protocol format. It processes
// the components in the following order:
//  1. Header - Reads protocol identifier, command code, status, flags, etc.
//  2. Parameters and Data blocks - For each ParameterDataBlock in the message:
//     a. Parameters - Reads command-specific parameters
//     b. Data - Reads command-specific data
//
// For AndX messages, multiple parameter-data blocks will be unmarshalled sequentially.
// The method will continue reading blocks until all data in the input slice is consumed.
//
// Parameters:
// - marshalledData: The byte slice containing the serialized Message structure
//
// Returns:
// - An error if unmarshalling any component fails, or nil if successful
func (m *Message) Unmarshal(marshalledData []byte) error {
	var err error
	bytesRead := 0

	// Check if data is long enough for the header
	if len(marshalledData) < header.SMB_HEADER_SIZE {
		return fmt.Errorf("data too short to unmarshal SMB message")
	}

	// Unmarshal the header (first 32 bytes)
	bytesRead, err = m.Header.Unmarshal(marshalledData[:header.SMB_HEADER_SIZE])
	if err != nil {
		return err
	}
	// Keep the full message buffer: AndX offsets are measured from the start of
	// the SMB header, so the chain can only be resolved against the whole message.
	fullData := marshalledData
	headerSize := bytesRead

	// newCommand builds the appropriate command type for a command code, based on
	// whether this message is a request or a response.
	newCommand := func(code codes.CommandCode) (command_interface.CommandInterface, error) {
		if m.Header.IsResponse() {
			return commands.CreateResponseCommand(code)
		}
		return commands.CreateRequestCommand(code)
	}

	// Decode the first command, located immediately after the header.
	c, err := newCommand(m.Header.Command)
	if err != nil {
		return err
	}
	// Init ensures the parameters and data structures are not nil.
	c.Init()
	// Propagate the message's Unicode setting (SMB_FLAGS2_UNICODE) so the command
	// decodes string fields with the correct character encoding.
	c.SetUnicode(m.Header.Flags2.IsUnicode())
	if _, err = c.Unmarshal(fullData[headerSize:]); err != nil {
		return err
	}
	m.Command = c

	// Follow the AndX chain. A batched ("AndX") command begins its parameter words
	// with AndXCommand(1) AndXReserved(1) AndXOffset(2, little-endian); AndXOffset
	// is the distance in bytes from the start of the SMB header to the next
	// command's WordCount field. Decode each chained command and link it in, until
	// the terminator (SMB_COM_NO_ANDX_COMMAND) or a null link is reached.
	prev := c
	commandStart := headerSize
	for prev.IsAndX() {
		// WordCount(1) precedes the AndX words, so the AndX block occupies
		// [commandStart+1 : commandStart+5].
		if commandStart+5 > len(fullData) {
			break
		}
		andxCommand := codes.CommandCode(fullData[commandStart+1])
		andxReserved := fullData[commandStart+2]
		andxOffset := int(binary.LittleEndian.Uint16(fullData[commandStart+3 : commandStart+5]))

		if andxCommand == codes.SMB_COM_NO_ANDX_COMMAND || andxOffset == 0 {
			break
		}
		if andxOffset < header.SMB_HEADER_SIZE || andxOffset >= len(fullData) {
			return fmt.Errorf("AndX offset %d out of bounds (message is %d bytes)", andxOffset, len(fullData))
		}
		// The chain is laid out forward in the buffer, so each command must
		// begin strictly after the current one. Requiring the offset to advance
		// rejects a cyclic chain (e.g. two commands whose AndXOffset fields point
		// at each other), which would otherwise loop forever and allocate a
		// command on every iteration.
		if andxOffset <= commandStart {
			return fmt.Errorf("AndX offset %d does not advance past current command at %d (cyclic chain)", andxOffset, commandStart)
		}

		next, err := newCommand(andxCommand)
		if err != nil {
			return err
		}
		next.Init()
		next.SetUnicode(m.Header.Flags2.IsUnicode())
		if _, err = next.Unmarshal(fullData[andxOffset:]); err != nil {
			return err
		}

		// GetNextCommand only traverses the chain when the command carries an AndX
		// struct, so record the parsed linkage on the previous command before linking.
		linkAndX := andx.NewAndX()
		linkAndX.AndXCommand = andxCommand
		linkAndX.AndXReserved = andxReserved
		linkAndX.AndXOffset = uint16(andxOffset)
		prev.SetAndX(linkAndX)
		prev.AddCommandToChain(next)

		prev = next
		commandStart = andxOffset
	}

	return nil
}
