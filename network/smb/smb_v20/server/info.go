package server

import (
	"encoding/binary"
	"hash/fnv"
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/logger"
	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
	"github.com/TheManticoreProject/Manticore/windows/filesystem/infoclass"
)

// QUERY_INFO and SET_INFO InfoType values (MS-SMB2 2.2.37).
const (
	infoFile       = 0x01
	infoFileSystem = 0x02
)

// Information classes served that windows/filesystem has no structure for.
const (
	fileModeInformation         = 16
	fileAlignmentInformation    = 17
	fileStreamInformation       = 22
	fileAttributeTagInformation = 35
)

// volumeAttributes are the FileSystemAttributes claimed: case-preserved and
// Unicode names, nothing this server does not implement.
const volumeAttributes = 0x00000002 | 0x00000004

// queryInfo answers QUERY_INFO for the file and file-system classes.
func (c *Connection) queryInfo(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.QueryInfoRequest)
	open, status := c.openFor(r, req.FileId)
	if !status.IsSuccess() {
		return status
	}

	var encoded []byte
	variable := false
	switch req.InfoType {
	case infoFile:
		attr, err := open.fs().Stat(open.path)
		if err != nil {
			return smb1.StatusForFSError(err)
		}
		encoded, variable, status = encodeFileInfo(req.FileInfoClass, open, attr)
	case infoFileSystem:
		volume, err := open.fs().VolumeInfo()
		if err != nil {
			return smb1.StatusForFSError(err)
		}
		encoded, status = encodeVolumeInfo(req.FileInfoClass, volume)
	default:
		return nt_status.NT_STATUS_NOT_SUPPORTED
	}
	if !status.IsSuccess() {
		return status
	}

	// A fixed-size class that does not fit is refused; a variable one is cut.
	if len(encoded) > int(req.OutputBufferLength) {
		if !variable {
			return nt_status.NT_STATUS_INFO_LENGTH_MISMATCH
		}
		encoded, status = encoded[:req.OutputBufferLength], nt_status.NT_STATUS_BUFFER_OVERFLOW
	}
	answer := commands.NewQueryInfoResponse()
	answer.OutputBuffer = encoded
	resp.command = answer
	return status
}

// encodeFileInfo renders a file information class, reporting whether its size
// depends on the file.
func encodeFileInfo(class uint8, open *Open, attr smb1.FileAttr) ([]byte, bool, nt_status.NT_STATUS) {
	var encoded []byte
	var err error
	variable := false
	switch infoclass.FileInformationClass(class) {
	case infoclass.FileBasicInformation:
		basic := basicInfo(attr)
		encoded, err = basic.Marshal()
	case infoclass.FileStandardInformation:
		standard := standardInfo(attr, open.deleteOnClose)
		encoded, err = standard.Marshal()
	case infoclass.FileInternalInformation:
		encoded = binary.LittleEndian.AppendUint64(nil, fileIndex(open.path))
	case infoclass.FileEaInformation:
		encoded = make([]byte, 4)
	case infoclass.FileAccessInformation:
		encoded = binary.LittleEndian.AppendUint32(nil, open.access)
	case infoclass.FilePositionInformation:
		// Every READ and WRITE carries its offset: there is no file pointer.
		encoded = make([]byte, 8)
	case fileModeInformation, fileAlignmentInformation:
		encoded = make([]byte, 4)
	case infoclass.FileNetworkOpenInformation:
		info := filesystem.FileNetworkOpenInformation{
			CreationTime: filetime(attr.Created), LastAccessTime: filetime(attr.Accessed),
			LastWriteTime: filetime(attr.Modified), ChangeTime: filetime(attr.Changed),
			AllocationSize: attr.AllocationSize, EndOfFile: attr.Size, FileAttributes: attributesOf(attr),
		}
		encoded, err = info.Marshal()
	case fileAttributeTagInformation:
		encoded = binary.LittleEndian.AppendUint32(nil, attributesOf(attr))
		encoded = binary.LittleEndian.AppendUint32(encoded, 0)
	case infoclass.FileAllInformation:
		all := filesystem.FileAllInformation{
			Basic: basicInfo(attr), Standard: standardInfo(attr, open.deleteOnClose),
			IndexNumber: int64(fileIndex(open.path)), AccessFlags: open.access,
			FileName: `\` + toWindowsPath(open.path),
		}
		encoded, err = all.Marshal()
		variable = true
	case infoclass.FileNameInformation, infoclass.FileAlternateNameInformation:
		// There are no 8.3 names, so the alternate name is the long one.
		name := `\` + toWindowsPath(open.path)
		if infoclass.FileInformationClass(class) == infoclass.FileAlternateNameInformation {
			name = open.path[strings.LastIndexByte(open.path, '/')+1:]
		}
		encodedName := utf16.EncodeUTF16LE(name)
		encoded = append(binary.LittleEndian.AppendUint32(nil, uint32(len(encodedName))), encodedName...)
		variable = true
	case fileStreamInformation:
		variable = true
		if !attr.IsDir {
			// The unnamed data stream, the only one there is.
			name := utf16.EncodeUTF16LE("::$DATA")
			encoded = binary.LittleEndian.AppendUint32(make([]byte, 4), uint32(len(name)))
			encoded = binary.LittleEndian.AppendUint64(encoded, uint64(attr.Size))
			encoded = binary.LittleEndian.AppendUint64(encoded, uint64(attr.AllocationSize))
			encoded = append(encoded, name...)
		}
	default:
		return nil, false, nt_status.NT_STATUS_INVALID_INFO_CLASS
	}
	if err != nil {
		return nil, false, nt_status.NT_STATUS_INTERNAL_ERROR
	}
	return encoded, variable, nt_status.NT_STATUS_SUCCESS
}

// encodeVolumeInfo renders a file-system information class.
func encodeVolumeInfo(class uint8, volume smb1.VolumeInfo) ([]byte, nt_status.NT_STATUS) {
	sectors, sectorSize := max(volume.SectorsPerAllocationUnit, 1), max(volume.BytesPerSector, 512)
	unit := int64(sectors) * int64(sectorSize)

	var encoded []byte
	var err error
	switch infoclass.FsInformationClass(class) {
	case infoclass.FileFsVolumeInformation:
		info := filesystem.FileFsVolumeInformation{VolumeSerialNumber: volume.SerialNumber, VolumeLabel: volume.Label}
		encoded, err = info.Marshal()
	case infoclass.FileFsSizeInformation:
		info := filesystem.FileFsSizeInformation{
			TotalAllocationUnits: volume.TotalBytes / unit, AvailableAllocationUnits: volume.FreeBytes / unit,
			SectorsPerAllocationUnit: sectors, BytesPerSector: sectorSize,
		}
		encoded, err = info.Marshal()
	case infoclass.FileFsFullSizeInformation:
		info := filesystem.FileFsFullSizeInformation{
			TotalAllocationUnits: volume.TotalBytes / unit, CallerAvailableAllocationUnits: volume.FreeBytes / unit,
			ActualAvailableAllocationUnits: volume.FreeBytes / unit, SectorsPerAllocationUnit: sectors, BytesPerSector: sectorSize,
		}
		encoded, err = info.Marshal()
	case infoclass.FileFsDeviceInformation:
		const fileDeviceDisk = 0x07
		info := filesystem.FileFsDeviceInformation{DeviceType: fileDeviceDisk}
		encoded, err = info.Marshal()
	case infoclass.FileFsAttributeInformation:
		name := volume.FileSystemName
		if name == "" {
			name = "NTFS"
		}
		info := filesystem.FileFsAttributeInformation{
			FileSystemAttributes: volumeAttributes, MaximumComponentNameLength: smb1.MaxPathComponentLength, FileSystemName: name,
		}
		encoded, err = info.Marshal()
	default:
		return nil, nt_status.NT_STATUS_INVALID_INFO_CLASS
	}
	if err != nil {
		return nil, nt_status.NT_STATUS_INTERNAL_ERROR
	}
	return encoded, nt_status.NT_STATUS_SUCCESS
}

// setInfo answers SET_INFO for the file classes that change a file.
func (c *Connection) setInfo(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.SetInfoRequest)
	open, status := c.openFor(r, req.FileId)
	if !status.IsSuccess() {
		return status
	}
	if req.InfoType != infoFile {
		return nt_status.NT_STATUS_NOT_SUPPORTED
	}
	if status := open.require(setInfoRights(req.FileInfoClass)); !status.IsSuccess() {
		return status
	}

	fs := open.fs()
	var err error
	switch infoclass.FileInformationClass(req.FileInfoClass) {
	case infoclass.FileBasicInformation:
		basic := filesystem.FileBasicInformation{}
		if basic.Unmarshal(req.Buffer) != nil {
			return nt_status.NT_STATUS_INVALID_PARAMETER
		}
		// A zero timestamp leaves that one alone.
		attr := smb1.FileAttr{
			Created: timeOf(basic.CreationTime), Accessed: timeOf(basic.LastAccessTime),
			Modified: timeOf(basic.LastWriteTime), Changed: timeOf(basic.ChangeTime),
			ReadOnly: basic.FileAttributes&fileflags.FILE_ATTRIBUTE_READONLY != 0,
		}
		err = fs.SetAttr(open.path, attr, smb1.AttrMask{
			ReadOnly: basic.FileAttributes != 0,
			Created:  !attr.Created.IsZero(), Accessed: !attr.Accessed.IsZero(),
			Modified: !attr.Modified.IsZero(), Changed: !attr.Changed.IsZero(),
		})
	case infoclass.FileEndOfFileInformation:
		eof := filesystem.FileEndOfFileInformation{}
		if eof.Unmarshal(req.Buffer) != nil || eof.EndOfFile < 0 || open.isDir {
			return nt_status.NT_STATUS_INVALID_PARAMETER
		}
		c.server.oplocks.breakOthers(open)
		err = open.file.Truncate(eof.EndOfFile)
	case infoclass.FileAllocationInformation:
		// Reserving space is not changing the length, and nothing here tells them apart.
	case infoclass.FileDispositionInformation:
		disposition := filesystem.FileDispositionInformation{}
		if disposition.Unmarshal(req.Buffer) != nil {
			return nt_status.NT_STATUS_INVALID_PARAMETER
		}
		if disposition.DeletePending && open.isDir {
			if entries, err := fs.ReadDir(open.path, ""); err == nil && len(entries) > 0 {
				return nt_status.NT_STATUS_DIRECTORY_NOT_EMPTY
			}
		}
		// A file carrying the read-only attribute cannot be marked for deletion,
		// [MS-FSA] section 2.1.5.14.3.
		if disposition.DeletePending && !open.isDir {
			if attr, err := fs.Stat(open.path); err == nil && attr.ReadOnly {
				return smb1.StatusForFSError(smb1.ErrCannotDelete)
			}
		}
		c.server.oplocks.breakOthers(open)
		open.deleteOnClose = disposition.DeletePending
	case infoclass.FileRenameInformation:
		rename := filesystem.FileRenameInformation{}
		if rename.Unmarshal(req.Buffer) != nil || rename.RootDirectory != 0 {
			return nt_status.NT_STATUS_INVALID_PARAMETER
		}
		target, perr := smb1.ResolvePath(rename.FileName)
		if perr != nil {
			return nt_status.NT_STATUS_OBJECT_NAME_INVALID
		}
		c.server.oplocks.breakOthers(open)
		if err = fs.Rename(open.path, target, rename.ReplaceIfExists); err == nil {
			c.server.oplocks.rename(open, target)
			c.server.shareModes.rename(open, target)
			open.path = target
		}
	default:
		return nt_status.NT_STATUS_INVALID_INFO_CLASS
	}
	if err != nil {
		return smb1.StatusForFSError(err)
	}
	resp.command = commands.NewSetInfoResponse()
	return nt_status.NT_STATUS_SUCCESS
}

// setInfoRights are the rights a SET_INFO class needs ([MS-SMB2] 3.3.5.21.1).
func setInfoRights(class uint8) uint32 {
	switch infoclass.FileInformationClass(class) {
	case infoclass.FileBasicInformation:
		return fileflags.FILE_WRITE_ATTRIBUTES
	case infoclass.FileDispositionInformation, infoclass.FileRenameInformation:
		return fileflags.DELETE
	}
	return fileflags.FILE_WRITE_DATA
}

// search is the state of a directory enumeration on an open.
type search struct {
	entries []smb1.DirEntry
	next    int
}

// queryDirectory answers QUERY_DIRECTORY (MS-SMB2 3.3.5.18).
func (c *Connection) queryDirectory(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.QueryDirectoryRequest)
	open, status := c.openFor(r, req.FileId)
	if !status.IsSuccess() {
		return status
	}
	if !open.isDir {
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}
	if status := open.require(fileflags.FILE_LIST_DIRECTORY); !status.IsSuccess() {
		return status
	}
	if !supportedDirectoryClass(req.FileInformationClass) {
		return nt_status.NT_STATUS_INVALID_INFO_CLASS
	}

	restart := req.Flags&(commands.SMB2_RESTART_SCANS|commands.SMB2_REOPEN) != 0
	if open.search == nil || restart {
		entries, err := open.fs().ReadDir(open.path, req.FileName)
		if err != nil {
			return smb1.StatusForFSError(err)
		}
		if len(entries) == 0 {
			return nt_status.NT_STATUS_NO_SUCH_FILE
		}
		open.search = &search{entries: entries}
	}
	s := open.search
	if s.next >= len(s.entries) {
		return nt_status.NT_STATUS_NO_MORE_FILES
	}

	// Entries are 8-byte aligned and chained; the last one ends the buffer.
	var output []byte
	last := 0
	for s.next < len(s.entries) {
		entry := encodeDirectoryEntry(req.FileInformationClass, open.path, s.entries[s.next].Attr)
		start := (len(output) + 7) &^ 7
		if start+len(entry) > int(req.OutputBufferLength) {
			break
		}
		if len(output) > 0 {
			binary.LittleEndian.PutUint32(output[last:], uint32(start-last))
		}
		output = append(output, make([]byte, start-len(output))...)
		output = append(output, entry...)
		last = start
		s.next++
		if req.Flags&commands.SMB2_RETURN_SINGLE_ENTRY != 0 {
			break
		}
	}
	if len(output) == 0 {
		return nt_status.NT_STATUS_INFO_LENGTH_MISMATCH
	}

	logger.Debugf("SMB2 server: %s listed %q (pattern %q, flags 0x%02x): %d of %d entries sent",
		c.remote, open.path, req.FileName, req.Flags, s.next, len(s.entries))
	answer := commands.NewQueryDirectoryResponse()
	answer.OutputBuffer = output
	resp.command = answer
	return nt_status.NT_STATUS_SUCCESS
}

// supportedDirectoryClass reports whether a listing can be returned in a class.
func supportedDirectoryClass(class uint8) bool {
	switch infoclass.FileInformationClass(class) {
	case infoclass.FileFullDirectoryInformation, infoclass.FileBothDirectoryInformation,
		infoclass.FileNamesInformation, infoclass.FileIdBothDirectoryInformation,
		infoclass.FileIdFullDirectoryInformation:
		return true
	}
	return false
}

// encodeDirectoryEntry renders one listing entry, its NextEntryOffset zero.
func encodeDirectoryEntry(class uint8, directory string, attr smb1.FileAttr) []byte {
	created, accessed, modified, changed := filetime(attr.Created), filetime(attr.Accessed), filetime(attr.Modified), filetime(attr.Changed)
	attributes := attributesOf(attr)
	id := fileIndex(joinPath(directory, attr.Name))

	var entry interface{ Marshal() ([]byte, error) }
	switch infoclass.FileInformationClass(class) {
	case infoclass.FileFullDirectoryInformation:
		entry = &filesystem.FileFullDirectoryInformation{
			CreationTime: created, LastAccessTime: accessed, LastWriteTime: modified, ChangeTime: changed,
			EndOfFile: attr.Size, AllocationSize: attr.AllocationSize, FileAttributes: attributes, FileName: attr.Name,
		}
	case infoclass.FileIdFullDirectoryInformation:
		entry = &filesystem.FileIdFullDirectoryInformation{
			CreationTime: created, LastAccessTime: accessed, LastWriteTime: modified, ChangeTime: changed,
			EndOfFile: attr.Size, AllocationSize: attr.AllocationSize, FileAttributes: attributes, FileId: id, FileName: attr.Name,
		}
	case infoclass.FileBothDirectoryInformation:
		entry = &filesystem.FileBothDirectoryInformation{
			CreationTime: created, LastAccessTime: accessed, LastWriteTime: modified, ChangeTime: changed,
			EndOfFile: attr.Size, AllocationSize: attr.AllocationSize, FileAttributes: attributes, FileName: attr.Name,
		}
	case infoclass.FileIdBothDirectoryInformation:
		entry = &filesystem.FileIdBothDirectoryInformation{
			CreationTime: created, LastAccessTime: accessed, LastWriteTime: modified, ChangeTime: changed,
			EndOfFile: attr.Size, AllocationSize: attr.AllocationSize, FileAttributes: attributes, FileId: id, FileName: attr.Name,
		}
	default:
		entry = &filesystem.FileNamesInformation{FileName: attr.Name}
	}
	encoded, _ := entry.Marshal()
	return encoded
}

func basicInfo(attr smb1.FileAttr) filesystem.FileBasicInformation {
	return filesystem.FileBasicInformation{
		CreationTime: filetime(attr.Created), LastAccessTime: filetime(attr.Accessed),
		LastWriteTime: filetime(attr.Modified), ChangeTime: filetime(attr.Changed), FileAttributes: attributesOf(attr),
	}
}

func standardInfo(attr smb1.FileAttr, deletePending bool) filesystem.FileStandardInformation {
	return filesystem.FileStandardInformation{
		AllocationSize: attr.AllocationSize, EndOfFile: attr.Size, NumberOfLinks: 1,
		DeletePending: deletePending, Directory: attr.IsDir,
	}
}

// attributesOf renders an entry as FILE_ATTRIBUTE bits. A regular file carries
// ARCHIVE, as on Windows, so NORMAL (valid only alone) is never needed.
func attributesOf(attr smb1.FileAttr) uint32 {
	attributes := fileflags.FILE_ATTRIBUTE_ARCHIVE
	if attr.IsDir {
		attributes = fileflags.FILE_ATTRIBUTE_DIRECTORY
	}
	if attr.ReadOnly {
		attributes |= fileflags.FILE_ATTRIBUTE_READONLY
	}
	return attributes
}

// fileIndex is a stable, non-zero identifier for a path. Clients use it as an
// inode number, so every file sharing one value would look like hard links.
func fileIndex(path string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(path))
	return h.Sum64() | 1
}

// unixEpochAsFiletime is 1970-01-01 in 100ns intervals since 1601-01-01.
const unixEpochAsFiletime = 116444736000000000

// filetime renders a time as a FILETIME; the zero time renders as zero.
func filetime(t time.Time) uint64 {
	if t.IsZero() {
		return 0
	}
	return uint64(t.UnixNano()/100 + unixEpochAsFiletime)
}

// timeOf converts a FILETIME; zero and the "leave alone" values give the zero time.
func timeOf(ft uint64) time.Time {
	if ft == 0 || ft == ^uint64(0) || ft < unixEpochAsFiletime {
		return time.Time{}
	}
	return time.Unix(0, int64(ft-unixEpochAsFiletime)*100).UTC()
}

func toWindowsPath(path string) string {
	return strings.ReplaceAll(path, "/", `\`)
}

func joinPath(directory, name string) string {
	if directory == "" {
		return name
	}
	return directory + "/" + name
}
