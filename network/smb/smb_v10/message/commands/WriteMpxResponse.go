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

// WriteMpxResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/25efbb00-5ad0-42a2-8861-99a79c4d63fe
type WriteMpxResponse struct {
	command_interface.Command

	// Parameters

	// ResponseMask (4 bytes): This field is the logical OR-ing of the RequestMask
	// value contained in each SMB_COM_WRITE_MPX (section 2.2.4.26) received since the
	// last sequenced SMB_COM_WRITE_MPX. The server responds only to the final
	// (sequenced) command. This response contains the accumulated ResponseMask from
	// all successfully received requests. The client uses the ResponseMask received to
	// determine which packets, if any, MUST be retransmitted.
	ResponseMask types.ULONG
}

// NewWriteMpxResponse creates a new WriteMpxResponse structure
//
// Returns:
// - A pointer to the new WriteMpxResponse structure
func NewWriteMpxResponse() *WriteMpxResponse {
	c := &WriteMpxResponse{
		// Parameters
		ResponseMask: types.ULONG(0),
	}

	c.Command.SetCommandCode(codes.SMB_COM_WRITE_MPX)

	return c
}

// Marshal marshals the WriteMpxResponse structure into a byte array
//
// Returns:
// - A byte array representing the WriteMpxResponse structure
// - An error if the marshaling fails
func (c *WriteMpxResponse) Marshal() ([]byte, error) {
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

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter ResponseMask
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ResponseMask))
	rawParametersContent = append(rawParametersContent, buf4...)

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
func (c *WriteMpxResponse) Unmarshal(data []byte) (int, error) {
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
	_ = c.GetData().GetBytes()

	// If the parameters and data are empty, this is a response containing an error code in
	// the SMB Header Status field
	if len(rawParametersContent) == 0 {
		return 0, nil
	}

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter ResponseMask
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for ResponseMask")
	}
	c.ResponseMask = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
