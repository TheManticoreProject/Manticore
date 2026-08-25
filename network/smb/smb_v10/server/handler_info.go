package server

import (
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// handleQueryPathInformation answers TRANS2_QUERY_PATH_INFORMATION: it describes a
// path without opening it, which is how a client inspects a file it does not
// intend to read.
//
// Request parameters: InformationLevel(2) Reserved(4) FileName(variable).
func handleQueryPathInformation(conn *Connection, req *message.Message, reassembly *transactionReassembly) ([]byte, []byte, nt_status.NT_STATUS) {
	parameters := reassembly.parameters
	if len(parameters) < 6 {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	tree, status := conn.treeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, nil, status
	}

	level := binary.LittleEndian.Uint16(parameters[0:2])
	requested := trimTerminator(string(parameters[6:]))

	path, err := resolvePath(requested)
	if err != nil {
		logger.Debugf("SMB1 server: %s asked about %q, which is refused: %v", conn.Remote, requested, err)
		return nil, nil, nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}

	attr, err := tree.Share.FS.Stat(path)
	if err != nil {
		return nil, nil, statusForFSError(err)
	}

	data, served := encodeFileInformation(level, attr, path)
	if !served {
		logger.Debugf("SMB1 server: %s asked for file information level 0x%04X, which is not served",
			conn.Remote, level)
		return nil, nil, nt_status.NT_STATUS_INVALID_INFO_CLASS
	}

	// The response parameters carry an EaErrorOffset, which is zero when no
	// extended attribute was involved.
	return make([]byte, 2), data, nt_status.NT_STATUS_SUCCESS
}

// handleQueryFileInformation answers TRANS2_QUERY_FILE_INFORMATION: the same, for
// a path a client already holds open.
//
// Request parameters: FID(2) InformationLevel(2).
func handleQueryFileInformation(conn *Connection, req *message.Message, reassembly *transactionReassembly) ([]byte, []byte, nt_status.NT_STATUS) {
	parameters := reassembly.parameters
	if len(parameters) < 4 {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	fid := binary.LittleEndian.Uint16(parameters[0:2])
	level := binary.LittleEndian.Uint16(parameters[2:4])

	open, status := conn.openFor(req, fid)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, nil, status
	}

	// Asking the handle rather than the path: the two can differ if the file was
	// renamed underneath, and the handle is what the client is talking about.
	attr, err := open.File.Stat()
	if err != nil {
		return nil, nil, statusForFSError(err)
	}

	data, served := encodeFileInformation(level, attr, open.Path)
	if !served {
		return nil, nil, nt_status.NT_STATUS_INVALID_INFO_CLASS
	}

	return make([]byte, 2), data, nt_status.NT_STATUS_SUCCESS
}

// handleSetPathInformation answers TRANS2_SET_PATH_INFORMATION.
//
// Request parameters: InformationLevel(2) Reserved(4) FileName(variable).
func handleSetPathInformation(conn *Connection, req *message.Message, reassembly *transactionReassembly) ([]byte, []byte, nt_status.NT_STATUS) {
	parameters := reassembly.parameters
	if len(parameters) < 6 {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	tree, status := conn.writableTreeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, nil, status
	}

	level := binary.LittleEndian.Uint16(parameters[0:2])
	requested := trimTerminator(string(parameters[6:]))

	path, err := resolvePath(requested)
	if err != nil {
		return nil, nil, nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}
	if path == "" {
		return nil, nil, nt_status.NT_STATUS_ACCESS_DENIED
	}

	// A path-based set has no handle, so the levels that are properties of a
	// handle rather than of a file cannot be applied through it.
	served, err := applyFileInformation(tree.Share.FS, path, level, reassembly.data, nil)
	if !served {
		return nil, nil, nt_status.NT_STATUS_INVALID_INFO_CLASS
	}
	if err != nil {
		return nil, nil, statusForFSError(err)
	}

	return make([]byte, 2), nil, nt_status.NT_STATUS_SUCCESS
}

// handleSetFileInformation answers TRANS2_SET_FILE_INFORMATION.
//
// Request parameters: FID(2) InformationLevel(2) Reserved(2).
func handleSetFileInformation(conn *Connection, req *message.Message, reassembly *transactionReassembly) ([]byte, []byte, nt_status.NT_STATUS) {
	parameters := reassembly.parameters
	if len(parameters) < 4 {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	fid := binary.LittleEndian.Uint16(parameters[0:2])
	level := binary.LittleEndian.Uint16(parameters[2:4])

	open, status := conn.openFor(req, fid)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, nil, status
	}
	if open.Tree.Share.ReadOnly {
		return nil, nil, nt_status.NT_STATUS_MEDIA_WRITE_PROTECTED
	}
	if !open.Writable {
		return nil, nil, nt_status.NT_STATUS_ACCESS_DENIED
	}

	served, err := applyFileInformation(open.Tree.Share.FS, open.Path, level, reassembly.data, open)
	if !served {
		return nil, nil, nt_status.NT_STATUS_INVALID_INFO_CLASS
	}
	if err != nil {
		return nil, nil, statusForFSError(err)
	}

	return make([]byte, 2), nil, nt_status.NT_STATUS_SUCCESS
}

// handleQueryFsInformation answers TRANS2_QUERY_FS_INFORMATION: it describes the
// volume behind the share.
//
// Request parameters: InformationLevel(2).
func handleQueryFsInformation(conn *Connection, req *message.Message, reassembly *transactionReassembly) ([]byte, []byte, nt_status.NT_STATUS) {
	parameters := reassembly.parameters
	if len(parameters) < 2 {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	tree, status := conn.treeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, nil, status
	}

	level := binary.LittleEndian.Uint16(parameters[0:2])

	volume, err := tree.Share.FS.VolumeInfo()
	if err != nil {
		return nil, nil, statusForFSError(err)
	}
	// A share reports its own name as the label when the backend has none, which
	// is what a client displays.
	if volume.Label == "" {
		volume.Label = tree.Share.Name
	}

	data, served := encodeVolumeInformation(level, volume)
	if !served {
		logger.Debugf("SMB1 server: %s asked for volume information level 0x%04X, which is not served",
			conn.Remote, level)
		return nil, nil, nt_status.NT_STATUS_INVALID_INFO_CLASS
	}

	// A volume query returns no response parameters.
	return nil, data, nt_status.NT_STATUS_SUCCESS
}
