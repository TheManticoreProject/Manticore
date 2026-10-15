package startup

import (
	"bytes"
	"errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

func TestFramesCarryTheServerID(t *testing.T) {
	seeds := ExchangeSeedsFrame(0xE007, [8]byte{1, 2, 3, 4, 5, 6, 7, 8}, SeedAttributes)
	b, err := seeds.Marshal(0)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0, 0, 0, 28, 0x01, 0x00, 0xE0, 0x07, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 0x70, 0x01, 1, 2, 3, 4, 5, 6, 7, 8}
	if !bytes.Equal(b, want) {
		t.Errorf("0x7001 = % x", b)
	}
	f, err := StartServerRequest{ServerID: 0xE007, CredentialType: CredentialTypeProfileToken, Substitute: make([]byte, 32)}.Frame()
	if err != nil {
		t.Fatal(err)
	}
	if f.Header.ServerID != 0xE007 || f.Header.Attributes != 0x0200 || !bytes.Equal(f.Variable[:6], []byte{0, 0, 0, 38, 0x11, 0x15}) || len(f.Variable) != 38 {
		t.Errorf("token 0x7002 = %+v", f)
	}
	if _, err := (StartServerRequest{CredentialType: CredentialTypeSHA1, Substitute: make([]byte, 8)}).Frame(); err == nil {
		t.Error("a substitute of the wrong length was accepted")
	}
	if _, err := (StartServerRequest{CredentialType: 0x04, Substitute: make([]byte, 8)}).Frame(); err == nil {
		t.Error("an undefined credential type was accepted")
	}
}

func TestParseReplies(t *testing.T) {
	reply := func(body []byte) hostserver.Frame {
		return hostserver.Frame{Header: hostserver.Header{ServerID: 0xE007, RequestID: ReplyStartServer}, Variable: body}
	}
	r, err := ParseStartServerReply(reply([]byte{0, 0, 0, 0, 0, 0, 0, 10, 0x11, 0x1F, 0, 0, 0, 0x25}))
	if err != nil || r.ReturnCode != 0 || !r.HasJobIdentity || len(r.JobIdentity) != 4 {
		t.Fatalf("start = %+v, %v", r, err)
	}
	for name, body := range map[string][]byte{
		"short":            {0, 0, 0},
		"short identity":   {0, 0, 0, 0, 0, 0, 0, 9, 0x11, 0x1F, 0, 0, 0},
		"duplicate job ID": {0, 0, 0, 0, 0, 0, 0, 6, 0x11, 0x1F, 0, 0, 0, 6, 0x11, 0x1F},
	} {
		if _, err := ParseStartServerReply(reply(body)); !errors.Is(err, hostserver.ErrMalformed) {
			t.Errorf("%s: %v", name, err)
		}
	}
	s, err := ParseExchangeSeedsReply(reply([]byte{0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8}))
	if err != nil || s.ServerSeed != [8]byte{1, 2, 3, 4, 5, 6, 7, 8} {
		t.Errorf("seeds = %+v, %v", s, err)
	}
	if s, err := ParseExchangeSeedsReply(reply([]byte{0, 1, 0, 0})); err != nil || s.ReturnCode != 0x00010000 {
		t.Errorf("failed seeds = %+v, %v", s, err)
	}
	if _, err := ParseExchangeSeedsReply(reply([]byte{0, 0, 0, 0, 1})); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("short seed = %v", err)
	}
}
