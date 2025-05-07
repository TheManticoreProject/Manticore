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

// TransactionSecondaryRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/79ece32a-139d-46b0-ba28-055f822a8c05
type TransactionSecondaryRequest struct {
	command_interface.Command

	// Parameters

	// TotalParameterCount (2 bytes): The total number of transaction parameter bytes
	// to be sent to the server over the course of this transaction. This value MAY be
	// less than or equal to the TotalParameterCount in preceding request messages that
	// are part of the same transaction. This value represents transaction parameter
	// bytes, not SMB parameter words.
	TotalParameterCount types.USHORT

	// TotalDataCount (2 bytes): The total number of transaction data bytes to be sent
	// to the server over the course of this transaction. This value MAY be less than
	// or equal to the TotalDataCount in preceding request messages that are part of
	// the same transaction. This value represents transaction data bytes, not SMB data
	// bytes.
	TotalDataCount types.USHORT

	// ParameterCount (2 bytes): The number of transaction parameter bytes being sent
	// in the SMB message. This value MUST be less than TotalParameterCount. The sum of
	// the ParameterCount values across all of the request messages in a transaction
	// MUST be equal to the TotalParameterCount reported in the last request message of
	// the transaction.
	ParameterCount types.USHORT

	// ParameterOffset (2 bytes): The offset, in bytes, from the start of the
	// SMB_Header to the transaction parameter bytes contained in this SMB message.
	// This MUST be the number of bytes from the start of the SMB message to the start
	// of the SMB_Data.Bytes.Trans_Parameters field. Server implementations MUST use
	// this value to locate the transaction parameter block within the SMB message. If
	// ParameterCount is zero, the client/server MAY set this field to zero.<47>
	ParameterOffset types.USHORT

	// ParameterDisplacement (2 bytes): The offset, relative to all of the transaction
	// parameter bytes sent to the server in this transaction, at which this block of
	// parameter bytes MUST be placed. This value can be used by the server to
	// correctly reassemble the transaction parameters even if the SMB request messages
	// are received out of order.
	ParameterDisplacement types.USHORT

	// DataCount (2 bytes): The number of transaction data bytes being sent in this SMB
	// message. This value MUST be less than the value of TotalDataCount. The sum of
	// the DataCount values across all of the request messages in a transaction MUST be
	// equal to the smallest TotalDataCount value reported to the server.
	DataCount types.USHORT

	// DataOffset (2 bytes): The offset, in bytes, from the start of the SMB_Header to
	// the transaction data bytes contained in this SMB message. This MUST be the
	// number of bytes from the start of the SMB message to the start of the
	// SMB_Data.Bytes.Trans_Data field. Server implementations MUST use this value to
	// locate the transaction data block within the SMB message. If DataCount is zero,
	// the client/server MAY set this field to zero.<48>
	DataOffset types.USHORT

	// DataDisplacement (2 bytes): The offset, relative to all of the transaction data
	// bytes sent to the server in this transaction, at which this block of parameter
	// bytes MUST be placed. This value can be used by the server to correctly
	// reassemble the transaction data block even if the SMB request messages are
	// received out of order.
	DataDisplacement types.USHORT

	// Data

	// Pad1 (variable): This field SHOULD be used as an array of padding bytes to align
	// the following field to a 4-byte boundary relative to the start of the SMB Header
	// (section 2.2.3.1). This constraint can cause this field to be a zero-length
	// field. This field SHOULD be set to zero by the client/server and MUST be ignored
	// by the server/client.
	Pad1 []types.UCHAR

	// Trans2_Parameters (variable): Transaction parameter bytes.
	Trans2_Parameters []types.UCHAR

	// Pad2 (variable): This field SHOULD be used as an array of padding bytes to align
	// the following field to a 4-byte boundary relative to the start of the SMB
	// Header. This constraint can cause this field to be a zero-length field. This
	// field SHOULD be set to zero by the client/server and MUST be ignored by the
	// server/client.
	Pad2 []types.UCHAR

	// Trans2_Data (variable): Transaction data bytes.
	Trans2_Data []types.UCHAR
}

// NewTransactionSecondaryRequest creates a new TransactionSecondaryRequest structure
//
// Returns:
// - A pointer to the new TransactionSecondaryRequest structure
func NewTransactionSecondaryRequest() *TransactionSecondaryRequest {
	c := &TransactionSecondaryRequest{
		// Parameters

		TotalParameterCount:   types.USHORT(0),
		TotalDataCount:        types.USHORT(0),
		ParameterCount:        types.USHORT(0),
		ParameterOffset:       types.USHORT(0),
		ParameterDisplacement: types.USHORT(0),
		DataCount:             types.USHORT(0),
		DataOffset:            types.USHORT(0),
		DataDisplacement:      types.USHORT(0),

		// Data

		Pad1:              []types.UCHAR{},
		Trans2_Parameters: []types.UCHAR{},
		Pad2:              []types.UCHAR{},
		Trans2_Data:       []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_TRANSACTION_SECONDARY)

	return c
}

// Marshal marshals the TransactionSecondaryRequest structure into a byte array
//
// Returns:
// - A byte array representing the TransactionSecondaryRequest structure
// - An error if the marshaling fails
func (c *TransactionSecondaryRequest) Marshal() ([]byte, error) {
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

	// Marshalling data Pad1
	rawDataContent = append(rawDataContent, c.Pad1...)

	// Marshalling data Trans2_Parameters
	rawDataContent = append(rawDataContent, c.Trans2_Parameters...)

	// Marshalling data Pad2
	rawDataContent = append(rawDataContent, c.Pad2...)

	// Marshalling data Trans2_Data
	rawDataContent = append(rawDataContent, c.Trans2_Data...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter TotalParameterCount
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.TotalParameterCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter TotalDataCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.TotalDataCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ParameterCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ParameterCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ParameterOffset
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ParameterOffset))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ParameterDisplacement
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ParameterDisplacement))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataOffset
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataOffset))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataDisplacement
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataDisplacement))
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
func (c *TransactionSecondaryRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter TotalParameterCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for TotalParameterCount")
	}
	c.TotalParameterCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter TotalDataCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for TotalDataCount")
	}
	c.TotalDataCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter ParameterCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for ParameterCount")
	}
	c.ParameterCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter ParameterOffset
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for ParameterOffset")
	}
	c.ParameterOffset = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter ParameterDisplacement
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for ParameterDisplacement")
	}
	c.ParameterDisplacement = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataCount")
	}
	c.DataCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataOffset
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataOffset")
	}
	c.DataOffset = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataDisplacement
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataDisplacement")
	}
	c.DataDisplacement = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Pad1
	if len(rawDataContent) < offset+int(c.ParameterOffset) {
		return offset, fmt.Errorf("rawParametersContent too short for Pad1")
	}
	c.Pad1 = rawDataContent[offset : offset+int(c.ParameterOffset)]
	offset += int(c.ParameterOffset)

	// Unmarshalling data Trans2_Parameters
	if len(rawDataContent) < offset+int(c.ParameterCount) {
		return offset, fmt.Errorf("rawParametersContent too short for Trans2_Parameters")
	}
	c.Trans2_Parameters = rawDataContent[offset : offset+int(c.ParameterCount)]
	offset += int(c.ParameterCount)

	// Unmarshalling data Pad2
	if len(rawDataContent) < offset+int(c.DataOffset) {
		return offset, fmt.Errorf("rawParametersContent too short for Pad2")
	}
	c.Pad2 = rawDataContent[offset : offset+int(c.DataOffset)]
	offset += int(c.DataOffset)

	// Unmarshalling data Trans2_Data
	if len(rawDataContent) < offset+int(c.DataCount) {
		return offset, fmt.Errorf("rawParametersContent too short for Trans2_Data")
	}
	c.Trans2_Data = rawDataContent[offset : offset+int(c.DataCount)]
	offset += int(c.DataCount)

	return offset, nil
}
