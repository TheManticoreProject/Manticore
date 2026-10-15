package message_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd/message"
)

func golden(t *testing.T, dump string) []byte {
	t.Helper()
	b, err := hex.DecodeString(strings.Join(strings.Fields(dump), ""))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func marshal(t *testing.T, f hostserver.Frame) []byte {
	t.Helper()
	b, err := f.Marshal(0)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// Golden frames from iSeriesNavigatorReverse
// docs/10-v5r4-compatibility-and-golden-vectors.md, produced by the JTOpen
// 4.8 request constructors as a bytecode oracle.
func TestExchangeSeedsGolden(t *testing.T) {
	want := golden(t, `00 00 00 1c 01 00 e0 08 00 00 00 00 00 00 00 00
		00 08 70 01 01 02 03 04 05 06 07 08`)
	if got := marshal(t, message.ExchangeSeedsFrame([8]byte{1, 2, 3, 4, 5, 6, 7, 8})); !bytes.Equal(got, want) {
		t.Errorf("7001 =\n% x\nwant\n% x", got, want)
	}
}

func TestStartServerGolden(t *testing.T) {
	want := golden(t, `00 00 00 34 02 00 e0 08 00 00 00 00 00 00 00 00
		00 02 70 02 01 01 00 00 00 0e 11 05 a0 a1 a2 a3
		a4 a5 a6 a7 00 00 00 10 11 04 e3 c5 e2 e3 e4 e2
		c5 d9 40 40`)
	f, err := message.StartServerRequest{
		CredentialType: message.CredentialTypeDES,
		Substitute:     []byte{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7},
		Profile:        [10]byte{0xe3, 0xc5, 0xe2, 0xe3, 0xe4, 0xe2, 0xc5, 0xd9, 0x40, 0x40},
	}.Frame()
	if err != nil {
		t.Fatal(err)
	}
	if got := marshal(t, f); !bytes.Equal(got, want) {
		t.Errorf("7002 =\n% x\nwant\n% x", got, want)
	}
	if _, err := (message.StartServerRequest{CredentialType: message.CredentialTypeSHA1, Substitute: make([]byte, 8)}).Frame(); err == nil {
		t.Error("SHA-1 type with an 8-byte substitute accepted")
	}
	if _, err := (message.StartServerRequest{CredentialType: 0x05, Substitute: make([]byte, 64)}).Frame(); err == nil {
		t.Error("credential type 0x05 accepted")
	}
	if f, err := (message.StartServerRequest{CredentialType: message.CredentialTypeSHA512, Substitute: make([]byte, 64)}).Frame(); err != nil || f.Template[0] != 0x07 {
		t.Errorf("level-4 7002 = %v, %v", f.Template, err)
	}
	if _, err := (message.StartServerRequest{CredentialType: message.CredentialTypeSHA512, Substitute: make([]byte, 20)}).Frame(); err == nil {
		t.Error("SHA-512 type with a 20-byte substitute accepted")
	}
	if f := message.ExchangeSeedsFrameWithAttributes([8]byte{}, message.SeedAttributesLevel4); f.Header.Attributes != 0x0300 {
		t.Errorf("level-4 seed attributes = 0x%04X", f.Header.Attributes)
	}
}

func TestExchangeAttributesGolden(t *testing.T) {
	want := golden(t, `00 00 00 22 00 00 e0 08 00 00 00 00 00 00 00 00
		00 0e 10 01 00 00 34 b0 f2 f9 f2 f4 00 00 00 01
		00 00`)
	f, err := message.ExchangeAttributesFrame(0, "2924")
	if err != nil {
		t.Fatal(err)
	}
	if got := marshal(t, f); !bytes.Equal(got, want) {
		t.Errorf("1001 =\n% x\nwant\n% x", got, want)
	}
	if f, _ := message.ExchangeAttributesFrame(1, "2924"); f.Header.CorrelationID != 1 {
		t.Error("correlation not applied")
	}
	for _, nlv := range []string{"", "292", "29245", "29a4", "２９２４"} {
		if _, err := message.ExchangeAttributesFrame(1, nlv); err == nil {
			t.Errorf("NLV %q accepted", nlv)
		}
	}
}

func TestEndServerFrame(t *testing.T) {
	want := golden(t, `00 00 00 14 00 00 e0 08 00 00 00 00 00 00 00 00 00 00 10 04`)
	if got := marshal(t, message.EndServerFrame()); !bytes.Equal(got, want) {
		t.Errorf("1004 = % x", got)
	}
}

func frame(id uint16, attrs uint16, body []byte) hostserver.Frame {
	return hostserver.Frame{Header: hostserver.Header{ServerID: hostserver.ServerIDRemoteCommand, RequestID: id, Attributes: attrs}, Variable: body}
}

func TestParseExchangeSeedsReply(t *testing.T) {
	b := append([]byte{0, 0, 0, 0}, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28)
	b, _ = hostserver.AppendField(b, 0x112E, []byte{1})
	r, err := message.ParseExchangeSeedsReply(frame(message.ReplyExchangeSeeds, 0x0001, b))
	if err != nil || r.ServerSeed != [8]byte{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28} || r.ServerAttributes != 1 || len(r.Fields) != 1 {
		t.Fatalf("reply = %+v, %v", r, err)
	}
	if r, err := message.ParseExchangeSeedsReply(frame(0xF001, 0, []byte{0, 1, 0, 4})); err != nil || r.ReturnCode != 0x00010004 {
		t.Errorf("failure reply = %+v, %v", r, err)
	}
	for name, body := range map[string][]byte{
		"no return code": {0, 0},
		"short seed":     {0, 0, 0, 0, 1, 2, 3},
		"bad tail":       {0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 0, 0},
	} {
		if _, err := message.ParseExchangeSeedsReply(frame(0xF001, 0, body)); !errors.Is(err, hostserver.ErrMalformed) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
}

func TestParseStartServerReply(t *testing.T) {
	b := []byte{0, 0, 0, 0}
	b, _ = hostserver.AppendField(b, message.CPUserID, []byte{0, 0, 0, 0x25, 0xe3, 0xc5})
	b, _ = hostserver.AppendField(b, message.CPJobIdentity, []byte{0, 0, 0, 0x25, 0xd8, 0xe9})
	r, err := message.ParseStartServerReply(frame(message.ReplyStartServer, 0, b))
	if err != nil || !r.HasJobIdentity || !bytes.Equal(r.JobIdentity, []byte{0, 0, 0, 0x25, 0xd8, 0xe9}) {
		t.Fatalf("reply = %+v, %v", r, err)
	}
	if r, err := message.ParseStartServerReply(frame(0xF002, 0, []byte{0, 3, 0, 0x0B})); err != nil || r.ReturnCode != 0x0003000B || r.HasJobIdentity {
		t.Errorf("rejection = %+v, %v", r, err)
	}
	short, _ := hostserver.AppendField([]byte{0, 0, 0, 0}, message.CPJobIdentity, []byte{1, 2})
	if _, err := message.ParseStartServerReply(frame(0xF002, 0, short)); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("short job identity error = %v", err)
	}
	dup, _ := hostserver.AppendField(b, message.CPJobIdentity, []byte{0, 0, 0, 0})
	if _, err := message.ParseStartServerReply(frame(0xF002, 0, dup)); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("duplicate job identity error = %v", err)
	}
}

func TestParseExchangeAttributesReply(t *testing.T) {
	b := make([]byte, 16)
	binary.BigEndian.PutUint16(b[0:], 0x0106)
	binary.BigEndian.PutUint32(b[2:], 37)
	copy(b[6:], []byte{9, 9, 9, 9, 8, 8, 8, 8})
	binary.BigEndian.PutUint16(b[14:], 7)
	r, err := message.ParseExchangeAttributesReply(frame(message.ReplyExchangeAttributes, 0, b))
	if err != nil || r.ReturnCode != 0x0106 || r.JobCCSID != 37 || r.DataStreamLevel != 7 || r.Reserved != [8]byte{9, 9, 9, 9, 8, 8, 8, 8} {
		t.Fatalf("reply = %+v, %v", r, err)
	}
	if _, err := message.ParseExchangeAttributesReply(frame(0x8001, 0, b[:15])); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("short reply error = %v", err)
	}
}

func TestStartServerWithToken(t *testing.T) {
	token := bytes.Repeat([]byte{0x5A}, 32)
	f, err := message.StartServerRequest{CredentialType: message.CredentialTypeProfileToken, Substitute: token, Profile: [10]byte{1}}.Frame()
	if err != nil {
		t.Fatal(err)
	}
	want := append([]byte{0, 0, 0, 38, 0x11, 0x15}, token...)
	if f.Template[0] != 0x02 || f.Template[1] != 0x01 || !bytes.Equal(f.Variable, want) {
		t.Errorf("7002 with token = % x / % x", f.Template, f.Variable)
	}
}
