package rpcpipe

import (
	"strings"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
)

// Share types reported by a share enumeration ([MS-SRVS] 2.2.2.4).
//
// The low byte names the kind of resource and the high bits carry flags, which is
// why STYPE_MASK exists: a client reads the kind by masking, so a share that is
// both a disk tree and administrative reports 0x80000000 and not some third kind.
const (
	// STYPE_DISKTREE is a disk directory tree.
	STYPE_DISKTREE uint32 = 0x00000000

	// STYPE_PRINTQ is a print queue.
	STYPE_PRINTQ uint32 = 0x00000001

	// STYPE_DEVICE is a communication device.
	STYPE_DEVICE uint32 = 0x00000002

	// STYPE_IPC is an interprocess communication share, which is what IPC$ is.
	STYPE_IPC uint32 = 0x00000003

	// STYPE_MASK selects the resource kind out of a share type.
	STYPE_MASK uint32 = 0x000000FF

	// STYPE_TEMPORARY marks a share that is not persisted.
	STYPE_TEMPORARY uint32 = 0x40000000

	// STYPE_SPECIAL marks an administrative share: one whose name ends in "$"
	// and which a client hides from an ordinary listing.
	STYPE_SPECIAL uint32 = 0x80000000
)

// ShareTypeOf reports the STYPE_* value for a server share.
//
// The flag bits are part of the answer and not a decoration: a client decides
// whether to show a share in a browse listing from STYPE_SPECIAL, so reporting
// IPC$ or C$ without it makes an administrative share look like an ordinary one.
//
// Parameters:
//   - name: the share name, which decides whether it is administrative
//   - shareType: the share's Service string
//
// Returns:
//   - The share type an enumeration should report
func ShareTypeOf(name string, shareType server.ShareType) uint32 {
	reported := STYPE_DISKTREE
	switch shareType {
	case server.ShareTypeNamedPipe:
		reported = STYPE_IPC
	case server.ShareTypePrinter:
		reported = STYPE_PRINTQ
	}

	// A name ending in "$" is administrative by convention, which is the only
	// thing that marks one: SMB carries no separate flag for it.
	if strings.HasSuffix(name, "$") {
		reported |= STYPE_SPECIAL
	}
	return reported
}
