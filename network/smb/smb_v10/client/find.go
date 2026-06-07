package client

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// trans2MessageOverhead is a conservative reservation, in bytes, for the SMB header
// and Transaction2 framing (WordCount, parameter words, ByteCount, Name, and
// alignment padding) when bounding the payload a single message may carry.
const trans2MessageOverhead = 128

// trans2RequestChunk is one message's worth of a (possibly fragmented) Transaction2
// request payload: a run of parameter bytes and/or data bytes together with their
// displacements within the full parameter and data buffers.
type trans2RequestChunk struct {
	params                []byte
	parameterDisplacement int
	data                  []byte
	dataDisplacement      int
}

// planTransaction2Chunks splits the parameter and data buffers into per-message
// chunks, each carrying at most maxPayload payload bytes. Parameters are emitted
// before data, and a chunk carries data only once all parameters have been placed
// (in this or a prior chunk). At least one chunk is always returned, even for an
// empty payload.
func planTransaction2Chunks(params, data []byte, maxPayload int) []trans2RequestChunk {
	if maxPayload < 1 {
		maxPayload = 1
	}

	chunks := []trans2RequestChunk{}
	paramPos, dataPos := 0, 0
	for {
		budget := maxPayload
		chunk := trans2RequestChunk{parameterDisplacement: paramPos, dataDisplacement: dataPos}

		if paramPos < len(params) {
			n := len(params) - paramPos
			if n > budget {
				n = budget
			}
			chunk.params = params[paramPos : paramPos+n]
			paramPos += n
			budget -= n
		}

		// Data is only placed once every parameter byte has been emitted.
		if paramPos == len(params) && budget > 0 && dataPos < len(data) {
			n := len(data) - dataPos
			if n > budget {
				n = budget
			}
			chunk.data = data[dataPos : dataPos+n]
			dataPos += n
		}

		chunks = append(chunks, chunk)
		if paramPos >= len(params) && dataPos >= len(data) {
			break
		}
	}
	return chunks
}

// smbFindFileBothDirectoryInfo is the TRANS2_FIND information level
// SMB_FIND_FILE_BOTH_DIRECTORY_INFO (MS-CIFS 2.2.8.1.7).
const smbFindFileBothDirectoryInfo = 0x0104

// fileAttributeDirectory is the SMB_EXT_FILE_ATTR bit marking a directory.
const fileAttributeDirectory = 0x00000010

// bothDirInfoFixedSize is the size of the fixed portion of an
// SMB_FIND_FILE_BOTH_DIRECTORY_INFO entry, before the variable FileName:
// NextEntryOffset(4) FileIndex(4) 4xFILETIME(32) EndOfFile(8) AllocationSize(8)
// ExtFileAttributes(4) FileNameLength(4) EaSize(4) ShortNameLength(1) Reserved(1)
// ShortName(24).
const bothDirInfoFixedSize = 94

// Entry is a high-level directory entry (file or directory) returned by ListDirectory and ListEntries.
type Entry struct {
	LongName    string
	ShortName   string
	Size        uint64
	Attributes  uint32
	CreatedAt   time.Time
	AccessedAt  time.Time
	ModifiedAt  time.Time
	ChangedAt   time.Time
	IsDirectory bool
}

// ListDirectory lists all entries (files and directories) in the given directory
// on the current tree. path is a directory path using backslash separators
// relative to the share root (e.g. "\", "\subdir", "\subdir\nested").
// An empty path defaults to the share root.
//
// Wire: TRANS2_FIND_FIRST2 / FIND_NEXT2 with "\path\*" pattern.
func (c *Client) ListDirectory(path string) ([]Entry, error) {
	if path == "" {
		path = "\\"
	}
	if path[len(path)-1] == '\\' {
		path += "*"
	} else {
		path += "\\*"
	}
	return c.ListEntries(path)
}

// ListEntries enumerates entries matching pattern in the current tree, using
// TRANS2_FIND_FIRST2 followed by as many TRANS2_FIND_NEXT2 calls as needed, and
// always closing the search with FindClose2.
//
// pattern uses SMB wildcards and backslash separators relative to the share root
// (e.g. "\*.txt", "\docs\report-*.pdf"). An empty pattern defaults to "\*".
func (c *Client) ListEntries(pattern string) ([]Entry, error) {
	if c.Session == nil {
		return nil, fmt.Errorf("no session established")
	}

	if pattern == "" {
		pattern = "\\*"
	}
	if pattern[0] != '\\' {
		pattern = "\\" + pattern
	}

	// TRANS2_FIND_FIRST2.
	respParams, respData, err := c.trans2(uint16(subcommands.TRANS2_FIND_FIRST2), buildFindFirst2Params(pattern), nil)
	if err != nil {
		return nil, err
	}
	if len(respParams) < 6 {
		return nil, fmt.Errorf("FIND_FIRST2 response parameters too short (%d bytes)", len(respParams))
	}

	sid := binary.LittleEndian.Uint16(respParams[0:2])
	searchCount := binary.LittleEndian.Uint16(respParams[2:4])
	endOfSearch := binary.LittleEndian.Uint16(respParams[4:6])

	entries := parseBothDirInfo(respData)

	// TRANS2_FIND_NEXT2 until the server reports end-of-search or returns nothing.
	for endOfSearch == 0 && searchCount > 0 {
		respParams, respData, err = c.trans2(uint16(subcommands.TRANS2_FIND_NEXT2), buildFindNext2Params(sid), nil)
		if err != nil {
			break
		}
		if len(respParams) < 4 {
			break
		}
		searchCount = binary.LittleEndian.Uint16(respParams[0:2])
		endOfSearch = binary.LittleEndian.Uint16(respParams[2:4])

		next := parseBothDirInfo(respData)
		if len(next) == 0 {
			break
		}
		entries = append(entries, next...)
	}

	// Always release the search handle, even on a partial enumeration.
	_ = c.findClose2(sid)

	return entries, nil
}

// trans2 issues a single SMB_COM_TRANSACTION2 carrying the given subcommand, with
// the supplied transaction parameter and data buffers, and returns the reassembled
// response parameter and data buffers. It does not implement Transaction2Secondary
// fragmentation, so the request must fit in a single message.
func (c *Client) trans2(subcommand uint16, trans2Params, trans2Data []byte) ([]byte, []byte, error) {
	// Bound the data the server may return by the negotiated buffer.
	maxData := 0xFFFF
	if c.Connection.Server != nil && c.Connection.Server.MaxBufferSize > 0 {
		if budget := int(c.Connection.Server.MaxBufferSize) - 512; budget > 0 && budget < maxData {
			maxData = budget
		}
	}

	// Bound the request payload carried by each message by the negotiated buffer,
	// fragmenting across TRANSACTION2_SECONDARY messages when it does not fit.
	maxPayload := len(trans2Params) + len(trans2Data)
	if c.Connection.Server != nil && c.Connection.Server.MaxBufferSize > 0 {
		if budget := int(c.Connection.Server.MaxBufferSize) - trans2MessageOverhead; budget > 0 && budget < maxPayload {
			maxPayload = budget
		}
	}
	chunks := planTransaction2Chunks(trans2Params, trans2Data, maxPayload)

	// Build the primary SMB_COM_TRANSACTION2 message carrying the first chunk. The
	// TotalParameterCount/TotalDataCount fields always advertise the full payload.
	msg := c.newFileIOMessage(codes.SMB_COM_TRANSACTION2)

	cmd := commands.NewTransaction2Request()
	cmd.Setup = []types.USHORT{types.USHORT(subcommand)}
	cmd.MaxParameterCount = types.USHORT(1024)
	cmd.MaxSetupCount = types.UCHAR(0)
	cmd.Flags = types.USHORT(0)
	cmd.Timeout = types.ULONG(0)
	cmd.MaxDataCount = types.USHORT(maxData)

	// Compute offsets. The request always carries exactly one setup word, so
	// WordCount = 14 fixed words + 1 setup word = 15.
	const wordCount = 14 + 1
	// Bytes from the start of the SMB header to the end of the Name byte: the data
	// block on the wire is ByteCount(2) + Name(1) + Pad1 + Trans2_Parameters + ...
	preParams := header.SMB_HEADER_SIZE + 1 /*WordCount*/ + 2*wordCount + 2 /*ByteCount*/ + 1 /*Name*/
	pad1 := (4 - (preParams % 4)) % 4
	parameterOffset := preParams + pad1

	cmd.TotalParameterCount = types.USHORT(len(trans2Params))
	cmd.TotalDataCount = types.USHORT(len(trans2Data))
	cmd.ParameterCount = types.USHORT(len(chunks[0].params))
	cmd.ParameterOffset = types.USHORT(parameterOffset)
	cmd.Pad1 = make([]types.UCHAR, pad1)
	cmd.Trans2_Parameters = []types.UCHAR(chunks[0].params)

	if len(chunks[0].data) > 0 {
		afterParams := parameterOffset + len(chunks[0].params)
		pad2 := (4 - (afterParams % 4)) % 4
		cmd.DataCount = types.USHORT(len(chunks[0].data))
		cmd.DataOffset = types.USHORT(afterParams + pad2)
		cmd.Pad2 = make([]types.UCHAR, pad2)
		cmd.Trans2_Data = []types.UCHAR(chunks[0].data)
	}

	msg.AddCommand(cmd)

	// Fast path: the whole request fits in a single message. This is byte-for-byte
	// the prior behavior, including request signing and response verification.
	if len(chunks) == 1 {
		response, raw, err := c.sendReceive(msg, "Transaction2")
		if err != nil {
			return nil, nil, err
		}
		if response.Header.Status != 0x00000000 {
			return nil, nil, fmt.Errorf("Transaction2 (subcommand 0x%04x) failed: 0x%08x", subcommand, response.Header.Status)
		}
		return reassembleTrans2(raw, c.recvTrans2Continuation)
	}

	// Fragmented request. Per-message signing of a multi-message transaction is not
	// supported; refuse rather than emit a wrongly-signed message.
	if c.Connection.IsSigningActive {
		return nil, nil, fmt.Errorf("Transaction2 request too large to send while message signing is active")
	}

	marshalledPrimary, err := msg.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal Transaction2 primary: %v", err)
	}
	fileIODump("Transaction2 request (primary)", marshalledPrimary)
	if _, err = c.Transport.Send(marshalledPrimary); err != nil {
		return nil, nil, fmt.Errorf("failed to send Transaction2 primary: %v", err)
	}

	// Send the remaining chunks as TRANSACTION2_SECONDARY continuation messages.
	for i := 1; i < len(chunks); i++ {
		secMsg := c.buildTransaction2Secondary(len(trans2Params), len(trans2Data), chunks[i])
		marshalledSec, err := secMsg.Marshal()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal Transaction2 secondary %d: %v", i, err)
		}
		fileIODump("Transaction2 request (secondary)", marshalledSec)
		if _, err = c.Transport.Send(marshalledSec); err != nil {
			return nil, nil, fmt.Errorf("failed to send Transaction2 secondary %d: %v", i, err)
		}
	}

	// The response is read only after every request message has been sent.
	raw, err := c.Transport.Receive()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to receive Transaction2 response: %v", err)
	}
	fileIODump("Transaction2 response", raw)

	response := message.NewMessage()
	if err = response.Unmarshal(raw); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal Transaction2 response: %v", err)
	}
	if response.Header.Status != 0x00000000 {
		return nil, nil, fmt.Errorf("Transaction2 (subcommand 0x%04x) failed: 0x%08x", subcommand, response.Header.Status)
	}
	return reassembleTrans2(raw, c.recvTrans2Continuation)
}

// recvTrans2Continuation reads one continuation fragment of a Transaction2 response
// directly from the transport.
func (c *Client) recvTrans2Continuation() ([]byte, error) {
	next, err := c.Transport.Receive()
	if err != nil {
		return nil, err
	}
	fileIODump("Transaction2 response (continuation)", next)
	return next, nil
}

// buildTransaction2Secondary builds one SMB_COM_TRANSACTION2_SECONDARY continuation
// message carrying chunk, advertising the full transaction totals and the chunk's
// displacements. A SMB_COM_TRANSACTION2_SECONDARY has 9 parameter words and, unlike
// the primary, no Name field.
func (c *Client) buildTransaction2Secondary(totalParams, totalData int, chunk trans2RequestChunk) *message.Message {
	msg := c.newFileIOMessage(codes.SMB_COM_TRANSACTION2_SECONDARY)

	cmd := commands.NewTransaction2SecondaryRequest()
	cmd.TotalParameterCount = types.USHORT(totalParams)
	cmd.TotalDataCount = types.USHORT(totalData)
	cmd.FID = types.USHORT(0xFFFF) // no FID for FIND/info transactions

	const wordCount = 9
	preParams := header.SMB_HEADER_SIZE + 1 /*WordCount*/ + 2*wordCount + 2 /*ByteCount*/

	// Align the parameter run to a 4-byte boundary. ParameterOffset and Pad1 are set
	// even when this message carries no parameters, because the decoder derives the
	// Pad2 length from (DataOffset - ParameterOffset - ParameterCount).
	pad1 := (4 - (preParams % 4)) % 4
	parameterOffset := preParams + pad1
	cmd.ParameterOffset = types.USHORT(parameterOffset)
	cmd.Pad1 = make([]types.UCHAR, pad1)

	if len(chunk.params) > 0 {
		cmd.ParameterCount = types.USHORT(len(chunk.params))
		cmd.ParameterDisplacement = types.USHORT(chunk.parameterDisplacement)
		cmd.Trans2_Parameters = []types.UCHAR(chunk.params)
	}

	if len(chunk.data) > 0 {
		end := parameterOffset + len(chunk.params)
		pad2 := (4 - (end % 4)) % 4
		cmd.DataCount = types.USHORT(len(chunk.data))
		cmd.DataOffset = types.USHORT(end + pad2)
		cmd.DataDisplacement = types.USHORT(chunk.dataDisplacement)
		cmd.Pad2 = make([]types.UCHAR, pad2)
		cmd.Trans2_Data = []types.UCHAR(chunk.data)
	}

	msg.AddCommand(cmd)
	return msg
}

// trans2Fragment holds the decoded contents of a single SMB_COM_TRANSACTION2
// response message: the total transaction sizes, the parameter/data runs carried
// by this message, and the displacement of each run within the full transaction.
type trans2Fragment struct {
	totalParameterCount   int
	totalDataCount        int
	parameters            []byte
	parameterDisplacement int
	data                  []byte
	dataDisplacement      int
}

// parseTrans2Fragment decodes one SMB_COM_TRANSACTION2 response message. The
// ParameterOffset/DataOffset fields locate the payload runs and are measured from
// the start of the SMB header. A short/interim response (WordCount < 10) decodes
// to an empty fragment.
func parseTrans2Fragment(raw []byte) (*trans2Fragment, error) {
	hdr := header.SMB_HEADER_SIZE
	if len(raw) < hdr+1 {
		return nil, fmt.Errorf("Transaction2 response too short")
	}

	wordCount := int(raw[hdr])
	wordsStart := hdr + 1
	if len(raw) < wordsStart+2*wordCount {
		return nil, fmt.Errorf("Transaction2 response parameter words truncated")
	}
	// The standard response has WordCount 0x0A (10 words) before any setup words.
	if wordCount < 10 {
		return &trans2Fragment{}, nil
	}

	w := raw[wordsStart:]
	frag := &trans2Fragment{
		totalParameterCount:   int(binary.LittleEndian.Uint16(w[0:2])),
		totalDataCount:        int(binary.LittleEndian.Uint16(w[2:4])),
		parameterDisplacement: int(binary.LittleEndian.Uint16(w[10:12])),
		dataDisplacement:      int(binary.LittleEndian.Uint16(w[16:18])),
	}
	paramCount := int(binary.LittleEndian.Uint16(w[6:8]))
	paramOffset := int(binary.LittleEndian.Uint16(w[8:10]))
	dataCount := int(binary.LittleEndian.Uint16(w[12:14]))
	dataOffset := int(binary.LittleEndian.Uint16(w[14:16]))

	if paramCount > 0 {
		if paramOffset < 0 || paramOffset+paramCount > len(raw) {
			return nil, fmt.Errorf("Transaction2 parameter window [%d:%d] out of bounds (%d bytes)", paramOffset, paramOffset+paramCount, len(raw))
		}
		frag.parameters = raw[paramOffset : paramOffset+paramCount]
	}
	if dataCount > 0 {
		if dataOffset < 0 || dataOffset+dataCount > len(raw) {
			return nil, fmt.Errorf("Transaction2 data window [%d:%d] out of bounds (%d bytes)", dataOffset, dataOffset+dataCount, len(raw))
		}
		frag.data = raw[dataOffset : dataOffset+dataCount]
	}

	return frag, nil
}

// parseTrans2Response extracts the parameter and data buffers from a single,
// unfragmented SMB_COM_TRANSACTION2 response.
func parseTrans2Response(raw []byte) ([]byte, []byte, error) {
	frag, err := parseTrans2Fragment(raw)
	if err != nil {
		return nil, nil, err
	}
	return frag.parameters, frag.data, nil
}

// reassembleTrans2 reassembles an SMB_COM_TRANSACTION2 response that the server
// may split across multiple SMB messages. first is the already-received first
// response message; recvNext returns each subsequent raw message. Fragments are
// placed into the full parameter/data buffers by their displacements, and reading
// stops once TotalParameterCount/TotalDataCount bytes have been collected.
func reassembleTrans2(first []byte, recvNext func() ([]byte, error)) ([]byte, []byte, error) {
	frag, err := parseTrans2Fragment(first)
	if err != nil {
		return nil, nil, err
	}

	totalParams := frag.totalParameterCount
	totalData := frag.totalDataCount
	params := make([]byte, totalParams)
	data := make([]byte, totalData)

	gotParams, err := placeTrans2Fragment(params, frag.parameters, frag.parameterDisplacement, "parameter")
	if err != nil {
		return nil, nil, err
	}
	gotData, err := placeTrans2Fragment(data, frag.data, frag.dataDisplacement, "data")
	if err != nil {
		return nil, nil, err
	}

	for gotParams < totalParams || gotData < totalData {
		next, err := recvNext()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to receive Transaction2 continuation: %v", err)
		}
		nf, err := parseTrans2Fragment(next)
		if err != nil {
			return nil, nil, err
		}
		// A continuation that advances neither run means the server will not
		// complete the transaction; stop rather than loop forever.
		if len(nf.parameters) == 0 && len(nf.data) == 0 {
			break
		}
		np, err := placeTrans2Fragment(params, nf.parameters, nf.parameterDisplacement, "parameter")
		if err != nil {
			return nil, nil, err
		}
		nd, err := placeTrans2Fragment(data, nf.data, nf.dataDisplacement, "data")
		if err != nil {
			return nil, nil, err
		}
		gotParams += np
		gotData += nd
	}

	return params, data, nil
}

// placeTrans2Fragment copies a fragment run into dst at displacement, bounds-checking
// the destination window, and returns the number of bytes copied.
func placeTrans2Fragment(dst, run []byte, displacement int, label string) (int, error) {
	if len(run) == 0 {
		return 0, nil
	}
	if displacement < 0 || displacement+len(run) > len(dst) {
		return 0, fmt.Errorf("Transaction2 %s fragment [%d:%d] exceeds total length %d", label, displacement, displacement+len(run), len(dst))
	}
	copy(dst[displacement:], run)
	return len(run), nil
}

// parseBothDirInfo decodes a buffer of consecutive SMB_FIND_FILE_BOTH_DIRECTORY_INFO
// entries (as returned in the FIND_FIRST2/FIND_NEXT2 data) into Entry values.
// Each entry may represent a file or a directory.
// FileName is decoded as OEM/ASCII because the client issues non-Unicode requests.
func parseBothDirInfo(data []byte) []Entry {
	entries := []Entry{}

	for pos := 0; pos+bothDirInfoFixedSize <= len(data); {
		next := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
		createdAt := filetimeToTime(binary.LittleEndian.Uint64(data[pos+8 : pos+16]))
		accessedAt := filetimeToTime(binary.LittleEndian.Uint64(data[pos+16 : pos+24]))
		modifiedAt := filetimeToTime(binary.LittleEndian.Uint64(data[pos+24 : pos+32]))
		changedAt := filetimeToTime(binary.LittleEndian.Uint64(data[pos+32 : pos+40]))
		size := binary.LittleEndian.Uint64(data[pos+40 : pos+48])
		attrs := binary.LittleEndian.Uint32(data[pos+56 : pos+60])
		nameLen := int(binary.LittleEndian.Uint32(data[pos+60 : pos+64]))
		shortNameLen := int(data[pos+68])

		shortName := ""
		if shortNameLen > 0 && shortNameLen <= 24 {
			shortName = decodeUTF16LE(data[pos+70 : pos+70+shortNameLen])
		}

		longName := ""
		if nameLen > 0 && pos+bothDirInfoFixedSize+nameLen <= len(data) {
			longName = string(data[pos+bothDirInfoFixedSize : pos+bothDirInfoFixedSize+nameLen])
		}

		entries = append(entries, Entry{
			LongName:    longName,
			ShortName:   shortName,
			Size:        size,
			Attributes:  attrs,
			CreatedAt:   createdAt,
			AccessedAt:  accessedAt,
			ModifiedAt:  modifiedAt,
			ChangedAt:   changedAt,
			IsDirectory: attrs&fileAttributeDirectory != 0,
		})

		if next == 0 {
			break
		}
		pos += next
	}

	return entries
}

// findClose2 releases a search handle obtained from FIND_FIRST2.
//
// Wire: FindClose2Request / FindClose2Response.
func (c *Client) findClose2(sid uint16) error {
	msg := c.newFileIOMessage(codes.SMB_COM_FIND_CLOSE2)

	cmd := commands.NewFindClose2Request()
	cmd.SearchHandle = types.USHORT(sid)

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "FindClose2")
	if err != nil {
		return err
	}
	if response.Header.Status != 0x00000000 {
		return fmt.Errorf("FindClose2 failed: 0x%08x", response.Header.Status)
	}
	return nil
}

// buildFindFirst2Params builds the TRANS2_FIND_FIRST2 transaction parameters.
func buildFindFirst2Params(pattern string) []byte {
	b := []byte{}
	b = binary.LittleEndian.AppendUint16(b, 0x0016) // SearchAttributes: include hidden/system/directory
	b = binary.LittleEndian.AppendUint16(b, 512)    // SearchCount: max entries per response
	b = binary.LittleEndian.AppendUint16(b, 0x0000) // Flags: none (the search is closed explicitly)
	b = binary.LittleEndian.AppendUint16(b, smbFindFileBothDirectoryInfo)
	b = binary.LittleEndian.AppendUint32(b, 0) // SearchStorageType
	b = append(b, []byte(pattern)...)
	b = append(b, 0x00) // null-terminated OEM pattern
	return b
}

// buildFindNext2Params builds the TRANS2_FIND_NEXT2 transaction parameters that
// continue a search from the server's cursor.
func buildFindNext2Params(sid uint16) []byte {
	b := []byte{}
	b = binary.LittleEndian.AppendUint16(b, sid) // SID (search handle)
	b = binary.LittleEndian.AppendUint16(b, 512) // SearchCount
	b = binary.LittleEndian.AppendUint16(b, smbFindFileBothDirectoryInfo)
	b = binary.LittleEndian.AppendUint32(b, 0)      // ResumeKey (unused; we did not request resume keys)
	b = binary.LittleEndian.AppendUint16(b, 0x0008) // Flags: SMB_FIND_CONTINUE_FROM_LAST
	b = append(b, 0x00)                             // FileName: empty (resume from server cursor)
	return b
}

// filetimeToTime converts a Windows FILETIME (100ns ticks since 1601-01-01 UTC)
// to a time.Time. A zero FILETIME maps to the zero time.
func filetimeToTime(ft uint64) time.Time {
	const epochDiff = 116444736000000000 // 100ns intervals between 1601-01-01 and 1970-01-01
	if ft == 0 || ft < epochDiff {
		return time.Time{}
	}
	return time.Unix(0, int64(ft-epochDiff)*100).UTC()
}

// decodeUTF16LE decodes a little-endian UTF-16 byte buffer into a string, trimming
// any trailing NUL units.
func decodeUTF16LE(b []byte) string {
	u16 := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		u16 = append(u16, binary.LittleEndian.Uint16(b[i:i+2]))
	}
	return strings.TrimRight(string(utf16.Decode(u16)), "\x00")
}
