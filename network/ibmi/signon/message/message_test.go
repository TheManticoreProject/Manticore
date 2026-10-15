package message_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon/message"
)

// golden decodes a hex dump; spaces and newlines are ignored.
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
// docs/10-v5r4-compatibility-and-golden-vectors.md ("Sign-on attribute
// requests" and "Per-service seed and synthetic authentication layouts"),
// produced by the JTOpen 4.8 request constructors as a bytecode oracle.
func TestExchangeAttributesRequestGolden(t *testing.T) {
	noSeed := golden(t, `00 00 00 26 00 00 e0 09 00 00 00 00 00 00 00 00
		00 00 70 03 00 00 00 0a 11 01 00 00 00 01 00 00
		00 08 11 02 00 02`)
	if got := marshal(t, message.ExchangeAttributesRequest{}.Frame()); !bytes.Equal(got, noSeed) {
		t.Errorf("no-seed 7003 =\n% x\nwant\n% x", got, noSeed)
	}
	seed := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	withSeed := golden(t, `00 00 00 34 00 00 e0 09 00 00 00 00 00 00 00 00
		00 00 70 03 00 00 00 0a 11 01 00 00 00 01 00 00
		00 08 11 02 00 02 00 00 00 0e 11 03 01 02 03 04
		05 06 07 08`)
	if got := marshal(t, message.ExchangeAttributesRequest{ClientSeed: &seed}.Frame()); !bytes.Equal(got, withSeed) {
		t.Errorf("seeded 7003 =\n% x\nwant\n% x", got, withSeed)
	}
}

func TestRetrieveInfoRequestGolden(t *testing.T) {
	// The a0..a7 substitute is the golden file's visual marker, not a real
	// substitute; the profile is TESTUSER.
	want := golden(t, `00 00 00 3d 00 00 e0 09 00 00 00 00 00 00 00 00
		00 01 70 04 01 00 00 00 0a 11 13 00 00 34 b0 00
		00 00 0e 11 05 a0 a1 a2 a3 a4 a5 a6 a7 00 00 00
		10 11 04 e3 c5 e2 e3 e4 e2 c5 d9 40 40`)
	req := message.RetrieveInfoRequest{
		CredentialType: message.CredentialTypeDES,
		Substitute:     []byte{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7},
		Profile:        [10]byte{0xe3, 0xc5, 0xe2, 0xe3, 0xe4, 0xe2, 0xc5, 0xd9, 0x40, 0x40},
	}
	f, err := req.Frame()
	if err != nil {
		t.Fatal(err)
	}
	if got := marshal(t, f); !bytes.Equal(got, want) {
		t.Errorf("7004 =\n% x\nwant\n% x", got, want)
	}
	sha := req
	sha.CredentialType, sha.Substitute = message.CredentialTypeSHA1, make([]byte, 20)
	if f, err := sha.Frame(); err != nil || f.Template[0] != 0x03 {
		t.Errorf("SHA-1 7004 = %v, %v", f.Template, err)
	}
	l4 := req
	l4.CredentialType, l4.Substitute = message.CredentialTypeSHA512, make([]byte, 64)
	if f, err := l4.Frame(); err != nil || f.Template[0] != 0x07 {
		t.Errorf("level-4 7004 = %v, %v", f.Template, err)
	}
	for name, bad := range map[string]message.RetrieveInfoRequest{
		"DES with 20 bytes":     {CredentialType: message.CredentialTypeDES, Substitute: make([]byte, 20)},
		"SHA-1 with 8 bytes":    {CredentialType: message.CredentialTypeSHA1, Substitute: make([]byte, 8)},
		"SHA-512 with 20 bytes": {CredentialType: message.CredentialTypeSHA512, Substitute: make([]byte, 20)},
		"type 0x05":             {CredentialType: 0x05, Substitute: make([]byte, 64)},
	} {
		if _, err := bad.Frame(); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestEndServerFrame(t *testing.T) {
	want := golden(t, `00 00 00 14 00 00 e0 09 00 00 00 00 00 00 00 00 00 00 70 06`)
	if got := marshal(t, message.EndServerFrame()); !bytes.Equal(got, want) {
		t.Errorf("7006 = % x", got)
	}
}

// reply builds a reply frame: 4-byte return code template, then fields.
func reply(t *testing.T, id uint16, rc uint32, fields ...hostserver.Field) hostserver.Frame {
	t.Helper()
	var v []byte
	for _, f := range fields {
		v, _ = hostserver.AppendField(v, f.CodePoint, f.Value)
	}
	return hostserver.Frame{
		Header:   hostserver.Header{ServerID: hostserver.ServerIDSignon, RequestID: id},
		Template: []byte{byte(rc >> 24), byte(rc >> 16), byte(rc >> 8), byte(rc)},
		Variable: v,
	}
}

func f(cp uint16, v ...byte) hostserver.Field { return hostserver.Field{CodePoint: cp, Value: v} }

func TestParseExchangeAttributesReply(t *testing.T) {
	ok := reply(t, message.ReplyExchangeAttributes, 0,
		f(message.CPVersion, 0x00, 0x05, 0x04, 0x00),
		f(message.CPDataStreamLevel, 0x00, 0x02),
		f(message.CPSeed, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18),
		f(message.CPPasswordLevel, 0x00),
		f(message.CPJobIdentity, 0, 0, 0, 37, 0xC1),
		f(0x9999), f(0x9999), // unknown code points may repeat
	)
	r, err := message.ParseExchangeAttributesReply(ok)
	if err != nil {
		t.Fatal(err)
	}
	if r.ServerVersion != 0x00050400 || r.ServerLevel != 2 || r.ServerSeed != [8]byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18} || !r.HasPasswordLevel || r.PasswordLevel != 0 {
		t.Errorf("reply = %+v", r)
	}
	if _, has := r.Fields.Get(message.CPJobIdentity); !has {
		t.Error("unconsumed job identity field dropped")
	}

	// The return code is read from the first four body bytes even if the
	// header declares no template.
	flat := ok
	flat.Variable = append(append([]byte(nil), ok.Template...), ok.Variable...)
	flat.Template = nil
	if r2, err := message.ParseExchangeAttributesReply(flat); err != nil || r2.ServerVersion != r.ServerVersion {
		t.Errorf("template-less reply = %+v, %v", r2, err)
	}

	noLevel := reply(t, message.ReplyExchangeAttributes, 0,
		f(message.CPVersion, 0, 5, 4, 0), f(message.CPDataStreamLevel, 0, 2), f(message.CPSeed, 1, 1, 1, 1, 1, 1, 1, 1))
	if r, err := message.ParseExchangeAttributesReply(noLevel); err != nil || r.HasPasswordLevel {
		t.Errorf("reply without password level = %+v, %v", r, err)
	}
	if r, err := message.ParseExchangeAttributesReply(reply(t, message.ReplyExchangeAttributes, 0x00010004)); err != nil || r.ReturnCode != 0x00010004 {
		t.Errorf("failure reply = %+v, %v", r, err)
	}

	for name, fr := range map[string]hostserver.Frame{
		"no return code":  {Template: []byte{0, 0}},
		"missing seed":    reply(t, 0xF003, 0, f(message.CPVersion, 0, 5, 4, 0), f(message.CPDataStreamLevel, 0, 2)),
		"short seed":      reply(t, 0xF003, 0, f(message.CPVersion, 0, 5, 4, 0), f(message.CPDataStreamLevel, 0, 2), f(message.CPSeed, 1, 2, 3)),
		"duplicate seed":  reply(t, 0xF003, 0, f(message.CPVersion, 0, 5, 4, 0), f(message.CPDataStreamLevel, 0, 2), f(message.CPSeed, 1, 1, 1, 1, 1, 1, 1, 1), f(message.CPSeed, 2, 2, 2, 2, 2, 2, 2, 2)),
		"level two bytes": reply(t, 0xF003, 0, f(message.CPVersion, 0, 5, 4, 0), f(message.CPDataStreamLevel, 0, 2), f(message.CPSeed, 1, 1, 1, 1, 1, 1, 1, 1), f(message.CPPasswordLevel, 0, 2)),
		"missing version": reply(t, 0xF003, 0, f(message.CPDataStreamLevel, 0, 2), f(message.CPSeed, 1, 1, 1, 1, 1, 1, 1, 1)),
		"bad field LL":    {Template: []byte{0, 0, 0, 0}, Variable: []byte{0, 0, 0, 3, 0x11, 0x03}},
	} {
		if _, err := message.ParseExchangeAttributesReply(fr); !errors.Is(err, hostserver.ErrMalformed) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
}

func TestParseRetrieveInfoReply(t *testing.T) {
	r, err := message.ParseRetrieveInfoReply(reply(t, message.ReplyRetrieveInfo, 0,
		f(message.CPUserID, 0, 0, 0, 37, 0xe3, 0xc5), f(message.CPServerCCSID, 0, 0, 0, 37), f(0x1106, 0x07, 0xd6, 10, 13, 12, 0, 0)))
	if err != nil || !r.HasServerCCSID || r.ServerCCSID != 37 || r.ReturnCode != 0 {
		t.Fatalf("reply = %+v, %v", r, err)
	}
	if r, err := message.ParseRetrieveInfoReply(reply(t, message.ReplyRetrieveInfo, 0x0003000B)); err != nil || r.ReturnCode != 0x0003000B || r.HasServerCCSID {
		t.Errorf("rejection reply = %+v, %v", r, err)
	}
	if _, err := message.ParseRetrieveInfoReply(reply(t, message.ReplyRetrieveInfo, 0, f(message.CPServerCCSID, 0, 37))); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("short CCSID error = %v", err)
	}
}
