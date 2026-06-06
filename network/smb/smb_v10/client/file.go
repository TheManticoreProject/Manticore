package client

import (
	"encoding/hex"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

const (
	// writeAndxWordsSize is the size in bytes of the WriteAndx request Words block
	// for the 32-bit-offset form (WordCount 0x0C, no OffsetHigh): AndX(4) + FID(2) +
	// Offset(4) + Timeout(4) + WriteMode(2) + Remaining(2) + Reserved(2) + DataLength(2)
	// + DataOffset(2).
	writeAndxWordsSize = 24

	// writeAndxDataOffset is the byte offset, measured from the start of the SMB
	// header, at which the Data block begins in our single (non-chained) WriteAndx
	// request when the 32-bit-offset form (WordCount 0x0C, no OffsetHigh) is used:
	// SMB header + WordCount(1) + Words + ByteCount(2) + Pad(1).
	writeAndxDataOffset = header.SMB_HEADER_SIZE + 1 + writeAndxWordsSize + 2 + 1

	// writeAndxDataOffset64 is the equivalent Data block offset when the 64-bit-offset
	// form (WordCount 0x0E) is used. WriteAndxRequest.Marshal appends the 4-byte
	// OffsetHigh word to the parameter block when OffsetHigh is non-zero, which shifts
	// the Data block 4 bytes further from the start of the header.
	writeAndxDataOffset64 = writeAndxDataOffset + 4
)

// FID is an opaque file handle returned by the server for an open file or directory.
type FID uint16

// NT access mask values (subset) used by OpenFile.
const (
	GENERIC_READ  uint32 = 0x80000000
	GENERIC_WRITE uint32 = 0x40000000
)

// FILE_SHARE_* values for the shareAccess argument of OpenFile.
const (
	FILE_SHARE_NONE   uint32 = 0x00000000
	FILE_SHARE_READ   uint32 = 0x00000001
	FILE_SHARE_WRITE  uint32 = 0x00000002
	FILE_SHARE_DELETE uint32 = 0x00000004
)

// CreateDisposition values for the createDisp argument of OpenFile.
const (
	FILE_SUPERSEDE    uint32 = 0x00000000
	FILE_OPEN         uint32 = 0x00000001
	FILE_CREATE       uint32 = 0x00000002
	FILE_OPEN_IF      uint32 = 0x00000003
	FILE_OVERWRITE    uint32 = 0x00000004
	FILE_OVERWRITE_IF uint32 = 0x00000005
)

// CreateOptions values for the createOptions argument of OpenFile.
const (
	FILE_DIRECTORY_FILE     uint32 = 0x00000001
	FILE_NON_DIRECTORY_FILE uint32 = 0x00000040
)

// FileIODebug, when true, dumps the raw bytes of file-I/O requests and responses
// to stdout. It is a coarse diagnostic aid until a structured logger is wired in.
var FileIODebug bool

func fileIODump(label string, data []byte) {
	if FileIODebug {
		fmt.Printf("[debug] %s (%d bytes)\n%s\n", label, len(data), hex.Dump(data))
	}
}

// newFileIOMessage builds a message pre-populated with the header fields common to
// every file-I/O command issued on the currently selected tree/session.
func (c *Client) newFileIOMessage(command codes.CommandCode) *message.Message {
	msg := message.NewMessage()
	msg.Header.Command = command
	msg.Header.Flags = flags.Flags(flags.FLAGS_CANONICALIZED_PATHS | flags.FLAGS_CASE_INSENSITIVE)
	msg.Header.Flags2 = flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES | flags2.FLAGS2_LONG_NAMES_ALLOWED)
	msg.Header.SetPID(msg.Header.GetPID())
	msg.Header.MID = c.Connection.MaxMpxCount
	msg.Header.TID = c.Session.TreeID
	msg.Header.UID = c.Session.SessionUID
	return msg
}

// sendReceive marshals msg, sends it on the transport, and returns both the parsed
// response message and the raw response bytes (the raw bytes are needed by callers
// that index into the message by an absolute offset, e.g. ReadAndx DataOffset).
func (c *Client) sendReceive(msg *message.Message, label string) (*message.Message, []byte, error) {
	marshalled, err := msg.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal %s: %v", label, err)
	}

	// Sign the request in place when message signing is active for this connection.
	responseSeq, _ := c.signOutgoing(marshalled)

	fileIODump(label+" request", marshalled)

	if _, err = c.Transport.Send(marshalled); err != nil {
		return nil, nil, fmt.Errorf("failed to send %s: %v", label, err)
	}

	raw, err := c.Transport.Receive()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to receive %s response: %v", label, err)
	}
	fileIODump(label+" response", raw)

	if err = c.verifyIncoming(raw, responseSeq); err != nil {
		return nil, nil, fmt.Errorf("%s: %v", label, err)
	}

	response := message.NewMessage()
	if err = response.Unmarshal(raw); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal %s response: %v", label, err)
	}

	return response, raw, nil
}

// OpenFile opens (or creates) a file on the currently connected tree and returns
// the server-assigned FID.
//
// Wire: NtCreateAndxRequest / NtCreateAndxResponse.
func (c *Client) OpenFile(path string, desiredAccess, shareAccess, createDisp, createOptions uint32) (FID, error) {
	if c.Session == nil {
		return 0, fmt.Errorf("no session established")
	}

	// NtCreate FileName is relative to the share root (the TID) and uses backslash separators.
	smbPath := path
	if len(smbPath) == 0 || smbPath[0] != '\\' {
		smbPath = "\\" + smbPath
	}

	msg := c.newFileIOMessage(codes.SMB_COM_NT_CREATE_ANDX)

	cmd := commands.NewNtCreateAndxRequest()
	cmd.DesiredAccess = types.ULONG(desiredAccess)
	cmd.ShareAccess = types.ULONG(shareAccess)
	cmd.CreateDisposition = types.ULONG(createDisp)
	cmd.CreateOptions = types.ULONG(createOptions)
	// SEC_IMPERSONATE (2): give the server a static snapshot of the client's security context.
	cmd.ImpersonationLevel = types.ULONG(0x00000002)
	// FILE_ATTRIBUTE_NORMAL
	cmd.ExtFileAttributes = types.SMB_EXT_FILE_ATTR(0x00000080)
	if err := cmd.FileName.SetString(smbPath); err != nil {
		return 0, fmt.Errorf("failed to set file name: %v", err)
	}
	cmd.NameLength = types.USHORT(len(smbPath))

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "NtCreateAndx")
	if err != nil {
		return 0, err
	}

	if response.Header.Status != 0x00000000 {
		return 0, fmt.Errorf("NtCreateAndx failed: 0x%08x", response.Header.Status)
	}

	createResponse, ok := response.Command.(*commands.NtCreateAndxResponse)
	if !ok {
		return 0, fmt.Errorf("unexpected response command: 0x%02x", response.Header.Command)
	}

	return FID(createResponse.FID), nil
}

// CloseFile releases an open FID on the server.
//
// Wire: CloseRequest / CloseResponse.
func (c *Client) CloseFile(fid FID) error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	msg := c.newFileIOMessage(codes.SMB_COM_CLOSE)

	cmd := commands.NewCloseRequest()
	cmd.FID = types.USHORT(fid)
	// LastTimeModified left at zero: the server uses its own default and does not
	// apply an explicit modification time on close.

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "Close")
	if err != nil {
		return err
	}

	if response.Header.Status != 0x00000000 {
		return fmt.Errorf("Close failed: 0x%08x", response.Header.Status)
	}

	return nil
}

// ReadFile reads up to maxLen bytes starting at offset from the file referenced by
// fid. It issues as many ReadAndx requests as required, stopping at maxLen or at the
// first short read (end of file).
//
// Wire: ReadAndxRequest / ReadAndxResponse.
func (c *Client) ReadFile(fid FID, offset uint64, maxLen uint32) ([]byte, error) {
	if c.Session == nil {
		return nil, fmt.Errorf("no session established")
	}

	// Bound each read by what the server will accept in a single response. Leave room
	// for the SMB header and the ReadAndx parameter/data framing.
	chunkSize := uint32(0xFF00)
	if c.Connection.Server != nil && c.Connection.Server.MaxBufferSize > 0 {
		if budget := c.Connection.Server.MaxBufferSize; budget < chunkSize+512 {
			if budget <= 512 {
				return nil, fmt.Errorf("negotiated MaxBufferSize (%d) too small to read", budget)
			}
			chunkSize = budget - 512
		}
	}

	result := make([]byte, 0, maxLen)

	for uint32(len(result)) < maxLen {
		want := maxLen - uint32(len(result))
		if want > chunkSize {
			want = chunkSize
		}

		msg := c.newFileIOMessage(codes.SMB_COM_READ_ANDX)

		absOffset := offset + uint64(len(result))

		cmd := commands.NewReadAndxRequest()
		cmd.FID = types.USHORT(fid)
		cmd.Offset = types.ULONG(uint32(absOffset))
		// Carry the upper 32 bits of the file offset so reads past 4 GiB target the
		// correct location. ReadAndxRequest emits the 0x0C (64-bit) form when OffsetHigh
		// is non-zero; for offsets below 4 GiB this stays zero and the 0x0A form is used.
		cmd.OffsetHigh = types.ULONG(uint32(absOffset >> 32))
		cmd.MaxCountOfBytesToReturn = types.USHORT(want)
		cmd.MinCountOfBytesToReturn = types.USHORT(0)
		cmd.Timeout = types.ULONG(0)
		cmd.Remaining = types.USHORT(0)

		msg.AddCommand(cmd)

		response, raw, err := c.sendReceive(msg, "ReadAndx")
		if err != nil {
			return result, err
		}

		if response.Header.Status != 0x00000000 {
			return result, fmt.Errorf("ReadAndx failed: 0x%08x", response.Header.Status)
		}

		readResponse, ok := response.Command.(*commands.ReadAndxResponse)
		if !ok {
			return result, fmt.Errorf("unexpected response command: 0x%02x", response.Header.Command)
		}

		dataLen := int(readResponse.DataLength)
		dataOff := int(readResponse.DataOffset)
		if dataLen == 0 {
			// End of file reached.
			break
		}
		if dataOff < 0 || dataOff+dataLen > len(raw) {
			return result, fmt.Errorf("ReadAndx data window [%d:%d] out of bounds (response is %d bytes)", dataOff, dataOff+dataLen, len(raw))
		}

		result = append(result, raw[dataOff:dataOff+dataLen]...)

		// A short read means we have reached the end of the file.
		if uint32(dataLen) < want {
			break
		}
	}

	return result, nil
}

// WriteFile writes data to the file referenced by fid starting at offset and
// returns the total number of bytes written. It issues as many WriteAndx requests
// as required, bounding each one so the request message stays within the negotiated
// MaxBufferSize.
//
// Wire: WriteAndxRequest / WriteAndxResponse.
func (c *Client) WriteFile(fid FID, offset uint64, data []byte) (int, error) {
	if c.Session == nil {
		return 0, fmt.Errorf("no session established")
	}

	// Each request carries: SMB header + WriteAndx parameters + Pad + Data. Keep the
	// whole message within the negotiated buffer by reserving the fixed overhead.
	chunkSize := 0xFF00
	if c.Connection.Server != nil && c.Connection.Server.MaxBufferSize > 0 {
		budget := int(c.Connection.Server.MaxBufferSize) - writeAndxDataOffset
		if budget <= 0 {
			return 0, fmt.Errorf("negotiated MaxBufferSize (%d) too small to write", c.Connection.Server.MaxBufferSize)
		}
		if budget < chunkSize {
			chunkSize = budget
		}
	}

	written := 0
	for written < len(data) {
		end := written + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[written:end]

		msg := c.newFileIOMessage(codes.SMB_COM_WRITE_ANDX)

		absOffset := offset + uint64(written)

		cmd := commands.NewWriteAndxRequest()
		cmd.FID = types.USHORT(fid)
		cmd.Offset = types.ULONG(uint32(absOffset))
		// Carry the upper 32 bits of the file offset so writes past 4 GiB land at the
		// correct location instead of wrapping at 2^32.
		cmd.OffsetHigh = types.ULONG(uint32(absOffset >> 32))
		cmd.Timeout = types.ULONG(0)
		cmd.WriteMode = types.USHORT(0)
		cmd.Remaining = types.USHORT(0)
		cmd.Reserved = types.USHORT(0)
		cmd.DataLength = types.USHORT(len(chunk))
		// DataOffset is measured from the start of the SMB header (not from the AndX
		// command's position). When OffsetHigh is non-zero, WriteAndxRequest.Marshal
		// appends the 4-byte OffsetHigh word to the parameter block, which moves the
		// Data block 4 bytes further out; use the matching offset so the server reads
		// the data from where it is actually placed.
		if cmd.OffsetHigh != 0 {
			cmd.DataOffset = types.USHORT(writeAndxDataOffset64)
		} else {
			cmd.DataOffset = types.USHORT(writeAndxDataOffset)
		}
		cmd.Pad = types.UCHAR(0)
		cmd.Data = []types.UCHAR(chunk)

		msg.AddCommand(cmd)

		response, _, err := c.sendReceive(msg, "WriteAndx")
		if err != nil {
			return written, err
		}

		if response.Header.Status != 0x00000000 {
			return written, fmt.Errorf("WriteAndx failed: 0x%08x", response.Header.Status)
		}

		writeResponse, ok := response.Command.(*commands.WriteAndxResponse)
		if !ok {
			return written, fmt.Errorf("unexpected response command: 0x%02x", response.Header.Command)
		}

		n := int(writeResponse.Count)
		if n <= 0 {
			return written, fmt.Errorf("server accepted 0 bytes of a %d-byte write", len(chunk))
		}

		written += n
	}

	return written, nil
}
