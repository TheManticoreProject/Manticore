package client

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// compoundSegment pairs a response command body with the MessageId and status its
// header should carry.
type compoundSegment struct {
	cmd       command_interface.CommandInterface
	messageId uint64
	status    uint32
}

// cannedCompound builds the wire bytes of a compounded SMB2 response from the
// given segments, stamping each with its MessageId and status and the
// SERVER_TO_REDIR flag.
func cannedCompound(t *testing.T, segs []compoundSegment) []byte {
	t.Helper()
	msgs := make([]*message.Message, 0, len(segs))
	for _, s := range segs {
		m := message.NewMessage()
		m.Header.AddFlags(flags.SMB2_FLAGS_SERVER_TO_REDIR)
		m.Header.MessageId = types.UINT64(s.messageId)
		m.Header.Status = s.status
		m.SetCommand(s.cmd)
		msgs = append(msgs, m)
	}
	wire, err := message.MarshalCompound(msgs)
	if err != nil {
		t.Fatalf("MarshalCompound: %v", err)
	}
	return wire
}

// createQueryCloseResponse builds the canonical CREATE+QUERY_INFO+CLOSE compound
// response with MessageIds 0,1,2 and the given query output buffer.
func createQueryCloseResponse(t *testing.T, queryOutput []byte) []byte {
	t.Helper()
	createResp := commands.NewCreateResponse()
	createResp.FileId = types.SMB2_FILEID{Persistent: 0xCAFE, Volatile: 0xBEEF}
	queryResp := commands.NewQueryInfoResponse()
	queryResp.OutputBuffer = queryOutput
	return cannedCompound(t, []compoundSegment{
		{createResp, 0, 0},
		{queryResp, 1, 0},
		{commands.NewCloseResponse(), 2, 0},
	})
}

func TestCreateQueryInfoCloseCompound(t *testing.T) {
	want := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	ft := &fakeTransport{responses: [][]byte{createQueryCloseResponse(t, want)}}
	c := withConnectedTree(ft)

	got, err := c.CreateQueryInfoClose("dir\\file.txt", 0x00100081, 0x07, 0x00000001, 0x00000000, commands.SMB2_0_INFO_FILE, 0x12, 0)
	if err != nil {
		t.Fatalf("CreateQueryInfoClose: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("query output = % x, want % x", got, want)
	}

	// One compound frame must have been sent (a single Send, not three).
	if len(ft.sent) != 1 {
		t.Fatalf("expected 1 compound request frame sent, got %d", len(ft.sent))
	}
	// The sent frame must parse into three segments, with the 2nd and 3rd marked
	// as related operations and carrying the related FileId sentinel.
	segs, err := compoundSegments(ft.sent[0])
	if err != nil {
		t.Fatalf("compoundSegments(sent): %v", err)
	}
	if len(segs) != 3 {
		t.Fatalf("sent compound has %d segments, want 3", len(segs))
	}
	reqs, err := message.UnmarshalCompound(ft.sent[0])
	if err != nil {
		t.Fatalf("UnmarshalCompound(sent): %v", err)
	}
	if reqs[0].Header.Flags.IsRelatedOperations() {
		t.Errorf("first request must not set RELATED_OPERATIONS")
	}
	for i := 1; i < 3; i++ {
		if !reqs[i].Header.Flags.IsRelatedOperations() {
			t.Errorf("request %d must set RELATED_OPERATIONS", i)
		}
	}
}

func TestSendReceiveCompoundEnforcesSigning(t *testing.T) {
	key := []byte("0123456789abcdef")

	signingClient := func(ft *fakeTransport) *Client {
		c := newTestClient(ft)
		c.Session = &Session{Client: c, SessionId: 0x99, TreeId: 0x5, SigningActive: true, SigningKey: key}
		c.Connection.SessionTable[0x99] = c.Session
		return c
	}

	t.Run("unsigned compound rejected", func(t *testing.T) {
		ft := &fakeTransport{responses: [][]byte{createQueryCloseResponse(t, []byte{0x01})}}
		c := signingClient(ft)
		if _, err := c.CreateQueryInfoClose("x", 0x00100081, 0x07, 0x00000001, 0, commands.SMB2_0_INFO_FILE, 0x12, 0); err == nil {
			t.Fatal("expected an unsigned compound response to be rejected when signing is active")
		}
	})

	t.Run("signed compound accepted", func(t *testing.T) {
		raw := createQueryCloseResponse(t, []byte{0x01})
		if err := signCompound(key, raw); err != nil {
			t.Fatalf("signCompound: %v", err)
		}
		ft := &fakeTransport{responses: [][]byte{raw}}
		c := signingClient(ft)
		if _, err := c.CreateQueryInfoClose("x", 0x00100081, 0x07, 0x00000001, 0, commands.SMB2_0_INFO_FILE, 0x12, 0); err != nil {
			t.Fatalf("expected a correctly signed compound response to be accepted, got: %v", err)
		}
	})
}

func TestCreateQueryInfoCloseSurfacesSegmentError(t *testing.T) {
	// CREATE succeeds but QUERY_INFO fails (STATUS_ACCESS_DENIED); the error must
	// be surfaced and the CLOSE segment (an error body) must not break parsing.
	createResp := commands.NewCreateResponse()
	createResp.FileId = types.SMB2_FILEID{Persistent: 1, Volatile: 2}
	raw := cannedCompound(t, []compoundSegment{
		{createResp, 0, 0},
		{commands.NewQueryInfoResponse(), 1, 0xC0000022}, // STATUS_ACCESS_DENIED
		{commands.NewCloseResponse(), 2, 0},
	})
	ft := &fakeTransport{responses: [][]byte{raw}}
	c := withConnectedTree(ft)

	if _, err := c.CreateQueryInfoClose("x", 0x00100081, 0x07, 0x00000001, 0, commands.SMB2_0_INFO_FILE, 0x12, 0); err == nil {
		t.Fatal("expected an error when the QUERY_INFO segment fails")
	}
}
