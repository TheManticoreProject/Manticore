package client

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb"
	smb2dialects "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	smb2msg "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	smb2commands "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	smb2flags "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
)

// readDirectTCPFrame reads one Direct TCP ([MS-SMB2] 2.1) framed message.
func readDirectTCPFrame(c net.Conn) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(c, hdr[:]); err != nil {
		return nil, err
	}
	n := int(hdr[1])<<16 | int(hdr[2])<<8 | int(hdr[3])
	body := make([]byte, n)
	_, err := io.ReadFull(c, body)
	return body, err
}

// writeDirectTCPFrame writes one Direct TCP framed message.
func writeDirectTCPFrame(c net.Conn, body []byte) error {
	hdr := []byte{0, byte(len(body) >> 16), byte(len(body) >> 8), byte(len(body))}
	_, err := c.Write(append(hdr, body...))
	return err
}

// smb2NegotiateResponse builds an SMB2 NEGOTIATE response selecting dialect.
func smb2NegotiateResponse(t *testing.T, dialect smb2dialects.Dialect, messageId uint64) []byte {
	t.Helper()
	resp := smb2commands.NewNegotiateResponse()
	resp.DialectRevision = dialect
	resp.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	resp.MaxTransactSize = 0x10000
	resp.MaxReadSize = 0x10000
	resp.MaxWriteSize = 0x10000

	m := smb2msg.NewMessage()
	m.Header.AddFlags(smb2flags.SMB2_FLAGS_SERVER_TO_REDIR)
	m.Header.MessageId = messageId
	m.Header.Credit = 1
	m.SetCommand(resp)
	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("building negotiate response: %v", err)
	}
	return wire
}

// TestDialHighestInSetWildcardFollowUpUsesMessageId1 scripts a server that
// answers the multi-protocol negotiate with the SMB2 wildcard (0x02FF) and, like
// Samba and Windows, drops the connection unless the follow-up SMB2 NEGOTIATE
// carries MessageId 1 ([MS-SMB2] 3.2.5.2).
func TestDialHighestInSetWildcardFollowUpUsesMessageId1(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	followUpMessageId := make(chan uint64, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		c.SetDeadline(time.Now().Add(5 * time.Second))

		if _, err := readDirectTCPFrame(c); err != nil { // SMB1 multi-protocol NEGOTIATE
			return
		}
		if err := writeDirectTCPFrame(c, smb2NegotiateResponse(t, smb2dialects.SMB2_DIALECT_WILDCARD, 0)); err != nil {
			return
		}
		req, err := readDirectTCPFrame(c) // follow-up SMB2 NEGOTIATE
		if err != nil || len(req) < 32 {
			return
		}
		messageId := binary.LittleEndian.Uint64(req[24:32])
		followUpMessageId <- messageId
		if messageId != 1 {
			return // a real server drops the connection here
		}
		writeDirectTCPFrame(c, smb2NegotiateResponse(t, smb2dialects.SMB2_DIALECT_2_1_0, 1))
	}()

	addr := ln.Addr().(*net.TCPAddr)
	c, err := Dial(addr.IP.String(), addr.Port, Options{
		Preferred:   []smb.SMBProtocolVersion{smb.SMB_VERSION_2_1, smb.SMB_VERSION_1_0},
		Policy:      PolicyHighestInSet,
		DialTimeout: 2 * time.Second,
	})

	select {
	case got := <-followUpMessageId:
		if got != 1 {
			t.Errorf("follow-up SMB2 NEGOTIATE MessageId = %d, want 1", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server never received a follow-up SMB2 NEGOTIATE")
	}
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Disconnect()
	if c.Dialect() != smb.SMB_VERSION_2_1 {
		t.Errorf("Dialect = %v, want SMB 2.1", c.Dialect())
	}
}
