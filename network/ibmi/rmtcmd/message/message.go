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
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/startup"
)

// Request and reply IDs.
const (
	RequestExchangeSeeds             = startup.RequestExchangeSeeds
	ReplyExchangeSeeds               = startup.ReplyExchangeSeeds
	RequestStartServer               = startup.RequestStartServer
	ReplyStartServer                 = startup.ReplyStartServer
	RequestExchangeAttributes uint16 = 0x1001
	ReplyExchangeAttributes   uint16 = 0x8001
	RequestEndServer          uint16 = 0x1004
)

// Code points used by the startup exchanges.
const (
	CPUserID             = startup.CPUserID
	CPPasswordSubstitute = startup.CPPasswordSubstitute
	CPJobIdentity        = startup.CPJobIdentity
	CPAuthToken          = startup.CPAuthToken
)

// maxReplyFields bounds the LL/CP fields accepted in one startup reply.
const maxReplyFields = 64

// body returns a reply's bytes after the header, whatever its declared
// template length.
func body(f hostserver.Frame) []byte { return startup.Body(f) }

func short(what string, have, want int) error { return startup.Short(what, have, want) }
