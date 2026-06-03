package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// NtTransactResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/dd00c842-2398-412f-b21d-bf5074a9a1c4
type NtTransactResponse struct {
	command_interface.Command

	// Parameters

	// Reserved1 (3 bytes): Reserved. This field MUST be 0x000000 in the server
	// response. The client MUST ignore the contents of this field.
	Reserved1 [3]types.UCHAR

	// TotalParameterCount (4 bytes): The total number of SMB_COM_NT_TRANSACT parameter
	// bytes to be sent in this transaction response. This value MAY be reduced in any
	// or all subsequent SMB_COM_NT_TRANSACT responses that are part of the same
	// transaction. This value represents transaction parameter bytes, not SMB
	// parameter words. Transaction parameter bytes are carried within the SMB_Data
	// block.
	TotalParameterCount types.ULONG

	// TotalDataCount (4 bytes): The total number of SMB_COM_NT_TRANSACT data bytes to
	// be sent in this transaction response. This value MAY be reduced in any or all
	// subsequent SMB_COM_NT_TRANSACT responses that are part of the same transaction.
	// This value represents transaction data bytes, not SMB data bytes.
	TotalDataCount types.ULONG

	// ParameterCount (4 bytes): The number of transaction parameter bytes being sent
	// in this SMB message. If the transaction fits within a single SMB_COM_NT_TRANSACT
	// response, then this value MUST be equal to TotalParameterCount. Otherwise, the
	// sum of the ParameterCount values in the transaction response messages MUST be
	// equal to the smallest TotalParameterCount value reported by the server.
	ParameterCount types.ULONG

	// ParameterOffset (4 bytes): The offset, in bytes, from the start of the
	// SMB_Header to the transaction parameter bytes. This MUST be the number of bytes
	// from the start of the SMB message to the start of the
	// SMB_Data.Bytes.Parameters field. Server implementations MUST use this value to
	// locate the transaction parameter block within the SMB message. If
	// ParameterCount is zero, the client/server MAY set this field to zero.
	ParameterOffset types.ULONG

	// ParameterDisplacement (4 bytes): The offset, relative to all of the transaction
	// parameter bytes in this transaction response, at which this block of parameter
	// bytes MUST be placed. This value can be used by the client to correctly
	// reassemble the transaction parameters even if the SMB response messages are
	// received out of order.
	ParameterDisplacement types.ULONG

	// DataCount (4 bytes): The number of transaction data bytes being sent in this SMB
	// message. If the transaction fits within a single SMB_COM_NT_TRANSACT response,
	// then this value MUST be equal to TotalDataCount. Otherwise, the sum of the
	// DataCount values in the transaction response messages MUST be equal to the
	// smallest TotalDataCount value reported by the server.
	DataCount types.ULONG

	// DataOffset (4 bytes): The offset, in bytes, from the start of the SMB_Header to
	// the transaction data bytes. This MUST be the number of bytes from the start of
	// the SMB message to the start of the SMB_Data.Bytes.Data field. Server
	// implementations MUST use this value to locate the transaction data block within
	// the SMB message. If DataCount is zero, the client/server MAY set this field to
	// zero.
	DataOffset types.ULONG

	// DataDisplacement (4 bytes): The offset, relative to all of the transaction data
	// bytes in this transaction response, at which this block of data bytes MUST be
	// placed. This value can be used by the client to correctly reassemble the
	// transaction data even if the SMB response messages are received out of order.
	DataDisplacement types.ULONG

	// SetupCount (1 byte): The number of Setup words that are included in the
	// transaction response.
	SetupCount types.UCHAR

	// Setup (variable): An array of two-byte words that provides transaction results
	// from the server. The size and content of the array are specific to individual
	// subcommand.
	Setup []types.USHORT

	// Data

	// Pad1 (variable): This field SHOULD be used as an array of padding bytes to align
	// the following field to a 4-byte boundary relative to the start of the SMB
	// Header. This constraint can cause this field to be a zero-length field. This
	// field SHOULD be set to zero by the client/server and MUST be ignored by the
	// server/client.
	Pad1 []types.UCHAR

	// Parameters (variable): Transaction parameter bytes. See the individual
	// SMB_COM_NT_TRANSACT subcommand descriptions for information on parameters
	// returned by the server for each subcommand.
	Parameters []types.UCHAR

	// Pad2 (variable): This field SHOULD be used as an array of padding bytes to align
	// the following field to a 4-byte boundary relative to the start of the SMB
	// Header. This constraint can cause this field to be a zero-length field. This
	// field SHOULD be set to zero by the client/server and MUST be ignored by the
	// server/client.
	Pad2 []types.UCHAR

	// Data (variable): Transaction data bytes. See the individual SMB_COM_NT_TRANSACT
	// subcommand descriptions for information on data returned by the server for each
	// subcommand.
	Data []types.UCHAR
}

// NewNtTransactResponse creates a new NtTransactResponse structure
//
// Returns:
// - A pointer to the new NtTransactResponse structure
func NewNtTransactResponse() *NtTransactResponse {
	c := &NtTransactResponse{
		// Parameters
		Reserved1:             [3]types.UCHAR{0, 0, 0},
		TotalParameterCount:   types.ULONG(0),
		TotalDataCount:        types.ULONG(0),
		ParameterCount:        types.ULONG(0),
		ParameterOffset:       types.ULONG(0),
		ParameterDisplacement: types.ULONG(0),
		DataCount:             types.ULONG(0),
		DataOffset:            types.ULONG(0),
		DataDisplacement:      types.ULONG(0),
		SetupCount:            types.UCHAR(0),
		Setup:                 []types.USHORT{},

		// Data
		Pad1:       []types.UCHAR{},
		Parameters: []types.UCHAR{},
		Pad2:       []types.UCHAR{},
		Data:       []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_NT_TRANSACT)

	return c
}

// Marshal marshals the NtTransactResponse structure into a byte array
//
// Returns:
// - A byte array representing the NtTransactResponse structure
// - An error if the marshaling fails
func (c *NtTransactResponse) Marshal() ([]byte, error) {
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

	// Marshalling data Parameters
	rawDataContent = append(rawDataContent, c.Parameters...)

	// Marshalling data Pad2
	rawDataContent = append(rawDataContent, c.Pad2...)

	// Marshalling data Data
	rawDataContent = append(rawDataContent, c.Data...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter Reserved1
	rawParametersContent = append(rawParametersContent, c.Reserved1[0], c.Reserved1[1], c.Reserved1[2])

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

	// Marshalling parameter SetupCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.SetupCount))

	// Marshalling parameter Setup
	for _, setupWord := range c.Setup {
		buf2 := make([]byte, 2)
		binary.BigEndian.PutUint16(buf2, uint16(setupWord))
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
func (c *NtTransactResponse) Unmarshal(rawData []byte) (int, error) {
	offset := 0

	// Create the Parameters structure if it is nil
	if c.GetParameters() == nil {
		c.SetParameters(parameters.NewParameters())
	}
	// Create the Data structure if it is nil
	if c.GetData() == nil {
		c.SetData(data.NewData())
	}

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(rawData)
	if err != nil {
		return 0, err
	}
	rawParametersContent := c.GetParameters().GetBytes()
	_, err = c.GetData().Unmarshal(rawData[bytesRead:])
	if err != nil {
		return 0, err
	}
	rawDataContent := c.GetData().GetBytes()

	// If the parameters and data are empty, this is the interim server response or a
	// response carrying an error code in the SMB Header Status field. Both have empty
	// Parameter and Data sections (WordCount and ByteCount are zero).
	if len(rawParametersContent) == 0 && len(rawDataContent) == 0 {
		return 0, nil
	}

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter Reserved1
	if len(rawParametersContent) < offset+3 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved1")
	}
	c.Reserved1 = [3]types.UCHAR{
		types.UCHAR(rawParametersContent[offset]),
		types.UCHAR(rawParametersContent[offset+1]),
		types.UCHAR(rawParametersContent[offset+2]),
	}
	offset += 3

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

	// Unmarshalling parameter SetupCount
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("rawParametersContent too short for SetupCount")
	}
	c.SetupCount = types.UCHAR(rawParametersContent[offset])
	offset++

	// Unmarshalling parameter Setup
	if len(rawParametersContent) < offset+int(c.SetupCount)*2 {
		return offset, fmt.Errorf("rawParametersContent too short for Setup")
	}
	c.Setup = make([]types.USHORT, c.SetupCount)
	for i := 0; i < int(c.SetupCount); i++ {
		c.Setup[i] = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
		offset += 2
	}

	// Then unmarshal the data
	// The Parameters and Data blocks are located within the SMB_Data.Bytes block.
	// ParameterOffset and DataOffset are measured from the start of the SMB Header, so
	// their position relative to the start of the SMB_Data.Bytes block is derived by
	// subtracting the bytes that precede it: the fixed SMB header, the marshalled
	// SMB_Parameters block (WordCount byte + parameter words, == bytesRead), and the
	// 2-byte ByteCount field of the SMB_Data block.
	dataBlockStart := header.SMB_HEADER_SIZE + bytesRead + 2

	offset = 0

	// Unmarshalling data Pad1 (gap between the start of the data block and Parameters)
	pad1Len := 0
	if c.ParameterCount > 0 {
		paramRel := int(c.ParameterOffset) - dataBlockStart
		if paramRel < offset || len(rawDataContent) < paramRel {
			return offset, fmt.Errorf("invalid ParameterOffset for Pad1")
		}
		pad1Len = paramRel - offset
	}
	c.Pad1 = rawDataContent[offset : offset+pad1Len]
	offset += pad1Len

	// Unmarshalling data Parameters
	if len(rawDataContent) < offset+int(c.ParameterCount) {
		return offset, fmt.Errorf("rawDataContent too short for Parameters")
	}
	c.Parameters = rawDataContent[offset : offset+int(c.ParameterCount)]
	offset += int(c.ParameterCount)

	// Unmarshalling data Pad2 (gap between Parameters and Data)
	pad2Len := 0
	if c.DataCount > 0 {
		dataRel := int(c.DataOffset) - dataBlockStart
		if dataRel < offset || len(rawDataContent) < dataRel {
			return offset, fmt.Errorf("invalid DataOffset for Pad2")
		}
		pad2Len = dataRel - offset
	}
	c.Pad2 = rawDataContent[offset : offset+pad2Len]
	offset += pad2Len

	// Unmarshalling data Data
	if len(rawDataContent) < offset+int(c.DataCount) {
		return offset, fmt.Errorf("rawDataContent too short for Data")
	}
	c.Data = rawDataContent[offset : offset+int(c.DataCount)]
	offset += int(c.DataCount)

	return offset, nil
}
