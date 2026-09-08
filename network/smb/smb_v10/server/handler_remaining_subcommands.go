package server

import (
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// handleOpen2 answers TRANS2_OPEN2: the TRANSACTION2 open, whose AccessMode and
// OpenMode are the same fields SMB_COM_OPEN_ANDX carries.
//
// Request parameters, per [MS-CIFS] section 2.2.6.1.1: Flags(2) AccessMode(2)
// Reserved1(2) FileAttributes(2) CreationTime(4) OpenMode(2) AllocationSize(4)
// Reserved[5](10) FileName(variable).
//
// The extended attributes a client may send in the data block are not applied:
// nothing here stores them, and the response reports an EA length of zero, which
// is how the format says a file has none. Accepting them silently would tell the
// client they had been kept.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleOpen2(
	conn *Connection,
	req *message.Message,
	reassembly *transactionReassembly,
) ([]byte, []byte, nt_status.NT_STATUS) {
	const fixed = 2 + 2 + 2 + 2 + 4 + 2 + 4 + 10

	if len(reassembly.parameters) < fixed {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	accessMode := binary.LittleEndian.Uint16(reassembly.parameters[2:4])
	openMode := binary.LittleEndian.Uint16(reassembly.parameters[12:14])
	name := decodeWireString([]types.UCHAR(reassembly.parameters[fixed:]), req.Header.Flags2.IsUnicode())

	flags, status := coreOpenFlagsFor(accessMode)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, nil, status
	}
	applyOpenMode(&flags, openMode)

	open, existed, status := conn.coreOpen(req, name, flags)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, nil, status
	}

	attr, err := open.File.Stat()
	if err != nil {
		return nil, nil, statusForFSError(err)
	}

	// Response parameters, per [MS-CIFS] section 2.2.6.1.2: FID(2)
	// FileAttributes(2) CreationTime(4) FileDataSize(4) AccessMode(2)
	// ResourceType(2) NMPipeStatus(2) ActionTaken(2) Reserved(4)
	// ExtendedAttributeErrorOffset(2) ExtendedAttributeLength(4).
	parameters := make([]byte, 30)
	binary.LittleEndian.PutUint16(parameters[0:2], open.FID)
	binary.LittleEndian.PutUint16(parameters[2:4], legacyAttributesFor(attr))
	binary.LittleEndian.PutUint32(parameters[4:8], utimeOf(attr.Created))
	binary.LittleEndian.PutUint32(parameters[8:12], uint32(attr.Size))
	binary.LittleEndian.PutUint16(parameters[12:14], accessMode)
	binary.LittleEndian.PutUint16(parameters[14:16], resourceTypeDisk)
	binary.LittleEndian.PutUint16(parameters[18:20], open2ActionTaken(existed))

	return parameters, nil, nt_status.NT_STATUS_SUCCESS
}

// The ActionTaken values of a TRANS2_OPEN2 response ([MS-CIFS] section 2.2.6.1.2).
const (
	open2ActionOpened    = 0x0001
	open2ActionCreated   = 0x0002
	open2ActionTruncated = 0x0003
)

// open2ActionTaken reports what the open did, which is the only way a client can
// tell a file it created from one it found.
func open2ActionTaken(existed bool) uint16 {
	if existed {
		return open2ActionOpened
	}
	return open2ActionCreated
}

// handleTrans2CreateDirectory answers TRANS2_CREATE_DIRECTORY.
//
// [MS-CIFS] section 2.2.6.14: "The directory MUST NOT exist. If the directory does
// exist, the request MUST fail and the server MUST return
// STATUS_OBJECT_NAME_COLLISION." The backend's Mkdir already reports that, so the
// status comes from there rather than from a separate check.
//
// Request parameters: Reserved(4) DirectoryName(variable).
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleTrans2CreateDirectory(
	conn *Connection,
	req *message.Message,
	reassembly *transactionReassembly,
) ([]byte, []byte, nt_status.NT_STATUS) {
	const fixed = 4

	if len(reassembly.parameters) < fixed {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	tree, status := conn.treeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, nil, status
	}
	if tree.Share.FS == nil {
		return nil, nil, nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}
	if tree.Share.ReadOnly {
		return nil, nil, nt_status.NT_STATUS_MEDIA_WRITE_PROTECTED
	}

	name := decodeWireString([]types.UCHAR(reassembly.parameters[fixed:]), req.Header.Flags2.IsUnicode())
	path, err := resolvePath(name)
	if err != nil {
		return nil, nil, nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}

	if err := tree.Share.FS.Mkdir(path); err != nil {
		logger.Debugf("SMB1 server: %s created directory %q, which failed: %v", conn.Remote, path, err)
		return nil, nil, statusForFSError(err)
	}

	// Response parameters are a 2-byte EaErrorOffset, which is zero because no
	// extended attributes were applied.
	return make([]byte, 2), nil, nt_status.NT_STATUS_SUCCESS
}

// handleTrans2SetFsInformation answers TRANS2_SET_FS_INFORMATION with the status
// the specification mandates for it.
//
// [MS-CIFS] section 2.2.6.5: the subcommand "is reserved but not implemented" and
// "Servers receiving requests with this command code MUST return
// STATUS_SMB_NO_SUPPORT (ERRSRV/ERRnosupport)". It is handled explicitly rather
// than left to fall through to STATUS_NOT_IMPLEMENTED because the specification
// names a different status for it than for its neighbours.
//
// STATUS_SMB_NO_SUPPORT has no constant in windows/nt_status, so the closest
// available is used: NT_STATUS_NOT_SUPPORTED, which carries the same meaning and
// maps to a not-supported legacy code. Nothing sends this subcommand — it is
// reserved — so the difference is a conformance detail rather than a live one.
func handleTrans2SetFsInformation(
	conn *Connection,
	_ *message.Message,
	_ *transactionReassembly,
) ([]byte, []byte, nt_status.NT_STATUS) {
	logger.Debugf("SMB1 server: %s sent TRANS2_SET_FS_INFORMATION, which is reserved and refused", conn.Remote)
	return nil, nil, nt_status.NT_STATUS_NOT_SUPPORTED
}

// handleNtTransactCreate answers NT_TRANSACT_CREATE, the NT_TRANSACT open.
//
// It is the same open as SMB_COM_NT_CREATE_ANDX with the parameters in a
// transaction rather than in the command words, so it shares openFlagsFor and the
// same read-only enforcement.
//
// Request parameters, per [MS-CIFS] section 2.2.7.1.1: Flags(4)
// RootDirectoryFID(4) DesiredAccess(4) AllocationSize(8) ExtFileAttributes(4)
// ShareAccess(4) CreateDisposition(4) CreateOptions(4) SecurityDescriptorLength(4)
// EALength(4) NameLength(4) ImpersonationLevel(4) SecurityFlags(1) Name(NameLength).
//
// The security descriptor and extended attributes a client may send in the data
// block are not applied: a descriptor is answered through the NT_TRANSACT security
// subcommands by a share that has a provider, and applying one here would let a
// create set access this server cannot then honour.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleNtTransactCreate(
	conn *Connection,
	req *message.Message,
	reassembly *transactionReassembly,
) ([]byte, []byte, nt_status.NT_STATUS) {
	const fixed = 4 + 4 + 4 + 8 + 4 + 4 + 4 + 4 + 4 + 4 + 4 + 4 + 1

	if len(reassembly.parameters) < fixed {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	rootDirectoryFID := binary.LittleEndian.Uint32(reassembly.parameters[4:8])
	desiredAccess := binary.LittleEndian.Uint32(reassembly.parameters[8:12])
	createDisposition := binary.LittleEndian.Uint32(reassembly.parameters[28:32])
	createOptions := binary.LittleEndian.Uint32(reassembly.parameters[32:36])
	nameLength := int(binary.LittleEndian.Uint32(reassembly.parameters[40:44]))

	// RootDirectoryFID names a directory handle the name is relative to. Nothing
	// here can resolve one, and treating the name as share-relative instead would
	// open a different file and report success.
	if rootDirectoryFID != 0 {
		logger.Debugf("SMB1 server: %s created relative to FID 0x%08X, which is refused",
			conn.Remote, rootDirectoryFID)
		return nil, nil, nt_status.NT_STATUS_NOT_SUPPORTED
	}

	if nameLength < 0 || fixed+nameLength > len(reassembly.parameters) {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}
	name := decodeWireString([]types.UCHAR(reassembly.parameters[fixed:fixed+nameLength]),
		req.Header.Flags2.IsUnicode())

	flags, status := openFlagsFor(desiredAccess, createDisposition, createOptions)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, nil, status
	}

	open, existed, status := conn.coreOpen(req, name, flags)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, nil, status
	}
	open.DeleteOnClose = createOptions&fileDeleteOnClose != 0

	attr, err := open.File.Stat()
	if err != nil {
		return nil, nil, statusForFSError(err)
	}

	// Response parameters are 69 bytes, per [MS-CIFS] section 2.2.7.1.2.
	parameters := make([]byte, 69)
	// OpLockLevel stays zero: no oplock is granted, and claiming one the server
	// cannot break would have the client cache on a promise nothing keeps.
	binary.LittleEndian.PutUint16(parameters[2:4], open.FID)
	binary.LittleEndian.PutUint32(parameters[4:8], createActionFor(existed, flags))
	binary.LittleEndian.PutUint64(parameters[12:20], filetimeOf(attr.Created))
	binary.LittleEndian.PutUint64(parameters[20:28], filetimeOf(attr.Accessed))
	binary.LittleEndian.PutUint64(parameters[28:36], filetimeOf(attr.Modified))
	binary.LittleEndian.PutUint64(parameters[36:44], filetimeOf(attr.Changed))
	binary.LittleEndian.PutUint32(parameters[44:48], attributesFor(attr))
	binary.LittleEndian.PutUint64(parameters[48:56], uint64(attr.AllocationSize))
	binary.LittleEndian.PutUint64(parameters[56:64], uint64(attr.Size))
	binary.LittleEndian.PutUint16(parameters[64:66], resourceTypeDisk)
	if attr.IsDir {
		parameters[68] = 1
	}

	return parameters, nil, nt_status.NT_STATUS_SUCCESS
}

// handleNtTransactRename answers NT_TRANSACT_RENAME with the status the
// specification mandates.
//
// [MS-CIFS] section 2.2.7.5: the subcommand "was reserved but not implemented" and
// "Servers receiving requests with this subcommand code MUST return
// STATUS_SMB_BAD_COMMAND (ERRSRV/ERRbadcmd)".
//
// It is handled explicitly rather than left to fall through, because the fall
// through answers STATUS_NOT_IMPLEMENTED and the specification names a different
// status here. Renaming through NT_TRANSACT was never a real operation, and
// SMB_COM_NT_RENAME is the command that does it.
func handleNtTransactRename(
	conn *Connection,
	_ *message.Message,
	_ *transactionReassembly,
) ([]byte, []byte, nt_status.NT_STATUS) {
	logger.Debugf("SMB1 server: %s sent NT_TRANSACT_RENAME, which is reserved and refused", conn.Remote)
	return nil, nil, nt_status.NT_STATUS_SMB_BAD_COMMAND
}

// fileDeleteOnClose is the CreateOptions bit that marks a handle for deletion.
const fileDeleteOnClose = 0x00001000
