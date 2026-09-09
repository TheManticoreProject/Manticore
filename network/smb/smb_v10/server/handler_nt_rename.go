package server

import (
	"strings"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// The InformationLevel values of SMB_COM_NT_RENAME ([MS-CIFS] section 2.2.4.66.1).
const (
	smbNtRenameSetLinkInfo = 0x0103
	smbNtRenameRenameFile  = 0x0104
)

// handleNtRename answers SMB_COM_NT_RENAME, which renames a file or gives it a
// second name depending on its information level.
//
// [MS-CIFS] section 3.3.5.51 defines three outcomes, and all three are here: at
// SMB_NT_RENAME_RENAME_FILE the request "is treated as if it is an SMB_COM_RENAME
// Request"; at SMB_NT_RENAME_SET_LINK_INFO "the original file MUST NOT be renamed.
// Instead, the server MUST attempt to create a hard link"; and any other level
// "SHOULD fail the request with STATUS_INVALID_SMB".
//
// That last clause is why the default below is INVALID_SMB rather than
// INVALID_PARAMETER or NOT_SUPPORTED: the specification names the status, and a
// client distinguishing "I asked for something unsupported" from "I asked for
// something that is not a level" would be misled by either alternative.
//
// Wire format: [MS-CIFS] section 2.2.4.66.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleNtRename(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.NtRenameRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	tree, status := conn.treeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}
	if tree.Share.FS == nil {
		return nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}
	if tree.Share.ReadOnly {
		return nt_status.NT_STATUS_MEDIA_WRITE_PROTECTED
	}

	unicode := req.Header.Flags2.IsUnicode()
	oldName := decodeWireString(request.OldFileName.Buffer, unicode)
	newName := decodeWireString(request.NewFileName.Buffer, unicode)

	// "OldFileName MUST NOT contain wildcard characters; otherwise, the server
	// MUST return an error response with a Status of
	// STATUS_OBJECT_PATH_SYNTAX_BAD." Unlike SMB_COM_RENAME, which takes a
	// pattern, this command acts on exactly one named file.
	if strings.ContainsAny(oldName, "*?") {
		logger.Debugf("SMB1 server: %s sent an NT rename of %q, which carries a wildcard", conn.Remote, oldName)
		return nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}

	source, err := resolvePath(oldName)
	if err != nil {
		return nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}
	target, err := resolvePath(newName)
	if err != nil {
		return nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}

	// The source has to exist and match the search attributes, which is what the
	// specification checks before looking at the level.
	attr, err := tree.Share.FS.Stat(source)
	if err != nil {
		return statusForFSError(err)
	}
	if !matchesSearchAttributes(attr, request.SearchAttributes.GetAttributes()) {
		logger.Debugf("SMB1 server: %s renamed %q, which does not match the search attributes 0x%04X",
			conn.Remote, source, request.SearchAttributes.GetAttributes())
		return nt_status.NT_STATUS_NO_SUCH_FILE
	}

	switch uint16(request.InformationLevel) {
	case smbNtRenameRenameFile:
		// Treated as SMB_COM_RENAME: the target must not already exist, which is
		// what a rename with replace=false gives.
		conn.breakOplocksOn(tree.Share, source, nil)
		if err := tree.Share.FS.Rename(source, target, false); err != nil {
			logger.Debugf("SMB1 server: %s renamed %q to %q, which failed: %v",
				conn.Remote, source, target, err)
			return statusForFSError(err)
		}
		logger.Debugf("SMB1 server: %s renamed %q to %q", conn.Remote, source, target)

	case smbNtRenameSetLinkInfo:
		linker, ok := tree.Share.FS.(Linker)
		if !ok {
			// A backend that cannot link says so, rather than the server
			// substituting a copy: a copy is a second file, not a second name,
			// and the two diverge as soon as either is written.
			logger.Debugf("SMB1 server: %s asked to link %q, which share %q cannot do",
				conn.Remote, source, tree.Share.Name)
			return nt_status.NT_STATUS_NOT_SUPPORTED
		}
		if err := linker.Link(source, target); err != nil {
			logger.Debugf("SMB1 server: %s linked %q to %q, which failed: %v",
				conn.Remote, source, target, err)
			return statusForFSError(err)
		}
		logger.Debugf("SMB1 server: %s linked %q as %q", conn.Remote, source, target)

	default:
		logger.Debugf("SMB1 server: %s sent an NT rename at level 0x%04X, which is not a level",
			conn.Remote, uint16(request.InformationLevel))
		return nt_status.NT_STATUS_INVALID_SMB
	}

	return conn.answer(w, commands.NewNtRenameResponse())
}

// matchesSearchAttributes reports whether an entry is one the request's
// SearchAttributes admit.
//
// The field is inclusive rather than exclusive: a directory is matched only when
// the directory bit is set, and a plain file always matches. That is what makes
// SearchAttributes able to say "rename this only if it is a file".
func matchesSearchAttributes(attr FileAttr, searchAttributes uint16) bool {
	if attr.IsDir {
		return searchAttributes&smbFileAttributeDirectory != 0
	}
	return true
}
