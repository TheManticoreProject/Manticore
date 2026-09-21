package client

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
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
		if err := signCompound(dialects.SMB2_DIALECT_2_0_2, -1, key, raw); err != nil {
			t.Fatalf("signCompound: %v", err)
		}
		ft := &fakeTransport{responses: [][]byte{raw}}
		c := signingClient(ft)
		if _, err := c.CreateQueryInfoClose("x", 0x00100081, 0x07, 0x00000001, 0, commands.SMB2_0_INFO_FILE, 0x12, 0); err != nil {
			t.Fatalf("expected a correctly signed compound response to be accepted, got: %v", err)
		}
	})
}

func TestSendReceiveCompoundEncryption(t *testing.T) {
	keyHex := "629BCBC54422A0F572B97F45989B6073"
	keyBytes := mustHex(t, keyHex)

	encryptingClient := func(ft *fakeTransport) *Client {
		c := newTestClient(ft)
		c.Connection.Dialect = dialects.SMB2_DIALECT_3_1_1
		c.Connection.Cipher = commands.SMB2_ENCRYPTION_AES128_GCM
		c.Session = &Session{
			Client:        c,
			SessionId:     0x99,
			TreeId:        0x5,
			SigningActive:  true,
			SigningKey:     keyBytes,
			EncryptionKey: keyBytes,
			DecryptionKey: keyBytes,
			EncryptData:   true,
		}
		c.Connection.SessionTable[0x99] = c.Session
		return c
	}

	t.Run("encrypted compound round-trip", func(t *testing.T) {
		plainResponse := createQueryCloseResponse(t, []byte{0xDE, 0xAD})

		// Build a client just to encrypt the response (simulating the server).
		serverC := encryptingClient(&fakeTransport{})
		encrypted, err := serverC.encryptMessage(plainResponse)
		if err != nil {
			t.Fatalf("encryptMessage: %v", err)
		}

		ft := &fakeTransport{responses: [][]byte{encrypted}}
		c := encryptingClient(ft)

		got, err := c.CreateQueryInfoClose("x", 0x00100081, 0x07, 0x00000001, 0, commands.SMB2_0_INFO_FILE, 0x12, 0)
		if err != nil {
			t.Fatalf("CreateQueryInfoClose: %v", err)
		}
		if !bytes.Equal(got, []byte{0xDE, 0xAD}) {
			t.Errorf("query output = % x, want de ad", got)
		}

		// The sent frame must be encrypted (TRANSFORM_HEADER, not raw SMB2).
		if len(ft.sent) != 1 {
			t.Fatalf("expected 1 frame sent, got %d", len(ft.sent))
		}
		if !isTransformHeader(ft.sent[0]) {
			t.Errorf("sent frame is not encrypted (missing TRANSFORM_HEADER)")
		}
	})

	t.Run("encrypted compound skips signature verification", func(t *testing.T) {
		// The plaintext response is NOT signed — on an encrypted session the
		// AEAD tag supersedes per-segment signatures. The compound path must
		// accept it.
		plainResponse := createQueryCloseResponse(t, []byte{0x01})

		serverC := encryptingClient(&fakeTransport{})
		encrypted, err := serverC.encryptMessage(plainResponse)
		if err != nil {
			t.Fatalf("encryptMessage: %v", err)
		}

		ft := &fakeTransport{responses: [][]byte{encrypted}}
		c := encryptingClient(ft)

		if _, err := c.CreateQueryInfoClose("x", 0x00100081, 0x07, 0x00000001, 0, commands.SMB2_0_INFO_FILE, 0x12, 0); err != nil {
			t.Fatalf("expected encrypted (unsigned) compound to be accepted, got: %v", err)
		}
	})
}

func TestSendUnrelatedCompound(t *testing.T) {
	queryOut1 := []byte{0xAA, 0xBB}
	queryOut2 := []byte{0xCC, 0xDD}

	resp1 := commands.NewQueryInfoResponse()
	resp1.OutputBuffer = queryOut1
	resp2 := commands.NewQueryInfoResponse()
	resp2.OutputBuffer = queryOut2

	wire := cannedCompound(t, []compoundSegment{
		{resp1, 0, 0},
		{resp2, 1, 0},
	})
	ft := &fakeTransport{responses: [][]byte{wire}}
	c := withConnectedTree(ft)

	q1 := commands.NewQueryInfoRequest()
	q1.InfoType = types.UCHAR(commands.SMB2_0_INFO_FILE)
	q1.FileInfoClass = 0x12
	q1.OutputBufferLength = 0x10000
	q1.FileId = types.SMB2_FILEID{Persistent: 0x11, Volatile: 0x22}
	msg1 := c.NewRequest(q1)

	q2 := commands.NewQueryInfoRequest()
	q2.InfoType = types.UCHAR(commands.SMB2_0_INFO_FILE)
	q2.FileInfoClass = 0x12
	q2.OutputBufferLength = 0x10000
	q2.FileId = types.SMB2_FILEID{Persistent: 0x33, Volatile: 0x44}
	msg2 := c.NewRequest(q2)

	responses, err := c.SendUnrelatedCompound([]*message.Message{msg1, msg2})
	if err != nil {
		t.Fatalf("SendUnrelatedCompound: %v", err)
	}
	if len(responses) != 2 {
		t.Fatalf("got %d responses, want 2", len(responses))
	}

	got1, ok := responses[0].Command.(*commands.QueryInfoResponse)
	if !ok {
		t.Fatalf("response 0 command type = %T, want *QueryInfoResponse", responses[0].Command)
	}
	if !bytes.Equal(got1.OutputBuffer, queryOut1) {
		t.Errorf("response 0 output = % x, want % x", got1.OutputBuffer, queryOut1)
	}
	got2, ok := responses[1].Command.(*commands.QueryInfoResponse)
	if !ok {
		t.Fatalf("response 1 command type = %T, want *QueryInfoResponse", responses[1].Command)
	}
	if !bytes.Equal(got2.OutputBuffer, queryOut2) {
		t.Errorf("response 1 output = % x, want % x", got2.OutputBuffer, queryOut2)
	}

	// Verify the sent frame: one compound with two segments, neither RELATED.
	if len(ft.sent) != 1 {
		t.Fatalf("expected 1 compound frame sent, got %d", len(ft.sent))
	}
	reqs, err := message.UnmarshalCompound(ft.sent[0])
	if err != nil {
		t.Fatalf("UnmarshalCompound(sent): %v", err)
	}
	if len(reqs) != 2 {
		t.Fatalf("sent compound has %d segments, want 2", len(reqs))
	}
	for i, r := range reqs {
		if r.Header.Flags.IsRelatedOperations() {
			t.Errorf("request %d must NOT set RELATED_OPERATIONS", i)
		}
	}
	if reqs[0].Header.MessageId == reqs[1].Header.MessageId {
		t.Errorf("both requests have the same MessageId %d; each must be unique", reqs[0].Header.MessageId)
	}
}

func TestSendUnrelatedCompoundRejectsRelatedFlag(t *testing.T) {
	ft := &fakeTransport{}
	c := withConnectedTree(ft)

	q := commands.NewQueryInfoRequest()
	q.InfoType = types.UCHAR(commands.SMB2_0_INFO_FILE)
	q.FileInfoClass = 0x12
	q.OutputBufferLength = 0x10000
	q.FileId = types.SMB2_FILEID{Persistent: 0x11, Volatile: 0x22}
	msg1 := c.NewRequest(q)
	msg2 := c.NewRequest(q)
	msg2.Header.AddFlags(flags.SMB2_FLAGS_RELATED_OPERATIONS)

	if _, err := c.SendUnrelatedCompound([]*message.Message{msg1, msg2}); err == nil {
		t.Fatal("expected error when a request carries RELATED_OPERATIONS")
	}
}

func TestSendUnrelatedCompoundRequiresTwo(t *testing.T) {
	ft := &fakeTransport{}
	c := withConnectedTree(ft)

	q := commands.NewQueryInfoRequest()
	q.FileId = types.SMB2_FILEID{Persistent: 1, Volatile: 2}
	msg := c.NewRequest(q)

	if _, err := c.SendUnrelatedCompound([]*message.Message{msg}); err == nil {
		t.Fatal("expected error for a single-request unrelated compound")
	}
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

// TestSendReceiveCompoundEncrypts verifies that when EncryptData is set the
// compound request is wrapped in a TRANSFORM_HEADER and the encrypted compound
// response is correctly decrypted.
func TestSendReceiveCompoundEncrypts(t *testing.T) {
	key := mustHex(t, "629BCBC54422A0F572B97F45989B6073")

	encryptingClient := func() (*Client, *fakeTransport) {
		ft := &fakeTransport{}
		c := newTestClient(ft)
		c.Connection.Dialect = dialects.SMB2_DIALECT_3_1_1
		c.Connection.Cipher = commands.SMB2_ENCRYPTION_AES128_GCM
		c.Session = &Session{
			Client:        c,
			SessionId:     0x99,
			TreeId:        0x5,
			EncryptionKey: key,
			DecryptionKey: key,
			EncryptData:   true,
		}
		c.Connection.SessionTable[0x99] = c.Session
		return c, ft
	}

	t.Run("request is encrypted", func(t *testing.T) {
		c, ft := encryptingClient()
		raw := createQueryCloseResponse(t, []byte{0x01})
		enc, err := c.encryptMessage(raw)
		if err != nil {
			t.Fatalf("encryptMessage: %v", err)
		}
		ft.responses = [][]byte{enc}
		if _, err := c.CreateQueryInfoClose("x", 0x00100081, 0x07, 0x00000001, 0, commands.SMB2_0_INFO_FILE, 0x12, 0); err != nil {
			t.Fatalf("CreateQueryInfoClose: %v", err)
		}
		if len(ft.sent) != 1 {
			t.Fatalf("expected 1 sent frame, got %d", len(ft.sent))
		}
		if !isTransformHeader(ft.sent[0]) {
			t.Fatal("compound request must be encrypted (TRANSFORM_HEADER) when EncryptData is set")
		}
	})

	t.Run("encrypted response decrypted", func(t *testing.T) {
		c, ft := encryptingClient()
		want := []byte{0xDE, 0xAD, 0xBE, 0xEF}
		raw := createQueryCloseResponse(t, want)
		enc, err := c.encryptMessage(raw)
		if err != nil {
			t.Fatalf("encryptMessage: %v", err)
		}
		ft.responses = [][]byte{enc}
		got, err := c.CreateQueryInfoClose("x", 0x00100081, 0x07, 0x00000001, 0, commands.SMB2_0_INFO_FILE, 0x12, 0)
		if err != nil {
			t.Fatalf("CreateQueryInfoClose: %v", err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("query output = % x, want % x", got, want)
		}
	})
}
