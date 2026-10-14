// Package message encodes and decodes the Remote Command host-server
// startup and shutdown exchanges: exchange seeds (0x7001/0xF001), start
// server (0x7002/0xF002), exchange attributes (0x1001/0x8001) and end
// server (0x1004).
//
// Layouts follow the JTOpen 4.8 generation that the V5R4 baseline targets,
// as recorded in iSeriesNavigatorReverse docs/04-host-server-core-protocol.md,
// docs/06-remote-command-and-program-call.md,
// docs/19-remote-command-replies-and-messages.md and the golden frames in
// docs/10-v5r4-compatibility-and-golden-vectors.md.
package message

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// Request and reply IDs.
const (
	RequestExchangeSeeds      uint16 = 0x7001
	ReplyExchangeSeeds        uint16 = 0xF001
	RequestStartServer        uint16 = 0x7002
	ReplyStartServer          uint16 = 0xF002
	RequestExchangeAttributes uint16 = 0x1001
	ReplyExchangeAttributes   uint16 = 0x8001
	RequestEndServer          uint16 = 0x1004
)

// Code points used by the startup exchanges.
const (
	CPUserID             uint16 = 0x1104
	CPPasswordSubstitute uint16 = 0x1105
	CPJobIdentity        uint16 = 0x111F
)

// maxReplyFields bounds the LL/CP fields accepted in one startup reply.
const maxReplyFields = 64

// body returns a reply's bytes after the header, whatever its declared
// template length.
func body(f hostserver.Frame) []byte {
	b := make([]byte, 0, len(f.Template)+len(f.Variable))
	return append(append(b, f.Template...), f.Variable...)
}

func short(what string, have, want int) error {
	return &hostserver.MalformedError{What: what, Offset: hostserver.HeaderLength, Reason: fmt.Sprintf("%d-byte body, need at least %d", have, want)}
}
