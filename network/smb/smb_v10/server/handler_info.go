package server

import (
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
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
	// The path follows the request's own encoding, as every name-carrying field
	// does.
	requested := decodeWireString([]types.UCHAR(parameters[6:]), req.Header.Flags2.IsUnicode())

	path, err := resolvePath(requested)
	if err != nil {
		logger.Debugf("SMB1 server: %s asked about %q, which is refused: %v", conn.Remote, requested, err)
		return nil, nil, nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}

	attr, err := tree.Share.FS.Stat(path)
	if err != nil {
		return nil, nil, statusForFSError(err)
	}

	data, served := encodeFileInformation(level, attr, path, req.Header.Flags2.IsUnicode())
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
	//
	// A handle with no file behind it — a directory the backend declined to open,
	// or a pipe — is described another way rather than refused, because querying is
	// exactly what such a handle exists for.
	var attr FileAttr
	var err error
	switch {
	case open.IsPipe:
		// There is no file system entry to describe, so the handle describes
		// itself: a pipe has no size and no timestamps to report.
		attr = FileAttr{Name: open.Path}
	case open.File != nil:
		attr, err = open.File.Stat()
	case open.Tree != nil && open.Tree.Share.FS != nil:
		attr, err = open.Tree.Share.FS.Stat(open.Path)
	default:
		return nil, nil, nt_status.NT_STATUS_INVALID_HANDLE
	}
	if err != nil {
		return nil, nil, statusForFSError(err)
	}

	data, served := encodeFileInformation(level, attr, open.Path, req.Header.Flags2.IsUnicode())
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
	// The path follows the request's own encoding, as every name-carrying field
	// does.
	requested := decodeWireString([]types.UCHAR(parameters[6:]), req.Header.Flags2.IsUnicode())

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

	data, served := encodeVolumeInformation(level, volume, req.Header.Flags2.IsUnicode())
	if !served {
		logger.Debugf("SMB1 server: %s asked for volume information level 0x%04X, which is not served",
			conn.Remote, level)
		return nil, nil, nt_status.NT_STATUS_INVALID_INFO_CLASS
	}

	// A volume query returns no response parameters.
	return nil, data, nt_status.NT_STATUS_SUCCESS
}

// handleQueryInformationDisk answers SMB_COM_QUERY_INFORMATION_DISK, the legacy
// way to ask how much room a share has.
//
// It is superseded by the TRANSACTION2 volume levels, but a client still sends it —
// after a listing, whether or not anything asked for the free space — so a server
// that does not answer it makes every session report an error that nothing was
// waiting on.
//
// Its fields are 16-bit, so a volume larger than the counts can express has to be
// scaled rather than truncated: reporting the low 16 bits of a large number would
// describe a tiny volume, which is worse than describing a coarse one. The block
// size is raised until the unit count fits ([MS-CIFS] section 2.2.4.55.2 notes the
// same practice).
func handleQueryInformationDisk(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	tree, status := conn.treeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	volume, err := tree.Share.FS.VolumeInfo()
	if err != nil {
		return statusForFSError(err)
	}

	blockSize := volume.BytesPerSector
	if blockSize == 0 {
		blockSize = defaultBytesPerSector
	}
	blocksPerUnit := volume.SectorsPerAllocationUnit
	if blocksPerUnit == 0 {
		blocksPerUnit = defaultSectorsPerAllocationUnit
	}

	unit := int64(blocksPerUnit) * int64(blockSize)
	total, free := volume.TotalBytes/unit, volume.FreeBytes/unit
	// Scale by doubling the blocks per unit until the counts fit. A client
	// multiplies the three fields back together, so the product stays honest to
	// within one unit while the fields stay representable.
	for (total > 0xFFFF || free > 0xFFFF) && blocksPerUnit < 0x8000 {
		blocksPerUnit *= 2
		unit = int64(blocksPerUnit) * int64(blockSize)
		total, free = volume.TotalBytes/unit, volume.FreeBytes/unit
	}
	if total > 0xFFFF {
		total = 0xFFFF
	}
	if free > 0xFFFF {
		free = 0xFFFF
	}

	response := commands.NewQueryInformationDiskResponse()
	response.TotalUnits = types.USHORT(total)
	response.BlocksPerUnit = types.USHORT(blocksPerUnit)
	response.BlockSize = types.USHORT(blockSize)
	response.FreeUnits = types.USHORT(free)

	if err := w.WriteResponse(response); err != nil {
		logger.Debugf("SMB1 server: failed to answer the disk query for %s: %v", conn.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// Fallbacks for a backend that does not describe its allocation geometry. A client
// divides by these, so zero is not a usable answer.
const (
	defaultBytesPerSector           = 512
	defaultSectorsPerAllocationUnit = 8
)
