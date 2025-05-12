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

// OpenPrintFileRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/a0199848-ec12-4408-9812-88a5f1c30ceb
type OpenPrintFileRequest struct {
	command_interface.Command

	// Parameters

	// SetupLength (2 bytes): Length, in bytes, of the printer-specific control data
	// that is to be included as the first part of the spool file. The server MUST pass
	// this initial portion of the spool file to the printer unmodified.
	SetupLength types.USHORT

	// Mode (2 bytes): A 16-bit field that contains a flag that specifies the print
	// file mode.
	Mode types.USHORT

	// Data

	// BufferFormat (1 byte): This field MUST be 0x04, representing an ASCII string.
	// Identifier (variable): A null-terminated string containing a suggested name for
	// the spool file. The server can ignore, modify, or use this information to
	// identify the print job.<126>
	Identifier types.SMB_STRING
}

// NewOpenPrintFileRequest creates a new OpenPrintFileRequest structure
//
// Returns:
// - A pointer to the new OpenPrintFileRequest structure
func NewOpenPrintFileRequest() *OpenPrintFileRequest {
	c := &OpenPrintFileRequest{
		// Parameters
		SetupLength: types.USHORT(0),
		Mode:        types.USHORT(0),

		// Data
		Identifier: types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_OPEN_PRINT_FILE)

	return c
}

// Marshal marshals the OpenPrintFileRequest structure into a byte array
//
// Returns:
// - A byte array representing the OpenPrintFileRequest structure
// - An error if the marshaling fails
func (c *OpenPrintFileRequest) Marshal() ([]byte, error) {
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

	// Marshalling data Identifier
	c.Identifier.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	bytesStream, err := c.Identifier.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter SetupLength
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.SetupLength))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Mode
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Mode))
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
func (c *OpenPrintFileRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter SetupLength
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for SetupLength")
	}
	c.SetupLength = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Mode
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Mode")
	}
	c.Mode = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Identifier
	bytesRead, err = c.Identifier.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	return offset, nil
}
