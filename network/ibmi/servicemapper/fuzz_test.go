package servicemapper_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ibmi/servicemapper"
)

func FuzzDecodeReply(f *testing.F) {
	f.Add([]byte{'+', 0, 0, 0x21, 0x1b})
	f.Add([]byte{'-'})
	f.Add([]byte{'+', 0xFF, 0xFF, 0xFF, 0xFF})
	f.Fuzz(func(t *testing.T, b []byte) {
		port, err := servicemapper.DecodeReply("as-rmtcmd", b)
		if err == nil && (port < 1 || port > 65535 || len(b) != servicemapper.ReplyLength) {
			t.Fatalf("accepted port %d from % x", port, b)
		}
	})
}
