package client

import (
	"os"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// A request is identified on a connection by the pair (PID, MID): the process
// that issued it, and which of that process's commands it is ([MS-CIFS]
// 2.2.3.1). The server echoes the pair back in the response, and the low half
// of the PID is what names the owner of a byte-range lock ([MS-CIFS]
// 2.2.4.32.1).

// clientProcessID is the PID stamped on every request this process sends. It is
// resolved once: a peer expects the identifier to be stable for the life of the
// connection, and to differ between processes on the same host.
var clientProcessID = processIdentifier(os.Getpid())

// processIdentifier turns an operating-system process identifier into the
// 16-bit value Windows carries in PIDLow. PIDHigh remains zero, avoiding both a
// host-PID leak and a wire shape Windows does not use. Zero cannot identify a
// lock owner, so an exact multiple of 65536 is nudged to one.
func processIdentifier(pid int) types.ULONG {
	id := types.ULONG(uint16(pid))
	if id == 0 {
		id = 1
	}
	return id
}

// nextMID returns the next multiplex identifier for a request on this
// connection. Zero is a valid MID and is handed out first, matching Windows;
// identifiers then advance through the entire 16-bit space before wrapping.
// A response can be attributed to the request it answers for as long as no more
// than 65536 requests are outstanding at once — far beyond any MaxMpxCount a
// server advertises.
func (c *Connection) nextMID() types.USHORT {
	return types.USHORT(uint16(c.messageIDCounter.Add(1) - 1))
}
