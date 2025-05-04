package message

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
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
	marshalled_message := []byte{}

	// Marshal the header
	marshalled_header, err := m.Header.Marshal()
	if err != nil {
		return nil, err
	}
	marshalled_message = append(marshalled_message, marshalled_header...)

	// Check if there is a command to marshal
	if m.Command != nil {
		marshalled_command, err := m.Command.Marshal()
		if err != nil {
			return nil, err
		}
		marshalled_message = append(marshalled_message, marshalled_command...)
	} else {
		return nil, fmt.Errorf("no command added to message")
	}

	return marshalled_message, nil
}

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
	bytesRead, err = m.Header.UnmarshalWithSecurityFeaturesReserved(marshalledData[:header.SMB_HEADER_SIZE])
	if err != nil {
		return err
	}
	marshalledData = marshalledData[bytesRead:]

	var c command_interface.CommandInterface
	if m.Header.IsResponse() {
		c, err = commands.CreateResponseCommand(m.Header.Command)
		if err != nil {
			return err
		}
	} else {
		c, err = commands.CreateRequestCommand(m.Header.Command)
		if err != nil {
			return err
		}
	}

	// Unmarshal the command
	_, err = c.Unmarshal(marshalledData)
	if err != nil {
		return err
	}
	m.Command = c

	return nil
}

// UnmarshalRequest deserializes a byte slice into the Message structure.
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
func (m *Message) UnmarshalRequest(marshalledData []byte) error {
	var err error
	bytesRead := 0

	// Check if data is long enough for the header
	if len(marshalledData) < header.SMB_HEADER_SIZE {
		return fmt.Errorf("data too short to unmarshal SMB message")
	}

	// Unmarshal the header (first 32 bytes)
	bytesRead, err = m.Header.UnmarshalWithSecurityFeaturesReserved(marshalledData[:header.SMB_HEADER_SIZE])
	if err != nil {
		return err
	}
	marshalledData = marshalledData[bytesRead:]

	c, err := commands.CreateRequestCommand(m.Header.Command)
	if err != nil {
		return err
	}

	// Unmarshal the command
	_, err = c.Unmarshal(marshalledData)
	if err != nil {
		return err
	}
	m.Command = c

	return nil
}
