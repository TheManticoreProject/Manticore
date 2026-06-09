package header

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

const (
	// SMB2_HEADER_SIZE is the fixed wire size, in bytes, of an SMB2 packet header.
	SMB2_HEADER_SIZE = 64

	// SMB2_HEADER_STRUCTURE_SIZE is the value the StructureSize field MUST carry.
	SMB2_HEADER_STRUCTURE_SIZE = 64

	// SMB2_SIGNATURE_SIZE is the size, in bytes, of the header Signature field.
	SMB2_SIGNATURE_SIZE = 16
)

// SMB2ProtocolId is the 4-byte protocol identifier prefixing every SMB2 message:
// 0xFE 'S' 'M' 'B' (0x424D53FE in little-endian). It distinguishes SMB2 from the
// SMB 1.0 marker 0xFF 'S' 'M' 'B'.
var SMB2ProtocolId = [4]byte{0xFE, 'S', 'M', 'B'}

// Header represents the fixed 64-byte SMB2 packet header.
//
// The header has two forms, selected by the SMB2_FLAGS_ASYNC_COMMAND bit in
// Flags: the SYNC form carries Reserved(4) + TreeId(4) at offsets 32..39, while
// the ASYNC form carries a single AsyncId(8) in the same span. Both forms are
// represented by this single struct; Marshal/Unmarshal select the layout from
// the Flags field.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/5cd64522-60b3-4f3e-a157-fe66f1228052
type Header struct {
	// ProtocolId (4 bytes): MUST be 0xFE 'S' 'M' 'B'.
	ProtocolId [4]byte

	// StructureSize (2 bytes): MUST be set to 64.
	StructureSize types.USHORT

	// CreditCharge (2 bytes): In the SMB 2.0.2 dialect this field MUST be 0 and
	// ignored on receipt. In later dialects it indicates the number of credits
	// the request consumes.
	CreditCharge types.USHORT

	// Status (4 bytes): The (ChannelSequence,Reserved)/Status union. In the
	// SMB 2.0.2 and SMB 2.1 dialects this is the Status field in both requests
	// (client sets 0) and responses. In the SMB 3.x family it is reinterpreted
	// as ChannelSequence(2)+Reserved(2) in a request; that reinterpretation is
	// deferred to the SMB 3.x implementation.
	Status types.ULONG

	// Command (2 bytes): The command code of this packet.
	Command codes.CommandCode

	// Credit (2 bytes): CreditRequest on a request / CreditResponse on a response.
	Credit types.USHORT

	// Flags (4 bytes): Indicates how to process the operation.
	Flags flags.Flags

	// NextCommand (4 bytes): Offset, in bytes, from the start of this header to
	// the next 8-byte-aligned header in a compounded message, or 0 if this is
	// the last (or only) command.
	NextCommand types.ULONG

	// MessageId (8 bytes): Uniquely identifies a request/response pair across all
	// messages on the same transport connection.
	MessageId types.UINT64

	// Reserved (4 bytes): SYNC form only; SHOULD be 0. Overlaps AsyncId in the
	// ASYNC form.
	Reserved types.ULONG

	// TreeId (4 bytes): SYNC form only; identifies the tree connect for the
	// command. Overlaps AsyncId in the ASYNC form.
	TreeId types.ULONG

	// AsyncId (8 bytes): ASYNC form only; identifies an operation being processed
	// asynchronously. Overlaps Reserved+TreeId in the SYNC form.
	AsyncId types.UINT64

	// SessionId (8 bytes): Uniquely identifies the established session for the
	// command. MUST be 0 for an SMB2 NEGOTIATE.
	SessionId types.UINT64

	// Signature (16 bytes): The message signature if SMB2_FLAGS_SIGNED is set;
	// otherwise all zero.
	Signature [SMB2_SIGNATURE_SIZE]byte
}

// NewHeader creates a new SMB2 SYNC Header with default values: the SMB2
// protocol identifier and StructureSize set to 64.
//
// Returns:
//   - *Header: A pointer to the newly created SMB2 Header.
func NewHeader() *Header {
	return &Header{
		ProtocolId:    SMB2ProtocolId,
		StructureSize: SMB2_HEADER_STRUCTURE_SIZE,
	}
}

// Marshal serializes the SMB2 Header into a 64-byte slice using little-endian
// byte order. The Reserved+TreeId vs AsyncId span at offsets 32..39 is chosen
// from the SMB2_FLAGS_ASYNC_COMMAND bit in Flags.
//
// Returns:
//   - []byte: The serialized header (always 64 bytes on success).
//   - error: Any error encountered during serialization.
func (h *Header) Marshal() ([]byte, error) {
	buf := make([]byte, SMB2_HEADER_SIZE)

	// ProtocolId (4 bytes)
	copy(buf[0:4], h.ProtocolId[:])

	// StructureSize (2 bytes)
	binary.LittleEndian.PutUint16(buf[4:6], h.StructureSize)

	// CreditCharge (2 bytes)
	binary.LittleEndian.PutUint16(buf[6:8], h.CreditCharge)

	// Status / (ChannelSequence,Reserved) (4 bytes)
	binary.LittleEndian.PutUint32(buf[8:12], h.Status)

	// Command (2 bytes)
	binary.LittleEndian.PutUint16(buf[12:14], uint16(h.Command))

	// Credit (2 bytes)
	binary.LittleEndian.PutUint16(buf[14:16], h.Credit)

	// Flags (4 bytes)
	binary.LittleEndian.PutUint32(buf[16:20], uint32(h.Flags))

	// NextCommand (4 bytes)
	binary.LittleEndian.PutUint32(buf[20:24], h.NextCommand)

	// MessageId (8 bytes)
	binary.LittleEndian.PutUint64(buf[24:32], h.MessageId)

	// Reserved+TreeId (SYNC) or AsyncId (ASYNC) (8 bytes)
	if h.Flags.IsAsync() {
		binary.LittleEndian.PutUint64(buf[32:40], h.AsyncId)
	} else {
		binary.LittleEndian.PutUint32(buf[32:36], h.Reserved)
		binary.LittleEndian.PutUint32(buf[36:40], h.TreeId)
	}

	// SessionId (8 bytes)
	binary.LittleEndian.PutUint64(buf[40:48], h.SessionId)

	// Signature (16 bytes)
	copy(buf[48:64], h.Signature[:])

	return buf, nil
}

// Unmarshal deserializes a 64-byte SMB2 Header from the input byte slice using
// little-endian byte order. The offsets 32..39 are decoded as AsyncId when
// SMB2_FLAGS_ASYNC_COMMAND is set in the parsed Flags, and as Reserved+TreeId
// otherwise.
//
// Parameters:
//   - data: The byte slice containing the serialized SMB2 header.
//
// Returns:
//   - int: The number of bytes read (64 on success).
//   - error: An error if the input is shorter than 64 bytes.
func (h *Header) Unmarshal(data []byte) (int, error) {
	if len(data) < SMB2_HEADER_SIZE {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 header: have %d bytes, need %d", len(data), SMB2_HEADER_SIZE)
	}

	// ProtocolId (4 bytes)
	copy(h.ProtocolId[:], data[0:4])

	// StructureSize (2 bytes)
	h.StructureSize = binary.LittleEndian.Uint16(data[4:6])

	// CreditCharge (2 bytes)
	h.CreditCharge = binary.LittleEndian.Uint16(data[6:8])

	// Status / (ChannelSequence,Reserved) (4 bytes)
	h.Status = binary.LittleEndian.Uint32(data[8:12])

	// Command (2 bytes)
	h.Command = codes.CommandCode(binary.LittleEndian.Uint16(data[12:14]))

	// Credit (2 bytes)
	h.Credit = binary.LittleEndian.Uint16(data[14:16])

	// Flags (4 bytes)
	h.Flags = flags.Flags(binary.LittleEndian.Uint32(data[16:20]))

	// NextCommand (4 bytes)
	h.NextCommand = binary.LittleEndian.Uint32(data[20:24])

	// MessageId (8 bytes)
	h.MessageId = binary.LittleEndian.Uint64(data[24:32])

	// Reserved+TreeId (SYNC) or AsyncId (ASYNC) (8 bytes). The two interpretations
	// overlap on the wire, so populate both views to keep the struct consistent
	// regardless of which the caller reads.
	if h.Flags.IsAsync() {
		h.AsyncId = binary.LittleEndian.Uint64(data[32:40])
		h.Reserved = 0
		h.TreeId = 0
	} else {
		h.Reserved = binary.LittleEndian.Uint32(data[32:36])
		h.TreeId = binary.LittleEndian.Uint32(data[36:40])
		h.AsyncId = 0
	}

	// SessionId (8 bytes)
	h.SessionId = binary.LittleEndian.Uint64(data[40:48])

	// Signature (16 bytes)
	copy(h.Signature[:], data[48:64])

	return SMB2_HEADER_SIZE, nil
}
