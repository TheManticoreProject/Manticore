package message_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon/message"
)

func FuzzParseReplies(f *testing.F) {
	ok := []byte{0, 0, 0, 0}
	for _, fl := range []hostserver.Field{
		{CodePoint: message.CPVersion, Value: []byte{0, 5, 4, 0}},
		{CodePoint: message.CPDataStreamLevel, Value: []byte{0, 2}},
		{CodePoint: message.CPSeed, Value: []byte{1, 2, 3, 4, 5, 6, 7, 8}},
		{CodePoint: message.CPPasswordLevel, Value: []byte{0}},
		{CodePoint: message.CPServerCCSID, Value: []byte{0, 0, 0, 37}},
	} {
		ok, _ = hostserver.AppendField(ok, fl.CodePoint, fl.Value)
	}
	f.Add(ok)
	f.Add([]byte{0, 3, 0, 0x0B})
	f.Fuzz(func(t *testing.T, body []byte) {
		fr := hostserver.Frame{Variable: body}
		if r, err := message.ParseExchangeAttributesReply(fr); err == nil && r.ReturnCode == 0 && r.ServerSeed == ([8]byte{}) && len(body) < 30 {
			t.Fatalf("accepted a successful 0xF003 without a seed: % x", body)
		}
		if r, err := message.ParseRetrieveInfoReply(fr); err == nil && r.HasServerCCSID && len(body) < 14 {
			t.Fatalf("CCSID from a %d-byte body", len(body))
		}
	})
}
