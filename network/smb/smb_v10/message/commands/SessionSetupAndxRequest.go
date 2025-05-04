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

// SessionSetupAndxRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/81e15dee-8fb6-4102-8644-7eaa7ded63f7
type SessionSetupAndxRequest struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR
	AndXCommand types.UCHAR
	AndXReserved types.UCHAR
	AndXOffset types.USHORT
	MaxBufferSize types.USHORT
	MaxMpxCount types.USHORT
	VcNumber types.USHORT
	SessionKey types.ULONG
	OEMPasswordLen types.USHORT
	UnicodePasswordLen types.USHORT
	Reserved types.ULONG
	Capabilities types.ULONG

	// Data
	OEMPassword []types.UCHAR
	UnicodePassword []types.UCHAR
	Pad []types.UCHAR
	AccountName types.SMB_STRING
	PrimaryDomain types.SMB_STRING
	NativeOS types.SMB_STRING
	NativeLanMan types.SMB_STRING

}

// NewSessionSetupAndxRequest creates a new SessionSetupAndxRequest structure
//
// Returns:
// - A pointer to the new SessionSetupAndxRequest structure
func NewSessionSetupAndxRequest() *SessionSetupAndxRequest {
	c := &SessionSetupAndxRequest{
		// Parameters
		WordCount: types.UCHAR(0),
		AndXCommand: types.UCHAR(0),
		AndXReserved: types.UCHAR(0),
		AndXOffset: types.USHORT(0),
		MaxBufferSize: types.USHORT(0),
		MaxMpxCount: types.USHORT(0),
		VcNumber: types.USHORT(0),
		SessionKey: types.ULONG(0),
		OEMPasswordLen: types.USHORT(0),
		UnicodePasswordLen: types.USHORT(0),
		Reserved: types.ULONG(0),
		Capabilities: types.ULONG(0),

		// Data
		OEMPassword: []types.UCHAR{},
		UnicodePassword: []types.UCHAR{},
		Pad: []types.UCHAR{},
		AccountName: types.SMB_STRING{},
		PrimaryDomain: types.SMB_STRING{},
		NativeOS: types.SMB_STRING{},
		NativeLanMan: types.SMB_STRING{},

	}

	c.Command.SetCommandCode(codes.SMB_COM_SESSION_SETUP_ANDX)

	return c
}


// IsAndX returns true if the command is an AndX
func (c *SessionSetupAndxRequest) IsAndX() bool {
	return true
}



// Marshal marshals the SessionSetupAndxRequest structure into a byte array
//
// Returns:
// - A byte array representing the SessionSetupAndxRequest structure
// - An error if the marshaling fails
func (c *SessionSetupAndxRequest) Marshal() ([]byte, error) {
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
	
	// Marshalling data OEMPassword
	rawDataContent = append(rawDataContent, types.UCHAR(c.OEMPassword))
	
	// Marshalling data UnicodePassword
	rawDataContent = append(rawDataContent, types.UCHAR(c.UnicodePassword))
	
	// Marshalling data Pad
	rawDataContent = append(rawDataContent, types.UCHAR(c.Pad))
	
	// Marshalling data AccountName
	bytesStream, err := c.AccountName.Marshal()
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
	
	// Marshalling parameter MaxBufferSize
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MaxBufferSize))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter MaxMpxCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MaxMpxCount))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter VcNumber
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.VcNumber))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter SessionKey
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.SessionKey))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter OEMPasswordLen
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.OEMPasswordLen))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter UnicodePasswordLen
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.UnicodePasswordLen))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter Reserved
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Reserved))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter Capabilities
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Capabilities))
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
func (c *SessionSetupAndxRequest) Unmarshal(data []byte) (int, error) {
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
	
	// Unmarshalling parameter MaxBufferSize
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for MaxBufferSize")
	}
	c.MaxBufferSize = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter MaxMpxCount
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for MaxMpxCount")
	}
	c.MaxMpxCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter VcNumber
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for VcNumber")
	}
	c.VcNumber = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter SessionKey
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for SessionKey")
	}
	c.SessionKey = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter OEMPasswordLen
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for OEMPasswordLen")
	}
	c.OEMPasswordLen = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter UnicodePasswordLen
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for UnicodePasswordLen")
	}
	c.UnicodePasswordLen = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for Reserved")
	}
	c.Reserved = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter Capabilities
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for Capabilities")
	}
	c.Capabilities = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Then unmarshal the data
	offset = 0
	
	// Unmarshalling data OEMPassword
	if len(rawDataContent) < offset+1 {
	    return offset, fmt.Errorf("rawParametersContent too short for OEMPassword")
	}
	c.OEMPassword = types.UCHAR(rawDataContent[offset])
	offset++
	
	// Unmarshalling data UnicodePassword
	if len(rawDataContent) < offset+1 {
	    return offset, fmt.Errorf("rawParametersContent too short for UnicodePassword")
	}
	c.UnicodePassword = types.UCHAR(rawDataContent[offset])
	offset++
	
	// Unmarshalling data Pad
	if len(rawDataContent) < offset+1 {
	    return offset, fmt.Errorf("rawParametersContent too short for Pad")
	}
	c.Pad = types.UCHAR(rawDataContent[offset])
	offset++
	
	// Unmarshalling data AccountName
	bytesRead, err := c.AccountName.Unmarshal(rawDataContent[offset:])
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

	return offset, nil
}
