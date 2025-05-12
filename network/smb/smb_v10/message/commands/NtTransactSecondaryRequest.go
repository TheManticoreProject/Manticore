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

// NtTransactSecondaryRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/4173c449-a6e1-4fa9-b980-708a229fdb3a
type NtTransactSecondaryRequest struct {
	command_interface.Command

	// Parameters

	// TotalParameterCount (4 bytes): The total number of transaction parameter bytes
	// to be sent to the server over the course of this transaction. This value MAY be
	// less than or equal to the TotalParameterCount in preceding request messages that
	// are part of the same transaction. This value represents transaction parameter
	// bytes, not SMB parameter words.
	TotalParameterCount types.ULONG

	// TotalDataCount (4 bytes): The total number of transaction data bytes to be sent
	// to the server over the course of this transaction. This value MAY be less than
	// or equal to the TotalDataCount in preceding request messages that are part of
	// the same transaction. This value represents transaction data bytes, not SMB data
	// bytes.
	TotalDataCount types.ULONG

	// ParameterCount (4 bytes): The number of transaction parameter bytes being sent
	// in the SMB message. This value MUST be less than TotalParameterCount. The sum of
	// the ParameterCount values across all of the request messages in a transaction
	// MUST be equal to the TotalParameterCount reported in the last request message of
	// the transaction.
	ParameterCount types.ULONG

	// ParameterOffset (4 bytes): The offset, in bytes, from the start of the
	// SMB_Header to the transaction parameter bytes contained in this SMB message.
	// This MUST be the number of bytes from the start of the SMB message to the start
	// of the SMB_Data.Bytes.Parameters field. Server implementations MUST use this
	// value to locate the transaction parameter block within the SMB message. If
	// ParameterCount is zero, the client/server MAY set this field to zero.<117>
	ParameterOffset types.ULONG

	ParameterDisplacement types.ULONG

	// DataCount (4 bytes): The number of transaction data bytes being sent in this SMB
	// message. This value MUST be less than the value of TotalDataCount. The sum of
	// the DataCount values across all of the request messages in a transaction MUST be
	// equal to the smallest TotalDataCount value reported to the server.
	DataCount types.ULONG

	// DataOffset (4 bytes): The offset, in bytes, from the start of the SMB_Header to
	// the transaction data bytes contained in this SMB message. This MUST be the
	// number of bytes from the start of the SMB message to the start of the
	// SMB_Data.Bytes.Data field. Server implementations MUST use this value to locate
	// the transaction data block within the SMB message. If DataCount is zero, the
	// client/server MAY set this field to zero.<118>
	DataOffset types.ULONG

	// DataDisplacement (4 bytes): The offset, relative to all of the transaction data
	// bytes sent to the server in this transaction, at which this block of parameter
	// bytes MUST be placed. This value can be used by the server to correctly
	// reassemble the transaction data block even if the SMB request messages are
	// received out of order.
	DataDisplacement types.ULONG

	// Reserved2 (1 byte): Reserved. MUST be 0x00. The server MUST ignore the contents
	// of this field.
	Reserved2 types.UCHAR

	// Data

	// Pad1 (variable): This field SHOULD be used as an array of padding bytes to align
	// the following field to a 4-byte boundary relative to the start of the SMB
	// Header. This constraint can cause this field to be a zero-length field. This
	// field SHOULD be set to zero by the client/server and MUST be ignored by the
	// server/client.
	Pad1 []types.UCHAR

	// NT_Trans_Parameters (variable): Transaction parameter bytes. See the individual
	// SMB_COM_NT_TRANSACT subcommand descriptions for information on parameters sent for
	// each subcommand.
	NT_Trans_Parameters []types.UCHAR

	// Pad2 (variable): This field SHOULD be used as an array of padding bytes to align
	// the following field to a 4-byte boundary relative to the start of the SMB
	// Header. This constraint can cause this field to be a zero-length field. This
	// field SHOULD be set to zero by the client/server, and MUST be ignored by the
	// server/client.
	Pad2 []types.UCHAR

	// NT_Trans_Data (variable): Transaction data bytes. See the individual
	// SMB_COM_NT_TRANSACT subcommand descriptions for information on data sent
	// for each subcommand.
	NT_Trans_Data []types.UCHAR
}

// NewNtTransactSecondaryRequest creates a new NtTransactSecondaryRequest structure
//
// Returns:
// - A pointer to the new NtTransactSecondaryRequest structure
func NewNtTransactSecondaryRequest() *NtTransactSecondaryRequest {
	c := &NtTransactSecondaryRequest{
		// Parameters
		TotalParameterCount:   types.ULONG(0),
		TotalDataCount:        types.ULONG(0),
		ParameterCount:        types.ULONG(0),
		ParameterOffset:       types.ULONG(0),
		ParameterDisplacement: types.ULONG(0),
		DataCount:             types.ULONG(0),
		DataOffset:            types.ULONG(0),
		DataDisplacement:      types.ULONG(0),
		Reserved2:             types.UCHAR(0),

		// Data
		Pad1:                []types.UCHAR{},
		NT_Trans_Parameters: []types.UCHAR{},
		Pad2:                []types.UCHAR{},
		NT_Trans_Data:       []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_NT_TRANSACT_SECONDARY)

	return c
}

// Marshal marshals the NtTransactSecondaryRequest structure into a byte array
//
// Returns:
// - A byte array representing the NtTransactSecondaryRequest structure
// - An error if the marshaling fails
func (c *NtTransactSecondaryRequest) Marshal() ([]byte, error) {
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

	// Marshalling data Pad1
	rawDataContent = append(rawDataContent, c.Pad1...)

	// Marshalling data NT_Trans_Parameters
	rawDataContent = append(rawDataContent, c.NT_Trans_Parameters...)

	// Marshalling data Pad2
	rawDataContent = append(rawDataContent, c.Pad2...)

	// Marshalling data NT_Trans_Data
	rawDataContent = append(rawDataContent, c.NT_Trans_Data...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter TotalParameterCount
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.TotalParameterCount))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter TotalDataCount
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.TotalDataCount))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter ParameterCount
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ParameterCount))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter ParameterOffset
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ParameterOffset))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter ParameterDisplacement
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ParameterDisplacement))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter DataCount
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.DataCount))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter DataOffset
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.DataOffset))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter DataDisplacement
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.DataDisplacement))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter Reserved2
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.Reserved2))

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
func (c *NtTransactSecondaryRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter TotalParameterCount
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for TotalParameterCount")
	}
	c.TotalParameterCount = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter TotalDataCount
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for TotalDataCount")
	}
	c.TotalDataCount = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter ParameterCount
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for ParameterCount")
	}
	c.ParameterCount = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter ParameterOffset
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for ParameterOffset")
	}
	c.ParameterOffset = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter ParameterDisplacement
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for ParameterDisplacement")
	}
	c.ParameterDisplacement = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter DataCount
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for DataCount")
	}
	c.DataCount = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter DataOffset
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for DataOffset")
	}
	c.DataOffset = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter DataDisplacement
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for DataDisplacement")
	}
	c.DataDisplacement = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter Reserved2
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("data too short for Reserved2")
	}
	c.Reserved2 = types.UCHAR(rawParametersContent[offset])
	offset++

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Pad1
	if len(rawDataContent) < offset+int(c.ParameterOffset) {
		return offset, fmt.Errorf("rawDataContent too short for Pad1")
	}
	c.Pad1 = rawDataContent[offset : offset+int(c.ParameterOffset)]
	offset += int(c.ParameterOffset)

	// Unmarshalling data NT_Trans_Parameters
	if len(rawDataContent) < offset+int(c.ParameterCount) {
		return offset, fmt.Errorf("rawDataContent too short for NT_Trans_Parameters")
	}
	c.NT_Trans_Parameters = rawDataContent[offset : offset+int(c.ParameterCount)]
	offset += int(c.ParameterCount)

	// Unmarshalling data Pad2
	if len(rawDataContent) < offset+int(c.DataOffset) {
		return offset, fmt.Errorf("rawDataContent too short for Pad2")
	}
	c.Pad2 = rawDataContent[offset : offset+int(c.DataOffset)]
	offset += int(c.DataOffset)

	// Unmarshalling data NT_Trans_Data
	if len(rawDataContent) < offset+int(c.DataCount) {
		return offset, fmt.Errorf("rawDataContent too short for NT_Trans_Data")
	}
	c.NT_Trans_Data = rawDataContent[offset : offset+int(c.DataCount)]
	offset += int(c.DataCount)

	return offset, nil
}
