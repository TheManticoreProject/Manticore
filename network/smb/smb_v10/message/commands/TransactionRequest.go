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

// TransactionRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/57bfc115-fe29-4482-a0fe-a935757e0a4f
type TransactionRequest struct {
	command_interface.Command

	// Parameters

	// TotalParameterCount (2 bytes): The total number of transaction parameter bytes
	// the client expects to send to the server for this request. Parameter bytes for a
	// transaction are carried within the SMB_Data.Trans_Parameters field of the
	// SMB_COM_TRANSACTION request. If the size of all of the required
	// SMB_Data.Trans_Parameters for a given transaction causes the request to exceed
	// the MaxBufferSize established during session setup, then the client MUST NOT
	// send all of the parameters in one request. The client MUST break up the
	// parameters and send additional requests using the SMB_COM_TRANSACTION_SECONDARY
	// command to send the additional parameters. Any single request MUST NOT exceed
	// the MaxBufferSize established during session setup. The client indicates to the
	// server to expect additional parameters, and thus at least one
	// SMB_COM_TRANSACTION_SECONDARY, by setting ParameterCount (see following) to be
	// less than TotalParameterCount. See SMB_COM_TRANSACTION_SECONDARY for more
	// information.
	TotalParameterCount types.USHORT

	// TotalDataCount (2 bytes): The total number of transaction data bytes that the
	// client attempts to send to the server for this request. Data bytes of a
	// transaction are carried within the SMB_Data.Trans_Data field of the
	// SMB_COM_TRANSACTION request. If the size of all of the required
	// SMB_Data.Trans_Data for a given transaction causes the request to exceed the
	// MaxBufferSize established during session setup, then the client MUST NOT send
	// all of the data in one request. The client MUST break up the data and send
	// additional requests using the SMB_COM_TRANSACTION_SECONDARY command to send the
	// additional data. Any single request MUST NOT exceed the MaxBufferSize
	// established during session setup. The client indicates to the server to expect
	// additional data, and thus at least one SMB_COM_TRANSACTION_SECONDARY, by setting
	// DataCount (see following) to be less than TotalDataCount. See
	// SMB_COM_TRANSACTION_SECONDARY for more information.
	TotalDataCount types.USHORT

	// MaxParameterCount (2 bytes): The maximum number of SMB_Data.Trans_Parameters
	// bytes that the client accepts in the transaction response. The server MUST NOT
	// return more than this number of bytes in the SMB_Data.Trans_Parameters field of
	// the response.
	MaxParameterCount types.USHORT

	// MaxDataCount (2 bytes): The maximum number of SMB_Data.Trans_Data bytes that the
	// client accepts in the transaction response. The server MUST NOT return more than
	// this number of bytes in the SMB_Data.Trans_Data field.
	MaxDataCount types.USHORT

	// MaxSetupCount (1 byte): The maximum number of bytes that the client accepts in
	// the Setup field of the transaction response. The server MUST NOT return more
	// than this number of bytes in the Setup field.
	MaxSetupCount types.UCHAR

	// Reserved1 (1 byte): A padding byte. This field MUST be 0x00. Existing CIFS
	// implementations MAY combine this field with MaxSetupCount to form a USHORT. If
	// MaxSetupCount is defined as a USHORT, the high order byte MUST be 0x00.
	Reserved1 types.UCHAR

	// Flags (2 bytes): A set of bit flags that alter the behavior of the requested
	// operation. Unused bit fields MUST be set to zero by the client sending the
	// request, and MUST be ignored by the server receiving the request. The client can
	// set either or both of the following bit flags.
	Flags types.USHORT

	// Timeout (4 bytes): The value of this field MUST be the maximum number of
	// milliseconds that the server SHOULD wait for completion of the transaction
	// before generating a time-out and returning a response to the client. The client
	// SHOULD set this field to 0x00000000 to indicate that no time-out is expected. A
	// value of 0x00000000 indicates that the server returns an error if the resource
	// is not immediately available. If the operation does not complete within the
	// specified time, the server MAY abort the request and send a failure
	// response.<42>
	Timeout types.ULONG

	// Reserved2 (2 bytes): Reserved. This field MUST be 0x0000 in the client request.
	// The server MUST ignore the contents of this field.
	Reserved2 types.USHORT

	// ParameterCount (2 bytes): The number of transaction parameter bytes that the
	// client attempts to send to the server in this request. Parameter bytes for a
	// transaction are carried within the SMB_Data.Trans_Parameters field of the
	// SMB_COM_TRANSACTION request. If the transaction request fits within a single
	// SMB_COM_TRANSACTION request (the request size does not exceed MaxBufferSize),
	// then this value SHOULD be equal to TotalParameterCount. Otherwise, the sum of
	// the ParameterCount values in the primary and secondary transaction request
	// messages MUST be equal to the smallest TotalParameterCount value reported to the
	// server. If the value of this field is less than the value of
	// TotalParameterCount, then at least one SMB_COM_TRANSACTION_SECONDARY message
	// MUST be used to transfer the remaining transaction SMB_Data.Trans_Parameters
	// bytes. The ParameterCount field MUST be used to determine the number of
	// transaction SMB_Data.Trans_Parameters bytes that are contained within the
	// SMB_COM_TRANSACTION message.
	ParameterCount types.USHORT

	// ParameterOffset (2 bytes): This field MUST contain the number of bytes from the
	// start of the SMB Header to the start of the SMB_Data.Trans_Parameters field.
	// Server implementations MUST use this value to locate the transaction parameter
	// block within the request. If ParameterCount is zero, the client/server MAY set
	// this field to zero.<43>
	ParameterOffset types.USHORT

	// DataCount (2 bytes): The number of transaction data bytes that the client sends
	// to the server in this request. Data bytes for a transaction are carried within
	// the SMB_Data.Trans_Data field of the SMB_COM_TRANSACTION request. If the
	// transaction request fits within a single SMB_COM_TRANSACTION request (the
	// request size does not exceed MaxBufferSize), then this value SHOULD be equal to
	// TotalDataCount. Otherwise, the sum of the DataCount values in the primary and
	// secondary transaction request messages MUST be equal to the smallest
	// TotalDataCount value reported to the server. If the value of this field is less
	// than the value of TotalDataCount, then at least one
	// SMB_COM_TRANSACTION_SECONDARY message MUST be used to transfer the remaining
	// transaction SMB_Data.Trans_Data bytes. The DataCount field MUST be used to
	// determine the number of transaction SMB_Data.Trans_Data bytes contained within
	// the SMB_COM_TRANSACTION message.
	DataCount types.USHORT

	// DataOffset (2 bytes): This field MUST be the number of bytes from the start of
	// the SMB Header of the request to the start of the SMB_Data.Trans_Data field.
	// Server implementations MUST use this value to locate the transaction data block
	// within the request. If DataCount is zero, the client/server MAY<44> set this
	// field to zero.
	DataOffset types.USHORT

	// SetupCount (1 byte): This field MUST be the number of setup words that are
	// included in the transaction request.
	SetupCount types.UCHAR

	// Reserved3 (1 byte): A padding byte. This field MUST be 0x00. Existing CIFS
	// implementations MAY combine this field with SetupCount to form a USHORT. If
	// SetupCount is defined as a USHORT, the high order byte MUST be 0x00.
	Reserved3 types.UCHAR

	// Setup (variable): A variable-length array of setup words.
	Setup []types.USHORT

	// Data

	// Name (variable): The pathname of the mailslot or named pipe to which the
	// transaction subcommand applies, or a client-supplied identifier that provides a
	// name for the transaction. See the individual SMB_COM_TRANSACTION subprotocol
	// subcommand descriptions for information about the value set for each subcommand.
	// If the field is not specified in the section for the subcommands, the field
	// SHOULD be set to \pipe\. If SMB_FLAGS2_UNICODE is set in the Flags2 field of the
	// SMB Header (section 2.2.3.1) of the request, this field MUST be a
	// null-terminated array of 16-bit Unicode characters which MUST be aligned to
	// start on a 2-byte boundary from the start of the SMB header. Otherwise, this
	// field MUST be a null-terminated array of OEM characters. The Name field MUST be
	// the first field in this section.
	Name types.SMB_STRING

	// Pad1 (variable): This field SHOULD be used as an array of padding bytes to align
	// the following field to a 4-byte boundary relative to the start of the SMB Header
	// (section 2.2.3.1). This constraint can cause this field to be a zero-length
	// field. This field SHOULD be set to zero by the client/server and MUST be ignored
	// by the server/client.
	Pad1 []types.UCHAR

	// Trans_Parameters (variable): Transaction parameter bytes. See the individual
	// SMB_COM_TRANSACTION subprotocol subcommands descriptions for information on the
	// parameters sent for each subcommand.
	Trans_Parameters []types.UCHAR

	// Pad2 (variable): This field SHOULD be used as an array of padding bytes to align
	// the following field to a 4-byte boundary relative to the start of the SMB
	// Header. This constraint can cause this field to be a zero-length field. This
	// field SHOULD be set to zero by the client/server and MUST be ignored by the
	// server/client.
	Pad2 []types.UCHAR

	// Data (variable): Transaction data bytes. See the individual SMB_COM_TRANSACTION
	// subprotocol subcommands descriptions for information on the data sent for each
	// subcommand.
	Trans_Data []types.UCHAR
}

// NewTransactionRequest creates a new TransactionRequest structure
//
// Returns:
// - A pointer to the new TransactionRequest structure
func NewTransactionRequest() *TransactionRequest {
	c := &TransactionRequest{
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
		Setup:               []types.USHORT{},

		// Data
		Name:             types.SMB_STRING{},
		Pad1:             []types.UCHAR{},
		Trans_Parameters: []types.UCHAR{},
		Pad2:             []types.UCHAR{},
		Trans_Data:       []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_TRANSACTION)

	return c
}

// Marshal marshals the TransactionRequest structure into a byte array
//
// Returns:
// - A byte array representing the TransactionRequest structure
// - An error if the marshaling fails
func (c *TransactionRequest) Marshal() ([]byte, error) {
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
	bytesStream, err := c.Name.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Marshalling data Pad1
	rawDataContent = append(rawDataContent, c.Pad1...)

	// Marshalling data Trans_Parameters
	rawDataContent = append(rawDataContent, c.Trans_Parameters...)

	// Marshalling data Pad2
	rawDataContent = append(rawDataContent, c.Pad2...)

	// Marshalling data Trans_Data
	rawDataContent = append(rawDataContent, c.Trans_Data...)

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

	// Marshalling parameter Setup
	for _, setup := range c.Setup {
		buf2 = make([]byte, 2)
		binary.BigEndian.PutUint16(buf2, uint16(setup))
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
func (c *TransactionRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter Setup
	if len(rawParametersContent) < offset+2*int(c.SetupCount) {
		return offset, fmt.Errorf("rawParametersContent too short for Setup")
	}
	c.Setup = make([]types.USHORT, c.SetupCount)
	for i := 0; i < int(c.SetupCount); i++ {
		c.Setup[i] = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
		offset += 2
	}

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Name
	bytesRead, err = c.Name.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling data Pad1
	if len(rawDataContent) < offset+int(c.ParameterOffset) {
		return offset, fmt.Errorf("rawParametersContent too short for Pad1")
	}
	c.Pad1 = rawDataContent[offset : offset+int(c.ParameterOffset)]
	offset += int(c.ParameterOffset)

	// Unmarshalling data Trans2_Parameters
	if len(rawDataContent) < offset+int(c.ParameterCount) {
		return offset, fmt.Errorf("rawParametersContent too short for Trans_Parameters")
	}
	c.Trans_Parameters = rawDataContent[offset : offset+int(c.ParameterCount)]
	offset += int(c.ParameterCount)

	// Unmarshalling data Pad2
	if len(rawDataContent) < offset+int(c.DataOffset) {
		return offset, fmt.Errorf("rawParametersContent too short for Pad2")
	}
	c.Pad2 = rawDataContent[offset : offset+int(c.DataOffset)]
	offset += int(c.DataOffset)

	// Unmarshalling data Trans2_Data
	if len(rawDataContent) < offset+int(c.DataCount) {
		return offset, fmt.Errorf("rawParametersContent too short for Trans_Data")
	}
	c.Trans_Data = rawDataContent[offset : offset+int(c.DataCount)]
	offset += int(c.DataCount)

	return offset, nil
}
