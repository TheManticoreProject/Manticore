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

// QueryInformationDiskResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/3d5291fd-33f5-4899-b3ad-949b0d4d7f93
type QueryInformationDiskResponse struct {
	command_interface.Command

	// Parameters

	// TotalUnits (2 bytes): This field is a 16-bit unsigned value that represents the
	// total count of logical allocation units available on the volume.
	TotalUnits types.USHORT

	// BlocksPerUnit (2 bytes): This field is a 16-bit unsigned value that represents
	// the number of blocks per allocation unit for the volume.
	BlocksPerUnit types.USHORT

	// BlockSize (2 bytes): This field is a 16-bit unsigned value that represents the
	// size in bytes of each allocation unit for the volume.
	BlockSize types.USHORT

	// FreeUnits (2 bytes): This field is a 16-bit unsigned value that represents the
	// total number of free allocation units available on the volume.
	FreeUnits types.USHORT

	// Reserved (2 bytes): This field is a 16-bit unsigned field and is reserved. The
	// client SHOULD ignore this field.
	Reserved types.USHORT
}

// NewQueryInformationDiskResponse creates a new QueryInformationDiskResponse structure
//
// Returns:
// - A pointer to the new QueryInformationDiskResponse structure
func NewQueryInformationDiskResponse() *QueryInformationDiskResponse {
	c := &QueryInformationDiskResponse{
		// Parameters

		TotalUnits:    types.USHORT(0),
		BlocksPerUnit: types.USHORT(0),
		BlockSize:     types.USHORT(0),
		FreeUnits:     types.USHORT(0),
		Reserved:      types.USHORT(0),
	}

	c.Command.SetCommandCode(codes.SMB_COM_QUERY_INFORMATION_DISK)

	return c
}

// Marshal marshals the QueryInformationDiskResponse structure into a byte array
//
// Returns:
// - A byte array representing the QueryInformationDiskResponse structure
// - An error if the marshaling fails
func (c *QueryInformationDiskResponse) Marshal() ([]byte, error) {
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

	// Marshalling parameter TotalUnits
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.TotalUnits))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter BlocksPerUnit
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.BlocksPerUnit))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter BlockSize
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.BlockSize))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter FreeUnits
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.FreeUnits))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Reserved
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Reserved))
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
func (c *QueryInformationDiskResponse) Unmarshal(data []byte) (int, error) {
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
	_ = c.GetData().GetBytes()

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter TotalUnits
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for TotalUnits")
	}
	c.TotalUnits = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter BlocksPerUnit
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for BlocksPerUnit")
	}
	c.BlocksPerUnit = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter BlockSize
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for BlockSize")
	}
	c.BlockSize = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter FreeUnits
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for FreeUnits")
	}
	c.FreeUnits = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved")
	}
	c.Reserved = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
