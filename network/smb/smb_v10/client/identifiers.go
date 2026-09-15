package client

import (
	"os"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// A request is identified on a connection by the pair (PID, MID): the process
// that issued it, and which of that process's commands it is ([MS-CIFS]
// 2.2.3.1). The server echoes the pair back in the response, and the low half
// of the PID is what names the owner of a byte-range lock ([MS-CIFS]
// 2.2.4.32.1). A constant in either field makes every request on the connection
// indistinguishable from every other one, so both halves are filled in here
// rather than at each call site.

// clientProcessID is the PID stamped on every request this process sends. It is
// resolved once: a peer expects the identifier to be stable for the life of the
// connection, and to differ between processes on the same host.
var clientProcessID = processIdentifier(os.Getpid())

// processIdentifier turns an operating-system process identifier into the value
// carried in PIDHigh/PIDLow. Only the low 16 bits reach a lock range, so an
// identifier that is an exact multiple of 65536 would name lock owner zero —
// the very value this is replacing. Such an identifier is nudged by one; every
// other one is carried through unchanged.
func processIdentifier(pid int) types.ULONG {
	id := types.ULONG(uint32(pid))
	if id&0xFFFF == 0 {
		id |= 1
	}
	return id
}

// firstMID is the first multiplex identifier handed out on a connection. Zero is
// skipped so that a request is never confused with a header whose MID was simply
// never set.
const firstMID = 1

// nextMID returns the next multiplex identifier for a request on this
// connection. Identifiers are handed out in sequence and wrap back to firstMID,
// so a response can be attributed to the request it answers for as long as no
// more than 65535 requests are outstanding at once — far beyond any MaxMpxCount
// a server advertises.
func (c *Connection) nextMID() types.USHORT {
	next := uint16(c.messageIDCounter.Add(1))
	if next < firstMID {
		next = uint16(c.messageIDCounter.Add(1))
	}
	return types.USHORT(next)
}
