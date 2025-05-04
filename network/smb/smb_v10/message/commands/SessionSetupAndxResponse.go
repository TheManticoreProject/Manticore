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

// SessionSetupAndxResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/e7514918-a0f6-4932-9f00-ced094445537
type SessionSetupAndxResponse struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR
	AndXCommand types.UCHAR
	AndXReserved types.UCHAR
	AndXOffset types.USHORT
	Action types.USHORT

	// Data
	Pad []types.UCHAR
	NativeOS types.SMB_STRING
	NativeLanMan types.SMB_STRING
	PrimaryDomain types.SMB_STRING

}

// NewSessionSetupAndxResponse creates a new SessionSetupAndxResponse structure
//
// Returns:
// - A pointer to the new SessionSetupAndxResponse structure
func NewSessionSetupAndxResponse() *SessionSetupAndxResponse {
	c := &SessionSetupAndxResponse{
		// Parameters
		WordCount: types.UCHAR(0),
		AndXCommand: types.UCHAR(0),
		AndXReserved: types.UCHAR(0),
		AndXOffset: types.USHORT(0),
		Action: types.USHORT(0),

		// Data
		Pad: []types.UCHAR{},
		NativeOS: types.SMB_STRING{},
		NativeLanMan: types.SMB_STRING{},
		PrimaryDomain: types.SMB_STRING{},

	}

	c.Command.SetCommandCode(codes.SMB_COM_SESSION_SETUP_ANDX)

	return c
}


// IsAndX returns true if the command is an AndX
func (c *SessionSetupAndxResponse) IsAndX() bool {
	return true
}



// Marshal marshals the SessionSetupAndxResponse structure into a byte array
//
// Returns:
// - A byte array representing the SessionSetupAndxResponse structure
// - An error if the marshaling fails
func (c *SessionSetupAndxResponse) Marshal() ([]byte, error) {
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
	
	// Marshalling data Pad
	rawDataContent = append(rawDataContent, types.UCHAR(c.Pad))
	
	// Marshalling data NativeOS
	bytesStream, err := c.NativeOS.Marshal()
	if err != nil {
			return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)
	
	// Marshalling data NativeLanMan
	bytesStream, err := c.NativeLanMan.Marshal()
	if err != nil {
			return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)
	
	// Marshalling data PrimaryDomain
	bytesStream, err := c.PrimaryDomain.Marshal()
	if err != nil {
			return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)
	
	// Then marshal the parameters
	rawParametersContent := []byte{}
	
	// Marshalling parameter WordCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.WordCount))
	
	// Marshalling parameter AndXCommand
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.AndXCommand))
	
	// Marshalling parameter AndXReserved
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.AndXReserved))
	
	// Marshalling parameter AndXOffset
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.AndXOffset))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter Action
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Action))
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
func (c *SessionSetupAndxResponse) Unmarshal(data []byte) (int, error) {
	offset := 0

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(data)
	if err != nil {
		return 0, err
	}
	rawParametersContent := c.GetParameters().GetBytes()
	bytesRead, err = c.GetData().Unmarshal(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	rawDataContent := c.GetData().GetBytes()

	// First unmarshal the parameters
	offset = 0
	
	// Unmarshalling parameter WordCount
	if len(rawParametersContent) < offset+1 {
	    return offset, fmt.Errorf("data too short for WordCount")
	}
	c.WordCount = types.UCHAR(rawParametersContent[offset])
	offset++
	
	// Unmarshalling parameter AndXCommand
	if len(rawParametersContent) < offset+1 {
	    return offset, fmt.Errorf("data too short for AndXCommand")
	}
	c.AndXCommand = types.UCHAR(rawParametersContent[offset])
	offset++
	
	// Unmarshalling parameter AndXReserved
	if len(rawParametersContent) < offset+1 {
	    return offset, fmt.Errorf("data too short for AndXReserved")
	}
	c.AndXReserved = types.UCHAR(rawParametersContent[offset])
	offset++
	
	// Unmarshalling parameter AndXOffset
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for AndXOffset")
	}
	c.AndXOffset = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter Action
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for Action")
	}
	c.Action = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Then unmarshal the data
	offset = 0
	
	// Unmarshalling data Pad
	if len(rawDataContent) < offset+1 {
	    return offset, fmt.Errorf("rawParametersContent too short for Pad")
	}
	c.Pad = types.UCHAR(rawDataContent[offset])
	offset++
	
	// Unmarshalling data NativeOS
	bytesRead, err := c.NativeOS.Unmarshal(rawDataContent[offset:])
	if err != nil {
	    return offset, err
	}
	offset += bytesRead
	
	// Unmarshalling data NativeLanMan
	bytesRead, err := c.NativeLanMan.Unmarshal(rawDataContent[offset:])
	if err != nil {
	    return offset, err
	}
	offset += bytesRead
	
	// Unmarshalling data PrimaryDomain
	bytesRead, err := c.PrimaryDomain.Unmarshal(rawDataContent[offset:])
	if err != nil {
	    return offset, err
	}
	offset += bytesRead

	return offset, nil
}
