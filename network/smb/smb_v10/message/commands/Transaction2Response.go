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

// Transaction2Response
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/216e606a-eee1-4c3f-b88e-0eb14dc380b2
// The SMB_COM_TRANSACTION2 response has two possible formats.
// The standard (final) format is used to return the results of the completed transaction.
// A shortened interim response message is sent following the initial SMB_COM_TRANSACTION2
// request if secondary request messages (SMB_COM_TRANSACTION2_SECONDARY) are pending. The
// interim response is an SMB Header with empty Parameter and Data sections (WordCount and
// ByteCount both zero); it is represented here by leaving all fields zero, which marshals
// to an empty parameter/data block.
type Transaction2Response struct {
	command_interface.Command

	// Parameters

	// TotalParameterCount (2 bytes): The total number of SMB_COM_TRANSACTION2 parameter
	// bytes to be sent in this transaction response. This value can be reduced in any or
	// all subsequent SMB_COM_TRANSACTION2 responses that are part of the same
	// transaction. This value represents transaction parameter bytes, not SMB parameter
	// words.
	TotalParameterCount types.USHORT

	// TotalDataCount (2 bytes): The total number of SMB_COM_TRANSACTION2 data bytes to be
	// sent in this transaction response. This value MAY be reduced in any or all
	// subsequent SMB_COM_TRANSACTION2 responses that are part of the same transaction.
	// This value represents transaction data bytes, not SMB data bytes.
	TotalDataCount types.USHORT

	// Reserved1 (2 bytes): Reserved. This field MUST be 0x0000. The receiver MUST ignore
	// the contents of this field.
	Reserved1 types.USHORT

	// ParameterCount (2 bytes): The number of transaction parameter bytes being sent in
	// this SMB message. If the transaction fits within a single SMB_COM_TRANSACTION2
	// response, this value MUST be equal to TotalParameterCount.
	ParameterCount types.USHORT

	// ParameterOffset (2 bytes): The offset, in bytes, from the start of the SMB_Header
	// to the transaction parameter bytes. If ParameterCount is zero, this field MAY be
	// set to zero.
	ParameterOffset types.USHORT

	// ParameterDisplacement (2 bytes): The offset relative to all of the transaction
	// parameter bytes in this transaction response at which this block of parameter bytes
	// MUST be placed. Used by the client to reassemble parameters received out of order.
	ParameterDisplacement types.USHORT

	// DataCount (2 bytes): The number of transaction data bytes being sent in this SMB
	// message. If the transaction fits within a single SMB_COM_TRANSACTION2 response,
	// then this value MUST be equal to TotalDataCount.
	DataCount types.USHORT

	// DataOffset (2 bytes): The offset, in bytes, from the start of the SMB_Header to the
	// transaction data bytes. If DataCount is zero, this field MAY be set to zero.
	DataOffset types.USHORT

	// DataDisplacement (2 bytes): The offset relative to all of the transaction data
	// bytes in this transaction response at which this block of data bytes MUST be
	// placed. Used by the client to reassemble data received out of order.
	DataDisplacement types.USHORT

	// SetupCount (1 byte): The number of setup words that are included in the transaction
	// response.
	SetupCount types.UCHAR

	// Reserved2 (1 byte): A padding byte. This field MUST be 0x00. If SetupCount is
	// defined as a USHORT, the high order byte MUST be 0x00.
	Reserved2 types.UCHAR

	// Setup (variable): An array of two-byte words that provides transaction results from
	// the server. The size and content of the array are specific to individual
	// subcommands.
	Setup []types.USHORT

	// Data

	// Pad1 (variable): An array of padding bytes used to align the following field to a
	// 4-byte boundary relative to the start of the SMB Header. This constraint can cause
	// this field to be a zero-length field.
	Pad1 []types.UCHAR

	// Trans2_Parameters (variable): Transaction parameter bytes. See the individual
	// SMB_COM_TRANSACTION2 subcommand descriptions for information on parameters returned
	// by the server for each subcommand.
	Trans2_Parameters []types.UCHAR

	// Pad2 (variable): An array of padding bytes used to align the following field to a
	// 4-byte boundary relative to the start of the SMB Header. This constraint can cause
	// this field to be a zero-length field.
	Pad2 []types.UCHAR

	// Trans2_Data (variable): Transaction data bytes. See the individual
	// SMB_COM_TRANSACTION2 subcommand descriptions for information on data returned by the
	// server for each subcommand.
	Trans2_Data []types.UCHAR
}

// NewTransaction2Response creates a new Transaction2Response structure
//
// Returns:
// - A pointer to the new Transaction2Response structure
func NewTransaction2Response() *Transaction2Response {
	c := &Transaction2Response{
		// Parameters
		TotalParameterCount:   types.USHORT(0),
		TotalDataCount:        types.USHORT(0),
		Reserved1:             types.USHORT(0),
		ParameterCount:        types.USHORT(0),
		ParameterOffset:       types.USHORT(0),
		ParameterDisplacement: types.USHORT(0),
		DataCount:             types.USHORT(0),
		DataOffset:            types.USHORT(0),
		DataDisplacement:      types.USHORT(0),
		SetupCount:            types.UCHAR(0),
		Reserved2:             types.UCHAR(0),
		Setup:                 []types.USHORT{},

		// Data
		Pad1:              []types.UCHAR{},
		Trans2_Parameters: []types.UCHAR{},
		Pad2:              []types.UCHAR{},
		Trans2_Data:       []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_TRANSACTION2)

	return c
}

// Marshal marshals the Transaction2Response structure into a byte array
//
// Returns:
// - A byte array representing the Transaction2Response structure
// - An error if the marshaling fails
func (c *Transaction2Response) Marshal() ([]byte, error) {
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
	binary.LittleEndian.PutUint16(buf2, uint16(c.TotalParameterCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter TotalDataCount
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.TotalDataCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Reserved1
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.Reserved1))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ParameterCount
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.ParameterCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ParameterOffset
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.ParameterOffset))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ParameterDisplacement
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.ParameterDisplacement))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataCount
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.DataCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataOffset
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.DataOffset))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataDisplacement
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.DataDisplacement))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter SetupCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.SetupCount))

	// Marshalling parameter Reserved2
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.Reserved2))

	// Marshalling parameter Setup
	for _, setupWord := range c.Setup {
		buf2 = make([]byte, 2)
		binary.LittleEndian.PutUint16(buf2, uint16(setupWord))
		rawParametersContent = append(rawParametersContent, buf2...)
	}

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
func (c *Transaction2Response) Unmarshal(data []byte) (int, error) {
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

	// If the parameters and data are empty, this is either the interim response or a
	// response carrying an error code in the SMB Header Status field.
	if len(rawParametersContent) == 0 && len(rawDataContent) == 0 {
		return 0, nil
	}

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter TotalParameterCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for TotalParameterCount")
	}
	c.TotalParameterCount = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter TotalDataCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for TotalDataCount")
	}
	c.TotalDataCount = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Reserved1
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved1")
	}
	c.Reserved1 = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter ParameterCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for ParameterCount")
	}
	c.ParameterCount = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter ParameterOffset
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for ParameterOffset")
	}
	c.ParameterOffset = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter ParameterDisplacement
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for ParameterDisplacement")
	}
	c.ParameterDisplacement = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataCount")
	}
	c.DataCount = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataOffset
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataOffset")
	}
	c.DataOffset = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataDisplacement
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataDisplacement")
	}
	c.DataDisplacement = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter SetupCount
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("rawParametersContent too short for SetupCount")
	}
	c.SetupCount = types.UCHAR(rawParametersContent[offset])
	offset++

	// Unmarshalling parameter Reserved2
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved2")
	}
	c.Reserved2 = types.UCHAR(rawParametersContent[offset])
	offset++

	// Unmarshalling parameter Setup
	c.Setup = make([]types.USHORT, 0, int(c.SetupCount))
	for i := 0; i < int(c.SetupCount); i++ {
		if len(rawParametersContent) < offset+2 {
			return offset, fmt.Errorf("rawParametersContent too short for Setup")
		}
		c.Setup = append(c.Setup, types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset:offset+2])))
		offset += 2
	}

	// Then unmarshal the data
	//
	// The SMB_Data.Bytes block is laid out as Pad1 | Trans2_Parameters | Pad2 |
	// Trans2_Data. ParameterCount and DataCount give the exact lengths of the two payload
	// runs; the lengths of the alignment padding fields Pad1 and Pad2 are derived from the
	// documented field offsets (ParameterOffset/DataOffset, both measured from the start
	// of the SMB Header) by subtracting the bytes that precede each payload run. This
	// keeps the parse robust to zero-length or present padding.
	offset = 0

	// The number of bytes that precede the SMB_Data.Bytes block within the SMB message:
	// the SMB_Parameters block (WordCount byte + parameter words, == bytesRead) plus the
	// 2-byte ByteCount field of the SMB_Data block.
	bytesBeforeData := bytesRead + 2

	// Unmarshalling data Pad1
	pad1Length := 0
	if c.ParameterCount > 0 && int(c.ParameterOffset) > bytesBeforeData {
		pad1Length = int(c.ParameterOffset) - bytesBeforeData
	}
	if len(rawDataContent) < offset+pad1Length {
		return offset, fmt.Errorf("rawDataContent too short for Pad1")
	}
	c.Pad1 = rawDataContent[offset : offset+pad1Length]
	offset += pad1Length

	// Unmarshalling data Trans2_Parameters
	if len(rawDataContent) < offset+int(c.ParameterCount) {
		return offset, fmt.Errorf("rawDataContent too short for Trans2_Parameters")
	}
	c.Trans2_Parameters = rawDataContent[offset : offset+int(c.ParameterCount)]
	offset += int(c.ParameterCount)

	// Unmarshalling data Pad2
	pad2Length := 0
	if c.DataCount > 0 && int(c.DataOffset) > bytesBeforeData+offset {
		pad2Length = int(c.DataOffset) - (bytesBeforeData + offset)
	}
	if len(rawDataContent) < offset+pad2Length {
		return offset, fmt.Errorf("rawDataContent too short for Pad2")
	}
	c.Pad2 = rawDataContent[offset : offset+pad2Length]
	offset += pad2Length

	// Unmarshalling data Trans2_Data
	if len(rawDataContent) < offset+int(c.DataCount) {
		return offset, fmt.Errorf("rawDataContent too short for Trans2_Data")
	}
	c.Trans2_Data = rawDataContent[offset : offset+int(c.DataCount)]
	offset += int(c.DataCount)

	return offset, nil
}
