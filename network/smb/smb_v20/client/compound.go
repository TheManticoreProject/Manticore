package client

import (
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// relatedFileId is the sentinel FileId (all 0xFF) that a request in a related
// compound chain uses to mean "the FileId produced by the preceding CREATE in
// this chain" (MS-SMB2 3.2.4.1.4).
var relatedFileId = types.SMB2_FILEID{
	Persistent: 0xFFFFFFFFFFFFFFFF,
	Volatile:   0xFFFFFFFFFFFFFFFF,
}

// compoundSegments returns the byte range of each segment in a marshalled
// compound buffer by walking the NextCommand chain in the segment headers. A
// non-compounded (single-message) buffer yields a one-element slice. Each
// returned segment slice aliases buf, so signing a segment in place updates buf.
func compoundSegments(buf []byte) ([][]byte, error) {
	var segments [][]byte
	offset := 0
	for offset < len(buf) {
		if len(buf)-offset < header.SMB2_HEADER_SIZE {
			return nil, fmt.Errorf("compound segment at offset %d is shorter than an SMB2 header", offset)
		}
		h := header.NewHeader()
		if _, err := h.Unmarshal(buf[offset:]); err != nil {
			return nil, fmt.Errorf("parsing compound segment header at offset %d: %w", offset, err)
		}
		next := int(h.NextCommand)
		if next == 0 {
			segments = append(segments, buf[offset:])
			break
		}
		if next < header.SMB2_HEADER_SIZE || offset+next > len(buf) {
			return nil, fmt.Errorf("compound NextCommand offset %d at segment offset %d out of bounds (buffer is %d bytes)", next, offset, len(buf))
		}
		segments = append(segments, buf[offset:offset+next])
		offset += next
	}
	return segments, nil
}

// signCompound signs each segment of a marshalled compound buffer in place. Each
// compounded PDU carries its own signature computed over its own region (header,
// body, and inter-segment padding for non-final segments), so they are signed
// individually rather than over the whole buffer.
func signCompound(key, buf []byte) error {
	segments, err := compoundSegments(buf)
	if err != nil {
		return err
	}
	for _, seg := range segments {
		signMessage(key, seg)
	}
	return nil
}

// sendReceiveCompound marshals a chain of requests into a single compounded SMB2
// request, signs each segment when signing is active, sends it as one frame, then
// parses the compounded response into one Message per segment. Each response
// segment is verified (protocol id, signature when required, and MessageId
// against the set of requests) and its command body is decoded only for statuses
// that carry one (mirroring sendReceive). Server-initiated frames (an interleaved
// OPLOCK_BREAK) are skipped.
//
// Interim STATUS_PENDING responses are not supported on the compound path: a
// compounded chain is expected to complete synchronously. A pending segment is
// reported as an error.
func (c *Client) sendReceiveCompound(msgs []*message.Message, label string) ([]*message.Message, error) {
	if !c.Transport.IsConnected() {
		return nil, fmt.Errorf("%s: transport is not connected", label)
	}
	if len(msgs) == 0 {
		return nil, fmt.Errorf("%s: no messages to send", label)
	}

	// Record the request MessageIds so each response segment can be matched back
	// to a request it answers.
	requestIds := make(map[uint64]bool, len(msgs))
	for _, m := range msgs {
		requestIds[uint64(m.Header.MessageId)] = true
	}

	marshalled, err := message.MarshalCompound(msgs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %s: %w", label, err)
	}

	if c.Session != nil && c.Session.SigningActive {
		if err := signCompound(c.Session.SigningKey, marshalled); err != nil {
			return nil, fmt.Errorf("failed to sign %s: %w", label, err)
		}
	}

	if _, err := c.Transport.Send(marshalled); err != nil {
		return nil, fmt.Errorf("failed to send %s: %w", label, err)
	}

	// Read the response frame, skipping any leading server-initiated message
	// (reserved MessageId), which is not a reply to this chain.
	var raw []byte
	for {
		raw, err = c.Transport.Receive()
		if err != nil {
			return nil, fmt.Errorf("failed to receive %s response: %w", label, err)
		}
		probe := header.NewHeader()
		if _, err := probe.Unmarshal(raw); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s response header: %w", label, err)
		}
		if uint64(probe.MessageId) != unsolicitedMessageId {
			break
		}
	}

	segments, err := compoundSegments(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}

	responses := make([]*message.Message, 0, len(segments))
	for i, seg := range segments {
		resp := message.NewMessage()
		if _, err := resp.Header.Unmarshal(seg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s response segment %d header: %w", label, i, err)
		}
		if !resp.Header.HasValidProtocolId() {
			return nil, fmt.Errorf("%s response segment %d is not an SMB2 message (ProtocolId % x)", label, i, resp.Header.ProtocolId)
		}

		// Enforce signing per segment, with the same exemptions as the
		// single-message path (MS-SMB2 3.2.5.1.3).
		if c.Session != nil && c.Session.SigningActive && signatureRequired(resp) {
			if !verifySignature(c.Session.SigningKey, seg) {
				return nil, fmt.Errorf("%s response segment %d failed SMB2 signature verification", label, i)
			}
		}
		if resp.Header.Credit > 0 {
			c.Connection.Credits = resp.Header.Credit
		}

		if uint64(resp.Header.MessageId) != unsolicitedMessageId && !requestIds[uint64(resp.Header.MessageId)] {
			return nil, fmt.Errorf("%s response segment %d MessageId %d does not match any request", label, i, uint64(resp.Header.MessageId))
		}
		if resp.Header.Status == ntStatusPending {
			return nil, fmt.Errorf("%s response segment %d returned STATUS_PENDING; the compound path does not support async completion", label, i)
		}

		// Decode the command body only for statuses that carry one (mirrors
		// sendReceive); an error segment carries the fixed SMB2 ERROR body.
		status := resp.Header.Status
		if status == 0x00000000 || status == ntStatusMoreProcessingRequired || status == ntStatusBufferOverflow {
			if _, err := resp.Unmarshal(seg); err != nil {
				return nil, fmt.Errorf("failed to unmarshal %s response segment %d: %w", label, i, err)
			}
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

// CreateQueryInfoClose opens path, queries the given information class on the
// resulting handle, and closes it — all in a single compounded request
// (SMB2_FLAGS_RELATED_OPERATIONS), so the three operations cost one network round
// trip instead of three. It returns the QUERY_INFO output buffer.
//
// infoType is one of commands.SMB2_0_INFO_* and fileInfoClass the class within
// it; additionalInformation carries the security-info bits for a security query
// and is 0 otherwise. Wire: SMB2 CREATE + QUERY_INFO + CLOSE compounded with
// related operations, where the QUERY_INFO and CLOSE inherit the CREATE's handle
// via the reserved related FileId.
func (c *Client) CreateQueryInfoClose(path string, desiredAccess, shareAccess, createDisposition, createOptions uint32, infoType, fileInfoClass uint8, additionalInformation uint32) ([]byte, error) {
	if c.Session == nil || c.Session.TreeId == 0 {
		return nil, fmt.Errorf("no tree connect established")
	}

	// CREATE — the first request in the chain, with the real share-relative name.
	create := commands.NewCreateRequest()
	create.RequestedOplockLevel = types.UCHAR(commands.SMB2_OPLOCK_LEVEL_NONE)
	create.DesiredAccess = desiredAccess
	create.ShareAccess = shareAccess
	create.CreateDisposition = createDisposition
	create.CreateOptions = createOptions
	create.ImpersonationLevel = 0x00000002 // Impersonation
	create.FileAttributes = 0x00000080     // FILE_ATTRIBUTE_NORMAL
	create.Name = strings.TrimPrefix(path, "\\")
	createMsg := c.newRequest(create)

	// QUERY_INFO — inherits the CREATE's handle via the related FileId sentinel.
	query := commands.NewQueryInfoRequest()
	query.InfoType = types.UCHAR(infoType)
	query.FileInfoClass = types.UCHAR(fileInfoClass)
	query.AdditionalInformation = types.ULONG(additionalInformation)
	query.OutputBufferLength = types.ULONG(c.Connection.Server.MaxTransactSize)
	query.FileId = relatedFileId
	queryMsg := c.newRequest(query)
	queryMsg.Header.AddFlags(flags.SMB2_FLAGS_RELATED_OPERATIONS)

	// CLOSE — also operates on the CREATE's handle.
	closeReq := commands.NewCloseRequest()
	closeReq.FileId = relatedFileId
	closeMsg := c.newRequest(closeReq)
	closeMsg.Header.AddFlags(flags.SMB2_FLAGS_RELATED_OPERATIONS)

	responses, err := c.sendReceiveCompound([]*message.Message{createMsg, queryMsg, closeMsg}, "CreateQueryInfoClose")
	if err != nil {
		return nil, err
	}
	if len(responses) != 3 {
		return nil, fmt.Errorf("expected 3 compounded responses, got %d", len(responses))
	}

	if status := statusFromResponse(responses[0]); status != 0x00000000 {
		return nil, fmt.Errorf("create %q failed: %s", path, formatNTStatus(status))
	}
	if status := statusFromResponse(responses[1]); status != 0x00000000 {
		return nil, fmt.Errorf("query info failed: %s", formatNTStatus(status))
	}
	if status := statusFromResponse(responses[2]); status != 0x00000000 {
		return nil, fmt.Errorf("close failed: %s", formatNTStatus(status))
	}

	queryResp, ok := responses[1].Command.(*commands.QueryInfoResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected query info response command: %T", responses[1].Command)
	}
	return queryResp.OutputBuffer, nil
}
