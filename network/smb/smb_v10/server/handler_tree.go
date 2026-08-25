package server

import (
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// nativeFileSystemName is reported for a disk share. A client uses it to decide
// what the storage supports, so it names something with the semantics this server
// actually offers rather than something more capable.
const nativeFileSystemName = "NTFS"

// handleTreeConnectAndx answers SMB_COM_TREE_CONNECT_ANDX: it connects a tree to
// a named share and returns the identifier the client uses to act on it.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/1e0e0e1d-9a1e-4d0e-9d1a-b1c9e5a8b2d3
func handleTreeConnectAndx(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.TreeConnectAndxRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	session := conn.Session(uint16(req.Header.UID))
	if session == nil {
		// The dispatcher checks this, so reaching here means the tables
		// disagree; refusing is still the right answer.
		return nt_status.NT_STATUS_SMB_BAD_UID
	}

	// The path is a UNC path, encoded in whichever character set THIS REQUEST
	// declared. Unicode is a per-message property: this repository's client
	// negotiates it on the connection but sends its tree connect in OEM, so
	// taking the connection's setting here decodes the path as garbage.
	unicode := req.Header.Flags2.IsUnicode()
	path := decodeWireString(request.Path, unicode)
	name := shareNameFromPath(path)
	if name == "" {
		logger.Debugf("SMB1 server: %s sent a tree connect to %q, which is not a UNC path", conn.Remote, path)
		return nt_status.NT_STATUS_BAD_NETWORK_NAME
	}

	share := conn.Server.Share(name)
	if share == nil {
		logger.Debugf("SMB1 server: %s asked for share %q, which is not served", conn.Remote, name)
		return nt_status.NT_STATUS_BAD_NETWORK_NAME
	}

	// A client may say what kind of resource it expects. Answering a mismatch
	// here is better than letting it discover the difference through a command
	// that makes no sense on the resource it actually got.
	if wanted := ShareType(decodeOEMString(request.Service)); wanted != "" && wanted != ShareTypeAny && wanted != share.Type {
		logger.Debugf("SMB1 server: %s asked for share %q as %q, but it is %q",
			conn.Remote, name, wanted, share.Type)
		return nt_status.NT_STATUS_BAD_DEVICE_TYPE
	}

	tid, err := conn.tids.Allocate()
	if err != nil {
		logger.Warnf("SMB1 server: refusing a tree connect from %s: %v", conn.Remote, err)
		return nt_status.NT_STATUS_INSUFF_SERVER_RESOURCES
	}

	tree := &Tree{
		TID:        tid,
		Share:      share,
		SessionUID: session.UID,
		Created:    time.Now().UTC(),
	}
	conn.addTree(tree)

	logger.Debugf("SMB1 server: %s connected share %q as TID 0x%04X", conn.Remote, share.Name, tid)

	response := commands.NewTreeConnectAndxResponse()
	// The Service string is OEM even when the connection negotiated Unicode,
	// which the response structure's own documentation notes.
	response.Service = []types.UCHAR(string(share.Type))
	if share.Type == ShareTypeDisk {
		response.NativeFileSystem = encodeNativeString(nativeFileSystemName, unicode)
	} else {
		// A resource with no file system behind it reports the empty string.
		response.NativeFileSystem = encodeNativeString("", unicode)
	}

	w.SetResponseTID(tid)
	if err := w.WriteResponse(response); err != nil {
		logger.Debugf("SMB1 server: failed to answer the tree connect for %s: %v", conn.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// handleTreeDisconnect answers SMB_COM_TREE_DISCONNECT: it drops the tree and
// everything opened on it.
func handleTreeDisconnect(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	tid := uint16(req.Header.TID)

	tree := conn.removeTree(tid)
	if tree == nil {
		return nt_status.NT_STATUS_SMB_BAD_TID
	}

	logger.Debugf("SMB1 server: %s disconnected share %q on TID 0x%04X", conn.Remote, tree.Share.Name, tid)

	if err := w.WriteResponse(commands.NewTreeDisconnectResponse()); err != nil {
		logger.Debugf("SMB1 server: failed to answer the tree disconnect for %s: %v", conn.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// decodeWireString reads a string field that carries wire bytes, in whichever
// character set the connection negotiated, and trims the terminator.
func decodeWireString(raw []types.UCHAR, useUnicode bool) string {
	if useUnicode {
		return trimTerminator(utf16.DecodeUTF16LE(raw))
	}
	return trimTerminator(string(raw))
}

// decodeOEMString reads a field that is OEM regardless of what the connection
// negotiated, which a few fields are.
func decodeOEMString(raw []types.UCHAR) string {
	return trimTerminator(string(raw))
}

// trimTerminator removes the NUL terminator a wire string ends with.
//
// Only a trailing terminator is removed. A NUL in the middle is left in place
// deliberately, so that a path carrying one is refused by resolvePath rather than
// silently truncated at it: truncating would mean a client asking for
// "file.txt\x00.png" was quietly given "file.txt", which is the whole hazard an
// embedded NUL represents. Two consumers disagreeing about where a string ends is
// resolved by rejecting the string, not by picking one of the answers.
func trimTerminator(value string) string {
	return strings.TrimRight(value, "\x00")
}
