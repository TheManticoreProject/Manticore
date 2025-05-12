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

// Transaction2Request
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/f7d148cd-e3d5-49ae-8b37-9633822bfeac
type Transaction2Request struct {
	command_interface.Command

	// Parameters

	// TotalParameterCount (2 bytes): The total number of SMB_COM_TRANSACTION2
	// parameter bytes to be sent in this transaction request. This value MAY be
	// reduced in any or all subsequent SMB_COM_TRANSACTION2_SECONDARY requests that
	// are part of the same transaction. This value represents transaction parameter
	// bytes, not SMB parameter words. Transaction parameter bytes are carried in the
	// SMB_Data block of the SMB_COM_TRANSACTION2 request.
	TotalParameterCount types.USHORT

	// TotalDataCount (2 bytes): The total number of SMB_COM_TRANSACTION2 data bytes to
	// be sent in this transaction request. This value MAY be reduced in any or all
	// subsequent SMB_COM_TRANSACTION2_SECONDARY requests that are part of the same
	// transaction. This value represents transaction data bytes, not SMB data bytes.
	TotalDataCount types.USHORT

	// MaxParameterCount (2 bytes): The maximum number of parameter bytes that the
	// client will accept in the transaction reply. The server MUST NOT return more
	// than this number of parameter bytes.
	MaxParameterCount types.USHORT

	// MaxDataCount (2 bytes): The maximum number of data bytes that the client will
	// accept in the transaction reply. The server MUST NOT return more than this
	// number of data bytes.
	MaxDataCount types.USHORT

	// MaxSetupCount (1 byte): The maximum number of setup bytes that the client will
	// accept in the transaction reply. The server MUST NOT return more than this
	// number of setup bytes.
	MaxSetupCount types.UCHAR

	// Reserved1 (1 byte): A padding byte. This field MUST be zero. Existing CIFS
	// implementations MAY combine this field with MaxSetupCount to form a USHORT. If
	// MaxSetupCount is defined as a USHORT, the high order byte MUST be 0x00.
	Reserved1 types.UCHAR

	// Flags (2 bytes): A set of bit flags that alter the behavior of the requested
	// operation. Unused bit fields MUST be set to zero by the client sending the
	// request, and MUST be ignored by the server receiving the request. The client MAY
	// set either or both of the following bit flags:
	Flags types.USHORT

	// Timeout (4 bytes): The number of milliseconds that the server waits for
	// completion of the transaction before generating a time-out. A value of
	// 0x00000000 indicates that the operation is not blocked.<66>
	Timeout types.ULONG

	// Reserved2 (2 bytes): Reserved. This field MUST be 0x0000 in the client request.
	// The server MUST ignore the contents of this field.
	Reserved2 types.USHORT

	// ParameterCount (2 bytes): The number of transaction parameter bytes being sent
	// in this SMB message. If the transaction fits within a single
	// SMB_COM_TRANSACTION2 request, then this value MUST be equal to
	// TotalParameterCount. Otherwise, the sum of the ParameterCount values in the
	// primary and secondary transaction request messages MUST be equal to the smallest
	// TotalParameterCount value reported to the server. If the value of this field is
	// less than the value of TotalParameterCount, then at least one
	// SMB_COM_TRANSACTION2_SECONDARY message MUST be used to transfer the remaining
	// parameter bytes. The ParameterCount field MUST be used to determine the number
	// of transaction parameter bytes contained within the SMB_COM_TRANSACTION2
	// message.
	ParameterCount types.USHORT

	// ParameterOffset (2 bytes): The offset, in bytes, from the start of the
	// SMB_Header to the transaction parameter bytes. This MUST be the number of bytes
	// from the start of the SMB message to the start of the SMB_Data.Bytes.Parameters
	// field. Server implementations MUST use this value to locate the transaction
	// parameter block within the SMB message. If ParameterCount is zero, the
	// client/server MAY set this field to zero.<67>
	ParameterOffset types.USHORT

	// DataCount (2 bytes): The number of transaction data bytes being sent in this SMB
	// message. If the transaction fits within a single SMB_COM_TRANSACTION2 request,
	// then this value MUST be equal to TotalDataCount. Otherwise, the sum of the
	// DataCount values in the primary and secondary transaction request messages MUST
	// be equal to the smallest TotalDataCount value reported to the server. If the
	// value of this field is less than the value of TotalDataCount, then at least one
	// SMB_COM_TRANSACTION2_SECONDARY message MUST be used to transfer the remaining
	// data bytes.
	DataCount types.USHORT

	// DataOffset (2 bytes): The offset, in bytes, from the start of the SMB Header
	// (section 2.2.3.1) to the transaction data bytes. This MUST be the number of
	// bytes from the start of the SMB message to the start of the SMB_Data.Bytes.Data
	// field. Server implementations MUST use this value to locate the transaction data
	// block within the SMB message. If DataCount is zero, the client/server MAY set
	// this field to zero.<68>
	DataOffset types.USHORT

	// SetupCount (1 byte): The number of setup words that are included in the
	// transaction request.
	SetupCount types.UCHAR

	// Reserved3 (1 byte): A padding byte. This field MUST be 0x00. Existing CIFS
	// implementations MAY combine this field with SetupCount to form a USHORT. If
	// SetupCount is defined as a USHORT, the high order byte MUST be0x00.
	Reserved3 types.UCHAR

	// Data

	// Name (1 byte): This field is not used in SMB_COM_TRANSACTION2 requests. This
	// field MUST be set to zero, and the server MUST ignore it on receipt.
	Name types.UCHAR

	// Pad1 (variable): This field MUST be used as an array of padding bytes to align
	// the following field to a 4-byte boundary relative to the start of the SMB Header
	// (section 2.2.3.1). This constraint can cause this field to be a zero-length
	// field. This field SHOULD be set to zero by the client/server and MUST be ignored
	// by the server/client.
	Pad1 []types.UCHAR

	// Trans2_Parameters (variable): Transaction parameter bytes. See the individual
	// SMB_COM_TRANSACTION2 subcommand descriptions for information on parameters sent
	// for each subcommand.
	Trans2_Parameters []types.UCHAR

	// Pad2 (variable): This field MUST be used as an array of padding bytes to align
	// the following field to a 4-byte boundary relative to the start of the SMB
	// Header. This constraint can cause this field to be a zero-length field. This
	// field SHOULD be set to zero by the client/server and MUST be ignored by the
	// server/client.
	Pad2 []types.UCHAR

	// Trans2_Data (variable): Transaction data bytes. See the individual
	// SMB_COM_TRANSACTION2 subcommand descriptions for information on data sent for
	// each subcommand.
	Trans2_Data []types.UCHAR
}

// NewTransaction2Request creates a new Transaction2Request structure
//
// Returns:
// - A pointer to the new Transaction2Request structure
func NewTransaction2Request() *Transaction2Request {
	c := &Transaction2Request{
		// Parameters
		TotalParameterCount: types.USHORT(0),
		TotalDataCount:      types.USHORT(0),
		MaxParameterCount:   types.USHORT(0),
		MaxDataCount:        types.USHORT(0),
		MaxSetupCount:       types.UCHAR(0),
		Reserved1:           types.UCHAR(0),
		Flags:               types.USHORT(0),
		Timeout:             types.ULONG(0),
		Reserved2:           types.USHORT(0),
		ParameterCount:      types.USHORT(0),
		ParameterOffset:     types.USHORT(0),
		DataCount:           types.USHORT(0),
		DataOffset:          types.USHORT(0),
		SetupCount:          types.UCHAR(0),
		Reserved3:           types.UCHAR(0),

		// Data
		Name:              types.UCHAR(0),
		Pad1:              []types.UCHAR{},
		Trans2_Parameters: []types.UCHAR{},
		Pad2:              []types.UCHAR{},
		Trans2_Data:       []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_TRANSACTION2)

	return c
}

// Marshal marshals the Transaction2Request structure into a byte array
//
// Returns:
// - A byte array representing the Transaction2Request structure
// - An error if the marshaling fails
func (c *Transaction2Request) Marshal() ([]byte, error) {
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

	// Marshalling data Name
	rawDataContent = append(rawDataContent, types.UCHAR(c.Name))

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

	// Marshalling parameter MaxParameterCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MaxParameterCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter MaxDataCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MaxDataCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter MaxSetupCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.MaxSetupCount))

	// Marshalling parameter Reserved1
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.Reserved1))

	// Marshalling parameter Flags
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Flags))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Timeout
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Timeout))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter Reserved2
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Reserved2))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ParameterCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ParameterCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ParameterOffset
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ParameterOffset))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataOffset
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataOffset))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter SetupCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.SetupCount))

	// Marshalling parameter Reserved3
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.Reserved3))

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
func (c *Transaction2Request) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter MaxParameterCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxParameterCount")
	}
	c.MaxParameterCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter MaxDataCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxDataCount")
	}
	c.MaxDataCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter MaxSetupCount
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("data too short for MaxSetupCount")
	}
	c.MaxSetupCount = types.UCHAR(rawParametersContent[offset])
	offset++

	// Unmarshalling parameter Reserved1
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("data too short for Reserved1")
	}
	c.Reserved1 = types.UCHAR(rawParametersContent[offset])
	offset++

	// Unmarshalling parameter Flags
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Flags")
	}
	c.Flags = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Timeout
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Timeout")
	}
	c.Timeout = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter Reserved2
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved2")
	}
	c.Reserved2 = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
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

	// Unmarshalling parameter SetupCount
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("data too short for SetupCount")
	}
	c.SetupCount = types.UCHAR(rawParametersContent[offset])
	offset++

	// Unmarshalling parameter Reserved3
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("data too short for Reserved3")
	}
	c.Reserved3 = types.UCHAR(rawParametersContent[offset])
	offset++

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Name
	if len(rawDataContent) < offset+1 {
		return offset, fmt.Errorf("rawParametersContent too short for Name")
	}
	c.Name = types.UCHAR(rawDataContent[offset])
	offset++

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
