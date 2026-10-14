package structures_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgyolaus/structures"
)

func FuzzDecoders(f *testing.F) {
	f.Add(append(head("ALICE", 0xF0, 0xF0), groups(2, "DEV", "OPS")...))
	f.Add(make([]byte, structures.AUTU0250Length))
	f.Fuzz(func(t *testing.T, b []byte) {
		if r, err := structures.DecodeAUTU0200(b, codec); err == nil && (len(b) != structures.AUTU0200Length || len(r.GroupProfiles) > structures.MaxGroupProfiles) {
			t.Fatalf("accepted %d bytes with %d groups", len(b), len(r.GroupProfiles))
		}
		if r, err := structures.DecodeAUTU0250(b, codec); err == nil && (len(b) != structures.AUTU0250Length || len(r.GroupProfiles) > structures.MaxGroupProfiles) {
			t.Fatalf("accepted %d bytes with %d groups", len(b), len(r.GroupProfiles))
		}
		_, _ = structures.DecodeAUTU0100(b, codec)
		_, _ = structures.DecodeAUTU0150(b, codec)
	})
}
