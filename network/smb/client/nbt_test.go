package client

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/netbios/nbt"
	"github.com/TheManticoreProject/Manticore/network/smb"
	smb2dialects "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	smb2msg "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	smb2commands "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	smb2flags "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
)

func TestOptionsNewTransport(t *testing.T) {
	if _, ok := (Options{}).newTransport().(*nbt.NBTTransport); ok {
		t.Error("the zero Options select NetBIOS, want Direct TCP")
	}
	if _, ok := (Options{Transport: TransportNetBIOS}).newTransport().(*nbt.NBTTransport); !ok {
		t.Error("TransportNetBIOS does not select the NetBIOS transport")
	}
}

func TestWithoutVersion(t *testing.T) {
	got := withoutVersion(defaultPreference(), smb.SMB_VERSION_3_1_1)
	if len(got) != len(defaultPreference())-1 {
		t.Fatalf("withoutVersion = %v", got)
	}
	for _, v := range got {
		if v == smb.SMB_VERSION_3_1_1 {
			t.Errorf("withoutVersion kept SMB 3.1.1: %v", got)
		}
	}
}

func TestDialNetBIOSRejectsSMB311Only(t *testing.T) {
	_, err := Dial("127.0.0.1", 139, Options{
		Transport: TransportNetBIOS,
		Preferred: []smb.SMBProtocolVersion{smb.SMB_VERSION_3_1_1},
	})
	if err == nil {
		t.Fatal("Dial offered SMB 3.1.1 over NetBIOS")
	}
}

// nbtReadFrame reads one RFC 1002 session-service packet: its type and body.
func nbtReadFrame(c net.Conn) (byte, []byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(c, hdr[:]); err != nil {
		return 0, nil, err
	}
	n := int(hdr[1]&0x01)<<16 | int(hdr[2])<<8 | int(hdr[3])
	body := make([]byte, n)
	_, err := io.ReadFull(c, body)
	return hdr[0], body, err
}

// TestDialNetBIOSSessionThenNegotiate scripts a NetBIOS session service: the
// client must send a SESSION REQUEST before any SMB, carry the Workstation as
// its calling name, and leave SMB 3.1.1 out of the SMB2 NEGOTIATE that follows.
func TestDialNetBIOSSessionThenNegotiate(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	type observed struct {
		firstType byte
		calling   []byte
		dialects  []smb2dialects.Dialect
	}
	seen := make(chan observed, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		c.SetDeadline(time.Now().Add(5 * time.Second))

		var o observed
		typ, body, err := nbtReadFrame(c)
		if err != nil {
			return
		}
		o.firstType = typ
		if typ == 0x81 && len(body) == 68 {
			o.calling = body[34:]
		}
		c.Write([]byte{0x82, 0, 0, 0}) // POSITIVE SESSION RESPONSE

		_, body, err = nbtReadFrame(c)
		if err != nil {
			seen <- o
			return
		}
		req := smb2msg.NewMessage()
		if _, err := req.Unmarshal(body); err == nil {
			if neg, ok := req.Command.(*smb2commands.NegotiateRequest); ok {
				o.dialects = neg.Dialects
			}
		}
		seen <- o

		resp := smb2commands.NewNegotiateResponse()
		resp.DialectRevision = smb2dialects.SMB2_DIALECT_3_0_2
		resp.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
		resp.MaxTransactSize, resp.MaxReadSize, resp.MaxWriteSize = 0x10000, 0x10000, 0x10000
		m := smb2msg.NewMessage()
		m.Header.AddFlags(smb2flags.SMB2_FLAGS_SERVER_TO_REDIR)
		m.Header.Credit = 1
		m.SetCommand(resp)
		wire, _ := m.Marshal()
		frame := make([]byte, 4, 4+len(wire))
		binary.BigEndian.PutUint32(frame, uint32(len(wire)))
		c.Write(append(frame, wire...))
	}()

	addr := ln.Addr().(*net.TCPAddr)
	c, err := Dial(addr.IP.String(), addr.Port, Options{
		Transport:   TransportNetBIOS,
		Workstation: "NBTCLIENT",
		DialTimeout: 2 * time.Second,
	})

	var o observed
	select {
	case o = <-seen:
	case <-time.After(5 * time.Second):
		t.Fatal("server observed nothing")
	}
	if o.firstType != 0x81 {
		t.Errorf("first packet type = %#02x, want 0x81 SESSION REQUEST", o.firstType)
	}
	// The calling name is second-level encoded ([RFC 1002] 4.1): decoding its
	// first label yields "NBTCLIENT" padded to 15 characters plus suffix 0x00.
	if len(o.calling) < 34 || decodeNetBIOSName(o.calling[1:33]) != "NBTCLIENT      \x00" {
		t.Errorf("calling name = %q, want NBTCLIENT", decodeNetBIOSName(o.calling[1:min(33, len(o.calling))]))
	}
	if len(o.dialects) == 0 {
		t.Fatal("no SMB2 NEGOTIATE followed the session request")
	}
	for _, d := range o.dialects {
		if d == smb2dialects.SMB2_DIALECT_3_1_1 {
			t.Errorf("SMB 3.1.1 offered over NetBIOS: %v", o.dialects)
		}
	}
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer c.Disconnect()
	if c.Dialect() != smb.SMB_VERSION_3_0_2 {
		t.Errorf("Dialect = %v, want SMB 3.0.2", c.Dialect())
	}
}

// decodeNetBIOSName reverses the RFC 1001 14.1 half-ASCII encoding of a
// 32-character NetBIOS name label.
func decodeNetBIOSName(label []byte) string {
	out := make([]byte, 0, len(label)/2)
	for i := 0; i+1 < len(label); i += 2 {
		out = append(out, (label[i]-'A')<<4|(label[i+1]-'A'))
	}
	return string(out)
}
