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

// SeekResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/089b214b-8501-4bcb-a9bd-9b2aa80b0042
type SeekResponse struct {
	command_interface.Command

	// Parameters

	// The response returns the new file pointer in Offset, expressed as the number of
	// bytes from the start of the file. The Offset MAY be beyond the current end of
	// file. An attempt to seek to before the start of file sets the current file
	// pointer to the start of the file (0x00000000).
	Offset types.ULONG
}

// NewSeekResponse creates a new SeekResponse structure
//
// Returns:
// - A pointer to the new SeekResponse structure
func NewSeekResponse() *SeekResponse {
	c := &SeekResponse{
		// Parameters
		Offset: types.ULONG(0),
	}

	c.Command.SetCommandCode(codes.SMB_COM_SEEK)

	return c
}

// Marshal marshals the SeekResponse structure into a byte array
//
// Returns:
// - A byte array representing the SeekResponse structure
// - An error if the marshaling fails
func (c *SeekResponse) Marshal() ([]byte, error) {
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
			c.GetAndX().AndXCommand = c.GetCommandCode()
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

	// Marshalling parameter Offset
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Offset))
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
func (c *SeekResponse) Unmarshal(data []byte) (int, error) {
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

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter Offset
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Offset")
	}
	c.Offset = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
