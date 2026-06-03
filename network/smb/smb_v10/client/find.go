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
	maxBuffer := 0xFFFF
	if c.Connection != nil && c.Connection.Server != nil && c.Connection.Server.MaxBufferSize > 0 {
		maxBuffer = int(c.Connection.Server.MaxBufferSize)
	}

	// Bound the data the server may return by the negotiated buffer.
	maxData := 0xFFFF
	if budget := maxBuffer - 512; budget > 0 && budget < maxData {
		maxData = budget
	}

	// Plan how the request parameter/data payload is split across the primary
	// SMB_COM_TRANSACTION2 message and any SMB_COM_TRANSACTION2_SECONDARY messages
	// when it does not fit in a single SMB buffer.
	plan := planTrans2Send(len(trans2Params), len(trans2Data), maxBuffer)

	// Send the primary request first, then any continuation messages. The server
	// only replies once the whole transaction has been received.
	for i, chunk := range plan {
		var msg *message.Message
		label := "Transaction2 request"
		if i == 0 {
			msg = c.buildTrans2Primary(subcommand, maxData, trans2Params, trans2Data, chunk)
		} else {
			msg = c.buildTrans2Secondary(trans2Params, trans2Data, chunk)
			label = "Transaction2_Secondary request"
		}
		marshalled, err := msg.Marshal()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal %s: %v", label, err)
		}
		fileIODump(label, marshalled)
		if _, err := c.Transport.Send(marshalled); err != nil {
			return nil, nil, fmt.Errorf("failed to send %s: %v", label, err)
		}
	}

	// Receive the (possibly fragmented) response.
	raw, err := c.Transport.Receive()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to receive Transaction2 response: %v", err)
	}
	fileIODump("Transaction2 response", raw)

	// Read the SMB header to check the status. The parameter/data payload is
	// parsed by reassembleTrans2 below, so a full command unmarshal is not needed.
	if len(raw) < header.SMB_HEADER_SIZE {
		return nil, nil, fmt.Errorf("Transaction2 response too short")
	}
	respHeader := header.Header{}
	if _, err := respHeader.Unmarshal(raw[:header.SMB_HEADER_SIZE]); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal Transaction2 response header: %v", err)
	}
	if respHeader.Status != 0x00000000 {
		return nil, nil, fmt.Errorf("Transaction2 (subcommand 0x%04x) failed: 0x%08x", subcommand, respHeader.Status)
	}

	// Reassemble the response, which the server may also split across several SMB
	// messages for a large parameter/data payload.
	return reassembleTrans2(raw, func() ([]byte, error) {
		next, err := c.Transport.Receive()
		if err != nil {
			return nil, err
		}
		fileIODump("Transaction2 response (continuation)", next)
		return next, nil
	})
}

// trans2SendChunk describes the slice of the request parameter and data payloads
// carried by one SMB message (the primary request or a continuation).
type trans2SendChunk struct {
	paramDisplacement int
	paramLen          int
	dataDisplacement  int
	dataLen           int
}

// Per-message overheads (bytes consumed before the parameter/data payload),
// including a small slack for the 4-byte alignment padding (Pad1 + Pad2 <= 6).
const (
	// header + WordCount(1) + 15 words + ByteCount(2) + Name(1) + pad slack.
	trans2PrimaryOverhead = header.SMB_HEADER_SIZE + 1 + 2*15 + 2 + 1 + 8
	// header + WordCount(1) + 9 words + ByteCount(2) + pad slack (no Name field).
	trans2SecondaryOverhead = header.SMB_HEADER_SIZE + 1 + 2*9 + 2 + 8
)

// planTrans2Send splits a request payload of totalParams parameter bytes and
// totalData data bytes into per-message chunks bounded by the negotiated buffer.
// Parameters are emitted before data (per [MS-CIFS] 2.2.4.46.1); a message only
// begins carrying data once all parameters have been placed. A payload that fits
// in the primary message yields a single chunk.
func planTrans2Send(totalParams, totalData, maxBuffer int) []trans2SendChunk {
	primaryBudget := maxBuffer - trans2PrimaryOverhead
	secondaryBudget := maxBuffer - trans2SecondaryOverhead
	if primaryBudget < 1 {
		primaryBudget = 1
	}
	if secondaryBudget < 1 {
		secondaryBudget = 1
	}

	chunks := []trans2SendChunk{}
	pOff, dOff := 0, 0
	for {
		budget := secondaryBudget
		if len(chunks) == 0 {
			budget = primaryBudget
		}

		pLen := totalParams - pOff
		if pLen > budget {
			pLen = budget
		}
		if pLen < 0 {
			pLen = 0
		}
		budget -= pLen

		dLen := 0
		if pOff+pLen >= totalParams { // all parameters placed; this message may carry data
			dLen = totalData - dOff
			if dLen > budget {
				dLen = budget
			}
			if dLen < 0 {
				dLen = 0
			}
		}

		chunks = append(chunks, trans2SendChunk{
			paramDisplacement: pOff,
			paramLen:          pLen,
			dataDisplacement:  dOff,
			dataLen:           dLen,
		})
		pOff += pLen
		dOff += dLen

		if pOff >= totalParams && dOff >= totalData {
			break
		}
		// Guard against a pathologically small buffer that cannot make progress.
		if pLen == 0 && dLen == 0 {
			break
		}
	}
	return chunks
}

// buildTrans2Primary builds the primary SMB_COM_TRANSACTION2 request carrying the
// given chunk of the parameter/data payload. TotalParameterCount/TotalDataCount
// advertise the full transaction size; ParameterCount/DataCount cover this message.
func (c *Client) buildTrans2Primary(subcommand uint16, maxData int, fullParams, fullData []byte, chunk trans2SendChunk) *message.Message {
	msg := c.newFileIOMessage(codes.SMB_COM_TRANSACTION2)

	cmd := commands.NewTransaction2Request()
	cmd.Setup = []types.USHORT{types.USHORT(subcommand)}
	cmd.MaxParameterCount = types.USHORT(1024)
	cmd.MaxSetupCount = types.UCHAR(0)
	cmd.Flags = types.USHORT(0)
	cmd.Timeout = types.ULONG(0)
	cmd.MaxDataCount = types.USHORT(maxData)

	// The request carries exactly one setup word, so WordCount = 14 + 1 = 15.
	const wordCount = 14 + 1
	preParams := header.SMB_HEADER_SIZE + 1 /*WordCount*/ + 2*wordCount + 2 /*ByteCount*/ + 1 /*Name*/
	pad1 := (4 - (preParams % 4)) % 4
	parameterOffset := preParams + pad1

	cmd.TotalParameterCount = types.USHORT(len(fullParams))
	cmd.TotalDataCount = types.USHORT(len(fullData))
	cmd.ParameterCount = types.USHORT(chunk.paramLen)
	cmd.ParameterOffset = types.USHORT(parameterOffset)
	cmd.Pad1 = make([]types.UCHAR, pad1)
	cmd.Trans2_Parameters = []types.UCHAR(fullParams[chunk.paramDisplacement : chunk.paramDisplacement+chunk.paramLen])

	if chunk.dataLen > 0 {
		afterParams := parameterOffset + chunk.paramLen
		pad2 := (4 - (afterParams % 4)) % 4
		cmd.DataCount = types.USHORT(chunk.dataLen)
		cmd.DataOffset = types.USHORT(afterParams + pad2)
		cmd.Pad2 = make([]types.UCHAR, pad2)
		cmd.Trans2_Data = []types.UCHAR(fullData[chunk.dataDisplacement : chunk.dataDisplacement+chunk.dataLen])
	}

	msg.AddCommand(cmd)
	return msg
}

// buildTrans2Secondary builds an SMB_COM_TRANSACTION2_SECONDARY continuation
// request carrying the given chunk of the parameter/data payload at its
// displacement within the full transaction.
func (c *Client) buildTrans2Secondary(fullParams, fullData []byte, chunk trans2SendChunk) *message.Message {
	msg := c.newFileIOMessage(codes.SMB_COM_TRANSACTION2_SECONDARY)

	cmd := commands.NewTransaction2SecondaryRequest()
	cmd.TotalParameterCount = types.USHORT(len(fullParams))
	cmd.TotalDataCount = types.USHORT(len(fullData))
	cmd.FID = types.USHORT(0xFFFF) // unused for these transactions

	// Secondary requests have no Setup or Name field: WordCount = 9.
	const wordCount = 9
	preParams := header.SMB_HEADER_SIZE + 1 /*WordCount*/ + 2*wordCount + 2 /*ByteCount*/
	pad1 := (4 - (preParams % 4)) % 4
	parameterOffset := preParams + pad1

	cmd.Pad1 = make([]types.UCHAR, pad1)
	if chunk.paramLen > 0 {
		cmd.ParameterCount = types.USHORT(chunk.paramLen)
		cmd.ParameterOffset = types.USHORT(parameterOffset)
		cmd.ParameterDisplacement = types.USHORT(chunk.paramDisplacement)
		cmd.Trans2_Parameters = []types.UCHAR(fullParams[chunk.paramDisplacement : chunk.paramDisplacement+chunk.paramLen])
	}

	if chunk.dataLen > 0 {
		afterParams := parameterOffset + chunk.paramLen
		pad2 := (4 - (afterParams % 4)) % 4
		cmd.DataCount = types.USHORT(chunk.dataLen)
		cmd.DataOffset = types.USHORT(afterParams + pad2)
		cmd.DataDisplacement = types.USHORT(chunk.dataDisplacement)
		cmd.Pad2 = make([]types.UCHAR, pad2)
		cmd.Trans2_Data = []types.UCHAR(fullData[chunk.dataDisplacement : chunk.dataDisplacement+chunk.dataLen])
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
