package message_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd/message"
)

func FuzzParseStartupReplies(f *testing.F) {
	f.Add([]byte{0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8})
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0, 10, 0x11, 0x1F, 0, 0, 0, 0x25})
	f.Add(make([]byte, 16))
	f.Fuzz(func(t *testing.T, body []byte) {
		fr := hostserver.Frame{Variable: body}
		if r, err := message.ParseExchangeSeedsReply(fr); err == nil && r.ReturnCode == 0 && len(body) < 12 {
			t.Fatalf("seed from a %d-byte body", len(body))
		}
		if r, err := message.ParseStartServerReply(fr); err == nil && r.HasJobIdentity && len(r.JobIdentity) < 4 {
			t.Fatalf("job identity of %d bytes", len(r.JobIdentity))
		}
		if _, err := message.ParseExchangeAttributesReply(fr); err == nil && len(body) < 16 {
			t.Fatalf("0x8001 from a %d-byte body", len(body))
		}
		_, _ = message.ParseOperationReply(fr)
	})
}
