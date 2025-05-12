package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// LockAndReadResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/ac41df92-de51-4d2d-935a-7eae93ae9bcc
type LockAndReadResponse struct {
	command_interface.Command

	// Parameters

	// CountOfBytesReturned (2 bytes): The actual number of bytes returned to the client. This MUST be equal to CountOfBytesToRead
	// unless the end of file was reached before reading CountOfBytesToRead bytes or the ReadOffsetInBytes pointed at or beyond the end of file.
	CountOfBytesReturned types.USHORT

	// Reserved (8 bytes): Reserved. All bytes MUST be 0x00.
	Reserved [4]types.USHORT

	// Data

	// BufferType (1 byte): This field MUST be 0x01.
	// CountOfBytesRead (2 bytes): The number of bytes read that are contained in the following array of bytes.
	// Bytes (variable): The array of bytes read from the file. The array is not null-terminated.
	BytesRead types.SMB_STRING
}

// NewLockAndReadResponse creates a new LockAndReadResponse structure
//
// Returns:
// - A pointer to the new LockAndReadResponse structure
func NewLockAndReadResponse() *LockAndReadResponse {
	c := &LockAndReadResponse{
		// Parameters
		CountOfBytesReturned: types.USHORT(0),
		Reserved:             [4]types.USHORT{},

		// Data
		BytesRead: types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_LOCK_AND_READ)

	return c
}

// Marshal marshals the LockAndReadResponse structure into a byte array
//
// Returns:
// - A byte array representing the LockAndReadResponse structure
// - An error if the marshaling fails
func (c *LockAndReadResponse) Marshal() ([]byte, error) {
	marshalledCommand := []byte{}

	// Create the Parameters structure if it is nil
	if c.GetParameters() == nil {
		c.SetParameters(parameters.NewParameters())
	}
	// Create the Data structure if it is nil
	if c.GetData() == nil {
		c.SetData(data.NewData())
	}

	// In case of AndX, we need to add the parameters to the Parameters structure first
	if c.IsAndX() {
		if c.GetAndX() == nil {
			c.SetAndX(andx.NewAndX())
			c.GetAndX().AndXCommand = codes.SMB_COM_NO_ANDX_COMMAND
		}

		for _, parameter := range c.GetAndX().GetParameters() {
			c.GetParameters().AddWord(parameter)
		}
	}

	// First marshal the data and then the parameters
	// This is because some parameters are dependent on the data, for example the size of some fields within
	// the data will be stored in the parameters
	rawDataContent := []byte{}

	// Marshalling data BytesRead
	marshalledBytesRead, err := c.BytesRead.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, marshalledBytesRead...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter CountOfBytesReturned
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.CountOfBytesReturned))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameters
	c.GetParameters().AddWordsFromBytesStream(rawParametersContent)
	marshalledParameters, err := c.GetParameters().Marshal()
	if err != nil {
		return nil, err
	}
	marshalledCommand = append(marshalledCommand, marshalledParameters...)

	// Marshalling data
	c.GetData().Add(rawDataContent)
	marshalledData, err := c.GetData().Marshal()
	if err != nil {
		return nil, err
	}
	marshalledCommand = append(marshalledCommand, marshalledData...)

	return marshalledCommand, nil
}

// Unmarshal unmarshals a byte array into the command structure
//
// Parameters:
// - data: The byte array to unmarshal
//
// Returns:
// - The number of bytes unmarshalled
func (c *LockAndReadResponse) Unmarshal(data []byte) (int, error) {
	offset := 0

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(data)
	if err != nil {
		return 0, err
	}
	rawParametersContent := c.GetParameters().GetBytes()
	_, err = c.GetData().Unmarshal(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	rawDataContent := c.GetData().GetBytes()

	// If the parameters and data are empty, this is a response containing an error code in
	// the SMB Header Status field
	if len(rawParametersContent) == 0 && len(rawDataContent) == 0 {
		return 0, nil
	}

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter CountOfBytesReturned
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for CountOfBytesReturned")
	}
	c.CountOfBytesReturned = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data BytesRead
	bytesRead, err = c.BytesRead.Unmarshal(rawDataContent)
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	return offset, nil
}
