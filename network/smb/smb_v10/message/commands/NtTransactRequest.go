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

// NtTransactRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/1e62725c-bb9e-4704-99a4-8db520a6f2da
type NtTransactRequest struct {
	command_interface.Command

	// Parameters

	// MaxSetupCount (1 byte): Maximum number of setup bytes that the client will
	// accept in the transaction reply. This field MUST be set as specified in the
	// subsections of Transaction Subcommands (section 2.2.5). The server MUST NOT
	// return more than this number of setup bytes.
	MaxSetupCount types.UCHAR

	// Reserved1 (2 bytes): Two padding bytes. This field MUST be 0x0000. This field is
	// used to align the next field to a 32-bit boundary.
	Reserved1 types.USHORT

	// TotalParameterCount (4 bytes): The total number of SMB_COM_NT_TRANSACT parameter
	// bytes to be sent in this transaction request. This value MAY be reduced in any
	// or all subsequent SMB_COM_NT_TRANSACT_SECONDARY requests that are part of the
	// same transaction. This value represents transaction parameter bytes, not SMB
	// parameter words. Transaction parameter bytes are carried in the SMB_Data block
	// of the SMB_COM_NT_TRANSACT request or in subsequent
	// SMB_COM_NT_TRANSACT_SECONDARY requests.
	TotalParameterCount types.ULONG

	// TotalDataCount (4 bytes): The total number of SMB_COM_NT_TRANSACT data bytes to
	// be sent in this transaction request. This value MAY be reduced in any or all
	// subsequent SMB_COM_NT_TRANSACT_SECONDARY requests that are part of the same
	// transaction. This value represents transaction data bytes, not SMB data bytes.
	TotalDataCount types.ULONG

	// MaxParameterCount (4 bytes): The maximum number of parameter bytes that the
	// client will accept in the transaction reply. This field MUST be set as specified
	// in the subsections of Transaction Subcommands. The server MUST NOT return more
	// than this number of parameter bytes.
	MaxParameterCount types.ULONG

	// MaxDataCount (4 bytes): The maximum number of data bytes that the client will
	// accept in the transaction reply. This field MUST be set as specified in the
	// subsections of Transaction Subcommands. The server MUST NOT return more than
	// this number of data bytes.
	MaxDataCount types.ULONG

	// ParameterCount (4 bytes): The number of transaction parameter bytes being sent
	// in this SMB message. If the transaction fits within a single SMB_COM_NT_TRANSACT
	// request, this value MUST be equal to TotalParameterCount. Otherwise, the sum of
	// the ParameterCount values in the primary and secondary transaction request
	// messages MUST be equal to the smallest TotalParameterCount value reported to the
	// server. If the value of this field is less than the value of
	// TotalParameterCount, then at least one SMB_COM_NT_TRANSACT_SECONDARY message
	// MUST be used to transfer the remaining parameter bytes.
	ParameterCount types.ULONG

	// ParameterOffset (4 bytes): The offset, in bytes, from the start of the
	// SMB_Header to the transaction parameter bytes. This MUST be the number of bytes
	// from the start of the SMB message to the start of the SMB_Data.Bytes.Parameters
	// field. Server implementations MUST use this value to locate the transaction
	// parameter block within the SMB message. If ParameterCount is zero, the
	// client/server MAY set this field to zero.<113>
	ParameterOffset types.ULONG

	// DataCount (4 bytes): The number of transaction data bytes being sent in this SMB
	// message. If the transaction fits within a single SMB_COM_NT_TRANSACT request,
	// then this value MUST be equal to TotalDataCount. Otherwise, the sum of the
	// DataCount values in the primary and secondary transaction request messages MUST
	// be equal to the smallest TotalDataCount value reported to the server. If the
	// value of this field is less than the value of TotalDataCount, then at least one
	// SMB_COM_NT_TRANSACT_SECONDARY message MUST be used to transfer the remaining
	// data bytes.
	DataCount types.ULONG

	// DataOffset (4 bytes): The offset, in bytes, from the start of the SMB Header
	// (section 2.2.3.1) to the transaction data bytes. This MUST be the number of
	// bytes from the start of the SMB message to the start of the SMB_Data.Bytes.Data
	// field. Server implementations MUST use this value to locate the transaction data
	// block within the SMB message. If DataCount is zero, the client/server MAY set
	// this field to zero.<114>
	DataOffset types.ULONG

	// SetupCount (1 byte): The number of setup words that are included in the
	// transaction request.
	SetupCount types.UCHAR

	// Function (2 bytes): The transaction subcommand code, which is used to identify
	// the operation to be performed by the server.
	Function types.USHORT

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

// NewNtTransactRequest creates a new NtTransactRequest structure
//
// Returns:
// - A pointer to the new NtTransactRequest structure
func NewNtTransactRequest() *NtTransactRequest {
	c := &NtTransactRequest{
		// Parameters
		MaxSetupCount:       types.UCHAR(0),
		Reserved1:           types.USHORT(0),
		TotalParameterCount: types.ULONG(0),
		TotalDataCount:      types.ULONG(0),
		MaxParameterCount:   types.ULONG(0),
		MaxDataCount:        types.ULONG(0),
		ParameterCount:      types.ULONG(0),
		ParameterOffset:     types.ULONG(0),
		DataCount:           types.ULONG(0),
		DataOffset:          types.ULONG(0),
		SetupCount:          types.UCHAR(0),
		Function:            types.USHORT(0),

		// Data
		Pad1:                []types.UCHAR{},
		NT_Trans_Parameters: []types.UCHAR{},
		Pad2:                []types.UCHAR{},
		NT_Trans_Data:       []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_NT_TRANSACT)

	return c
}

// Marshal marshals the NtTransactRequest structure into a byte array
//
// Returns:
// - A byte array representing the NtTransactRequest structure
// - An error if the marshaling fails
func (c *NtTransactRequest) Marshal() ([]byte, error) {
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

	// Marshalling parameter MaxSetupCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.MaxSetupCount))

	// Marshalling parameter Reserved1
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Reserved1))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter TotalParameterCount
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.TotalParameterCount))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter TotalDataCount
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.TotalDataCount))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter MaxParameterCount
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.MaxParameterCount))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter MaxDataCount
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.MaxDataCount))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter ParameterCount
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ParameterCount))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter ParameterOffset
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ParameterOffset))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter DataCount
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.DataCount))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter DataOffset
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.DataOffset))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter SetupCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.SetupCount))

	// Marshalling parameter Function
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Function))
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
func (c *NtTransactRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter MaxSetupCount
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("data too short for MaxSetupCount")
	}
	c.MaxSetupCount = types.UCHAR(rawParametersContent[offset])
	offset++

	// Unmarshalling parameter Reserved1
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved1")
	}
	c.Reserved1 = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

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

	// Unmarshalling parameter MaxParameterCount
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxParameterCount")
	}
	c.MaxParameterCount = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter MaxDataCount
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxDataCount")
	}
	c.MaxDataCount = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
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

	// Unmarshalling parameter SetupCount
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("data too short for SetupCount")
	}
	c.SetupCount = types.UCHAR(rawParametersContent[offset])
	offset++

	// Unmarshalling parameter Function
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Function")
	}
	c.Function = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

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
