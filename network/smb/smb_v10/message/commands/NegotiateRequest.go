package commands

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// NegotiateRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/25c8c3c9-58fc-4bb8-aa8f-0272dede84c5
type NegotiateRequest struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR

	// Data

	// Dialects (variable): This is a variable length list of dialect identifiers in order of preference from least to most
	// preferred. The client MUST list only dialects that it supports. The structure of the list entries is as follows:
	// BufferFormat (1 byte): This field MUST be 0x02. This is a buffer format indicator that identifies the next field as a null-terminated array of characters.
	// DialectString (variable): A null-terminated string identifying an SMB dialect. A list of common dialects is presented in section 1.7.
	Dialects dialects.Dialects
}

// NewNegotiateRequest creates a new NegotiateRequest structure
//
// Returns:
// - A pointer to the new NegotiateRequest structure
func NewNegotiateRequest() *NegotiateRequest {
	c := &NegotiateRequest{
		// Parameters
		WordCount: types.UCHAR(0),

		// Data
		Dialects: dialects.Dialects{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_NEGOTIATE)

	return c
}

// Marshal marshals the NegotiateRequest structure into a byte array
//
// Returns:
// - A byte array representing the NegotiateRequest structure
// - An error if the marshaling fails
func (c *NegotiateRequest) Marshal() ([]byte, error) {
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

	// Marshalling data Dialects
	marshalledDialects, err := c.Dialects.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, marshalledDialects...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

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
func (c *NegotiateRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter WordCount
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("data too short for WordCount")
	}
	c.WordCount = types.UCHAR(rawParametersContent[offset])
	offset++

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Dialects
	if len(rawDataContent) < offset+1 {
		return offset, fmt.Errorf("rawParametersContent too short for Dialects")
	}
	bytesRead, err = c.Dialects.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	return offset, nil
}
