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

// The parameter block sizes of the two forms of the request. WordCount is 0x0C
// without OffsetHigh and 0x0E with it, and the four extra bytes move the data
// four bytes further from the header ([MS-CIFS] section 2.2.4.43.1).
const (
	writeAndxWordsSize   = 2 * 0x0C
	writeAndxWordsSize64 = 2 * 0x0E
)

// writeAndxDataOffset and writeAndxDataOffset64 are where this request's data
// begins, in bytes from the start of the SMB header, for a request at the front
// of the message: header + WordCount(1) + words + ByteCount(2) + Pad(1).
//
// A request batched later in an AndX chain adds its chain offset, which is what
// SetChainOffset records.
const (
	writeAndxDataOffset   = header.SMB_HEADER_SIZE + 1 + writeAndxWordsSize + 2 + 1
	writeAndxDataOffset64 = header.SMB_HEADER_SIZE + 1 + writeAndxWordsSize64 + 2 + 1
)

// WriteAndxRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/a66126d2-a1db-446b-8736-b9f5559c49bd
type WriteAndxRequest struct {
	command_interface.Command

	// Parameters

	// FID (2 bytes): This field MUST be a valid FID indicating the file to which the
	// data SHOULD be written.
	FID types.USHORT

	// Offset (4 bytes): If WordCount is 0x0C, this field represents a 32-bit offset,
	// measured in bytes, of where the write SHOULD start relative to the beginning of
	// the file. If WordCount is 0xE, this field represents the lower 32 bits of a
	// 64-bit offset.
	Offset types.ULONG

	// Timeout (4 bytes): This field is the time-out, in milliseconds, to wait for the
	// write to complete. This field is used only when writing to a named pipe or an
	// I/O device. It does not apply and MUST be 0x00000000 when writing to a regular
	// file.
	Timeout types.ULONG

	// WriteMode (2 bytes): A 16-bit field containing flags defined as follows:
	WriteMode types.USHORT

	// Remaining (2 bytes): This field is an advisory field telling the server
	// approximately how many bytes are to be written to this file before the next
	// non-write operation. It SHOULD include the number of bytes to be written by this
	// request. The server MAY either ignore this field or use it to perform
	// optimizations. If a pipe write spans multiple requests, the client SHOULD set
	// this field to the number of bytes remaining to be written.
	Remaining types.USHORT

	// Reserved (2 bytes): This field MUST be 0x0000.
	Reserved types.USHORT

	// DataLength (2 bytes): This field is the number of bytes included in the SMB_Data
	// that are to be written to the file.
	DataLength types.USHORT

	// The DataOffset field can be used to relocate the SMB_Data.Bytes.Data block to
	// the end of the message, even if the message is a multi-part AndX chain. If the
	// SMB_Data.Bytes.Data block is relocated, the contents of SMB_Data.Bytes will not
	// be contiguous.
	DataOffset types.USHORT

	// OffsetHigh (4 bytes): This field is optional. If WordCount is 0x0C, this field
	// is not included in the request. If WordCount is 0x0E, this field represents the
	// upper 32 bits of a 64-bit offset, measured in bytes, of where the write SHOULD
	// start relative to the beginning of the file.
	OffsetHigh types.ULONG

	// Data

	// Pad (1 byte): Padding byte that MUST be ignored.
	Pad types.UCHAR

	// Data (variable): The raw bytes to be written to the file.
	Data []types.UCHAR

	// chainOffset is how far past the SMB header this request begins, which is
	// non-zero only when it is batched behind another command. DataOffset is
	// measured from the header rather than from the request, so the two have to
	// be added.
	chainOffset int
}

// SetChainOffset records where this request begins, so DataOffset describes where
// its data actually is.
//
// Parameters:
//   - offset: bytes from the end of the SMB header to the start of this request
func (c *WriteAndxRequest) SetChainOffset(offset int) {
	c.chainOffset = offset
}

// dataOffset is the offset this request's data sits at, from the start of the SMB
// header.
//
// The 64-bit-offset form carries OffsetHigh in its parameter block, which pushes
// the data four bytes further out. Marshal emits that word only for a non-zero
// OffsetHigh, so the same condition decides the offset.
func (c *WriteAndxRequest) dataOffset() int {
	if c.OffsetHigh != 0 {
		return writeAndxDataOffset64 + c.chainOffset
	}
	return writeAndxDataOffset + c.chainOffset
}

// NewWriteAndxRequest creates a new WriteAndxRequest structure
//
// Returns:
// - A pointer to the new WriteAndxRequest structure
func NewWriteAndxRequest() *WriteAndxRequest {
	c := &WriteAndxRequest{
		// Parameters
		FID:        types.USHORT(0),
		Offset:     types.ULONG(0),
		Timeout:    types.ULONG(0),
		WriteMode:  types.USHORT(0),
		Remaining:  types.USHORT(0),
		Reserved:   types.USHORT(0),
		DataLength: types.USHORT(0),
		DataOffset: types.USHORT(0),
		OffsetHigh: types.ULONG(0),

		// Data
		Pad:  types.UCHAR(0),
		Data: []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_WRITE_ANDX)

	return c
}

// IsAndX returns true if the command is an AndX
func (c *WriteAndxRequest) IsAndX() bool {
	return true
}

// Marshal marshals the WriteAndxRequest structure into a byte array
//
// Returns:
// - A byte array representing the WriteAndxRequest structure
// - An error if the marshaling fails
func (c *WriteAndxRequest) Marshal() ([]byte, error) {
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

	// Marshalling data Pad
	rawDataContent = append(rawDataContent, types.UCHAR(c.Pad))

	// Marshalling data Data
	rawDataContent = append(rawDataContent, c.Data...)

	// The three fields that describe the data are derived from the data, not
	// taken from the caller: they are the only way a receiver can find and size
	// it, and a caller computing them by hand has to know this command's own
	// wire layout to do it.
	//
	// The length spans two words. [MS-SMB] section 2.2.4.3.1 allocates the CIFS
	// Reserved field as DataLengthHigh, and requires a write of 0x10000 bytes or
	// more to put the low half in DataLength and the high half there. Below that
	// the high half is zero, which is also what [MS-CIFS] section 2.2.4.43.1
	// requires of Reserved, so writing it unconditionally is right either way.
	c.DataLength = types.USHORT(len(c.Data) & 0xFFFF)
	c.Reserved = types.USHORT(len(c.Data) >> 16)
	c.DataOffset = types.USHORT(c.dataOffset())

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter FID
	buf2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.FID))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Offset
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(c.Offset))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter Timeout
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(c.Timeout))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter WriteMode
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.WriteMode))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Remaining
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.Remaining))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Reserved
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.Reserved))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataLength
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.DataLength))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataOffset
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.DataOffset))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter OffsetHigh
	if c.OffsetHigh != 0 {
		buf4 = make([]byte, 4)
		binary.LittleEndian.PutUint32(buf4, uint32(c.OffsetHigh))
		rawParametersContent = append(rawParametersContent, buf4...)
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

// DataLengthHigh is the high 16 bits of the write's length.
//
// [MS-SMB] section 2.2.4.3.1 allocates the CIFS Reserved field for it under
// CAP_LARGE_WRITEX. The field keeps its CIFS name on the struct because that is
// what it is called when the capability is not in force, and this accessor names
// what it means when it is.
//
// Returns:
//   - The high 16 bits of the number of data bytes the request carries
func (c *WriteAndxRequest) DataLengthHigh() types.USHORT {
	return c.Reserved
}

// Unmarshal unmarshals a byte array into the command structure
//
// Parameters:
// - data: The byte array to unmarshal
//
// Returns:
// - The number of bytes unmarshalled
func (c *WriteAndxRequest) Unmarshal(rawData []byte) (int, error) {
	// Initialize the Parameters structure if it is nil to avoid a nil
	// pointer dereference when Unmarshal is called on a freshly constructed value.
	if c.GetParameters() == nil {
		c.SetParameters(parameters.NewParameters())
	}
	// Initialize the Data structure if it is nil for the same reason.
	if c.GetData() == nil {
		c.SetData(data.NewData())
	}
	offset := 0

	// First unmarshal the parameters
	bytesRead, err := c.GetParameters().Unmarshal(rawData)
	if err != nil {
		return 0, err
	}
	rawParametersContent := c.GetParameters().GetBytes()

	// Then the data block, for the ByteCount the sender put on the wire — but not
	// for the data itself, and not fatally.
	//
	// ByteCount cannot be trusted to bound this command's data. [MS-CIFS] section
	// 2.2.4.43.1 describes relocating the data to the end of an AndX chain and
	// keeps ByteCount at "1 + SMB_Parameters.Words.DataLength", counting bytes
	// that are not in the block at all; and a write of 0x10000 bytes or more
	// cannot state its length in a USHORT. Windows servers "ignore the ByteCount
	// field, and calculate the number of bytes to be written as DataLength |
	// DataLengthHigh <<16" ([MS-SMB] section 3.3.5.8), which is what happens
	// below. So a ByteCount that overruns the buffer is the sender describing a
	// relocated block, not a truncated message.
	_, _ = c.GetData().Unmarshal(rawData[bytesRead:])

	// If the parameters are empty, this is a response containing an error code in
	// the SMB Header Status field
	if len(rawParametersContent) == 0 {
		return 0, nil
	}

	// First unmarshal the parameters
	offset = 0
	if c.IsAndX() {
		offset += 4
	}

	// Unmarshalling parameter FID
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for FID")
	}
	c.FID = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Offset
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Offset")
	}
	c.Offset = types.ULONG(binary.LittleEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter Timeout
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Timeout")
	}
	c.Timeout = types.ULONG(binary.LittleEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter WriteMode
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for WriteMode")
	}
	c.WriteMode = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Remaining
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Remaining")
	}
	c.Remaining = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved")
	}
	c.Reserved = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataLength
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataLength")
	}
	c.DataLength = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataOffset
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataOffset")
	}
	c.DataOffset = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter OffsetHigh
	if c.GetParameters().WordCount == 0x0E {
		if len(rawParametersContent) < offset+4 {
			return offset, fmt.Errorf("rawParametersContent too short for OffsetHigh")
		}
		c.OffsetHigh = types.ULONG(binary.LittleEndian.Uint32(rawParametersContent[offset : offset+4]))
		offset += 4
	}

	// Then locate the data.
	//
	// The length spans two fields. [MS-SMB] section 2.2.4.3.1 allocates the CIFS
	// Reserved field as DataLengthHigh, which is the only way a write above
	// 0xFFFF bytes can describe itself, and [MS-CIFS] section 2.2.4.43.1 requires
	// Reserved to be zero, so a non-zero value is a length rather than a client
	// leaving a field dirty. Reading DataLength alone would silently truncate
	// every large write to its low word: the server would write 4464 of 70000
	// bytes and report success.
	declared := int(c.DataLengthHigh())<<16 | int(c.DataLength)
	if declared == 0 {
		c.Data = []types.UCHAR{}
		return 0, nil
	}

	// The position comes from DataOffset, measured from the start of the SMB
	// header "regardless of the command request's position in an AndX chain"
	// ([MS-CIFS] section 2.2.4.43.1). rawData begins at this command, so the
	// header and everything batched ahead of it come off the offset first.
	//
	// This is the only way to find data the sender relocated, and the only way to
	// find data at all once the block is too large for ByteCount to describe.
	commandStart := header.SMB_HEADER_SIZE + c.chainOffset
	start := int(c.DataOffset) - commandStart

	// A DataOffset that lands inside this command's own parameter block, or that
	// would run the data past the end of the message, is not a relocation but a
	// malformed request: [MS-CIFS] section 3.3.5.37 has the server fail such a
	// request with STATUS_INVALID_SMB, which is what an error here becomes.
	minimumDataStart := bytesRead + 2
	if start < minimumDataStart {
		return 0, fmt.Errorf(
			"WriteAndx DataOffset %d places the data %d bytes into a command whose data block starts at %d",
			c.DataOffset, start, minimumDataStart)
	}
	if start+declared > len(rawData) {
		return 0, fmt.Errorf(
			"WriteAndx declares %d bytes of data at offset %d, which runs past the %d bytes of the message that remain",
			declared, c.DataOffset, len(rawData))
	}

	// The pad, if the sender inserted one, is the byte before the data. Reading it
	// from the front of the data block instead would pick up a data byte for a
	// sender that padded with nothing.
	if start > minimumDataStart {
		c.Pad = types.UCHAR(rawData[start-1])
	}

	c.Data = rawData[start : start+declared]
	return declared, nil
}
