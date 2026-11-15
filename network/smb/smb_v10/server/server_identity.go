package server

import (
	"net"
	"os"
	"sort"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// StableServerGUID returns a ServerGUID that stays the same across process
// restarts for one host, deriving it from the host identity (host name plus the
// sorted hardware addresses of the non-loopback interfaces). A real Windows
// server keeps a fixed ServerGuid for the host; regenerating a fresh one on every
// start is a server fingerprint, since an observer that sees the same host over
// time watches the GUID change.
//
// When no host identity is available the function falls back to a fresh random
// GUID, which preserves the previous per-start behaviour rather than emitting a
// predictable constant. The derived GUID is still a well-formed RFC 4122 version
// 4 GUID, and the ServerGuid is sent in the clear during NEGOTIATE, so deriving
// it from the public host identity leaks nothing a peer could not already see.
//
// The SMB2 server shares this helper (it imports this package) so both servers
// present one consistent, stable identity.
func StableServerGUID() guid.GUID {
	seed := hostIdentitySeed()
	if len(seed) == 0 {
		return *guid.NewGUID()
	}
	return *guid.NewGUIDFromSeed(seed)
}

// hostIdentitySeed builds a stable per-host seed from the host name and the
// non-loopback interface hardware addresses. The addresses are sorted so the seed
// does not depend on interface enumeration order. It returns nil when neither a
// host name nor any hardware address is available, so the caller can fall back.
func hostIdentitySeed() []byte {
	var parts []string

	if name, err := os.Hostname(); err == nil && name != "" {
		parts = append(parts, "host="+name)
	}

	if ifaces, err := net.Interfaces(); err == nil {
		var macs []string
		for _, ifc := range ifaces {
			if ifc.Flags&net.FlagLoopback != 0 || len(ifc.HardwareAddr) == 0 {
				continue
			}
			macs = append(macs, ifc.HardwareAddr.String())
		}
		sort.Strings(macs)
		parts = append(parts, macs...)
	}

	if len(parts) == 0 {
		return nil
	}
	return []byte(strings.Join(parts, "|"))
}
