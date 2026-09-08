package server

import (
	"time"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// The SMB_FILE_ATTRIBUTES bits ([MS-CIFS] section 2.2.1.2.4).
//
// These are the legacy 16-bit attributes, not the 32-bit extended ones. The
// overlapping bits happen to share values, but SMB_FILE_ATTRIBUTE_NORMAL is
// 0x0000 rather than the extended form's 0x0080, so a normal file is described by
// the absence of every bit. Reporting the extended value in this field would set
// a bit the specification reserves.
const (
	smbFileAttributeNormal    = 0x0000
	smbFileAttributeReadOnly  = 0x0001
	smbFileAttributeHidden    = 0x0002
	smbFileAttributeSystem    = 0x0004
	smbFileAttributeDirectory = 0x0010
	smbFileAttributeArchive   = 0x0020
)

// legacyAttributesFor renders what the backend reported as SMB_FILE_ATTRIBUTES.
func legacyAttributesFor(attr FileAttr) uint16 {
	attributes := uint16(smbFileAttributeNormal)
	if attr.IsDir {
		attributes |= smbFileAttributeDirectory
	}
	if attr.ReadOnly {
		attributes |= smbFileAttributeReadOnly
	}
	return attributes
}

// utimeOf renders a time as UTIME: seconds since 1 January 1970 ([MS-CIFS]
// section 2.2.1.4.3).
//
// The zero time renders as zero, which is what the set commands read as "leave
// this alone", so a backend that reports no timestamp does not accidentally claim
// the epoch.
func utimeOf(when time.Time) uint32 {
	if when.IsZero() {
		return 0
	}
	seconds := when.UTC().Unix()
	if seconds < 0 {
		return 0
	}
	return uint32(seconds)
}

// timeFromUTIME converts a UTIME back to a time, mapping zero to the zero time so
// a caller can tell "unset" from "the epoch".
func timeFromUTIME(utime uint32) time.Time {
	if utime == 0 {
		return time.Time{}
	}
	return time.Unix(int64(utime), 0).UTC()
}

// handleQueryInformation answers SMB_COM_QUERY_INFORMATION: the core-set query of
// a file's attributes, write time and size, by path.
//
// [MS-CIFS] section 3.3.5.11 describes the answer in terms of
// FILE_NETWORK_OPEN_INFORMATION, which is the same three values the backend's Stat
// already reports.
//
// FileSize is 32 bits and the specification is explicit that a larger file returns
// only the low 32 bits with "No error message [...] to indicate this condition",
// so a truncating cast here is the specified behaviour rather than an oversight.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleQueryInformation(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.QueryInformationRequest)
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

	path, err := resolvePath(decodeWireString(request.FileName.Buffer, req.Header.Flags2.IsUnicode()))
	if err != nil {
		return nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}

	attr, err := tree.Share.FS.Stat(path)
	if err != nil {
		logger.Debugf("SMB1 server: %s queried %q, which failed: %v", conn.Remote, path, err)
		return statusForFSError(err)
	}

	response := commands.NewQueryInformationResponse()
	response.FileAttributes.SetAttributes(legacyAttributesFor(attr))
	response.LastWriteTime = types.ULONG(utimeOf(attr.Modified))
	response.FileSize = types.ULONG(uint32(attr.Size))

	return conn.answer(w, response)
}

// handleSetInformation answers SMB_COM_SET_INFORMATION: the core-set change of a
// file's attributes and write time, by path.
//
// [MS-CIFS] section 3.3.5.12: a LastWriteTime of zero means the write time "MUST
// NOT be changed", which is why the mask below is built from the field rather than
// set unconditionally. Attributes carry no such sentinel, so they are always
// applied.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleSetInformation(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.SetInformationRequest)
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

	path, err := resolvePath(decodeWireString(request.FileName.Buffer, req.Header.Flags2.IsUnicode()))
	if err != nil {
		return nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}

	// The file has to exist: [MS-CIFS] section 3.3.5.12 names the status for one
	// that does not, and a backend asked to set attributes on a missing path
	// might otherwise create it.
	if _, err := tree.Share.FS.Stat(path); err != nil {
		return statusForFSError(err)
	}

	written := timeFromUTIME(uint32(request.LastWriteTime))
	attr := FileAttr{
		ReadOnly: request.FileAttributes.GetAttributes()&smbFileAttributeReadOnly != 0,
		Modified: written,
	}
	mask := AttrMask{ReadOnly: true, Modified: !written.IsZero()}

	if err := tree.Share.FS.SetAttr(path, attr, mask); err != nil {
		logger.Debugf("SMB1 server: %s set information on %q, which failed: %v", conn.Remote, path, err)
		return statusForFSError(err)
	}

	return conn.answer(w, commands.NewSetInformationResponse())
}

// handleQueryInformation2 answers SMB_COM_QUERY_INFORMATION2: the same
// information as SMB_COM_QUERY_INFORMATION but for an open handle, and with all
// three timestamps rather than one.
//
// The timestamps go out as MS-DOS date and time pairs, which have two-second
// resolution, so a client reading them back cannot see a finer difference than
// that. [MS-CIFS] section 3.3.5.28 requires the FID to name a regular file.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleQueryInformation2(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.QueryInformation2Request)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	open, status := conn.openFor(req, uint16(request.FID))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}
	if open.IsPipe {
		return nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}
	if open.Tree == nil || open.Tree.Share.FS == nil {
		return nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}

	attr, err := open.Tree.Share.FS.Stat(open.Path)
	if err != nil {
		return statusForFSError(err)
	}

	response := commands.NewQueryInformation2Response()
	response.CreateDate, response.CreationTime = dosDateTimeOf(attr.Created)
	response.LastAccessDate, response.LastAccessTime = dosDateTimeOf(attr.Accessed)
	response.LastWriteDate, response.LastWriteTime = dosDateTimeOf(attr.Modified)
	response.FileDataSize = types.ULONG(uint32(attr.Size))
	response.FileAllocationSize = types.ULONG(uint32(attr.AllocationSize))
	response.FileAttributes.SetAttributes(legacyAttributesFor(attr))

	return conn.answer(w, response)
}

// handleSetInformation2 answers SMB_COM_SET_INFORMATION2: it sets the three
// timestamps of an open handle.
//
// A date and time pair of all zeroes means the client is not setting that stamp,
// which is the same convention every other set path here follows. Without it a
// client changing only the write time would move the other two to 1 January 1980,
// which is what an all-zero MS-DOS date denotes.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleSetInformation2(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.SetInformation2Request)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	open, status := conn.openFor(req, uint16(request.FID))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}
	if open.IsPipe {
		return nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}
	if open.Tree == nil || open.Tree.Share.FS == nil {
		return nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}
	if open.Tree.Share.ReadOnly {
		return nt_status.NT_STATUS_MEDIA_WRITE_PROTECTED
	}

	created := timeFromDosDateTime(request.CreateDate, request.CreationTime)
	accessed := timeFromDosDateTime(request.LastAccessDate, request.LastAccessTime)
	written := timeFromDosDateTime(request.LastWriteDate, request.LastWriteTime)

	attr := FileAttr{Created: created, Accessed: accessed, Modified: written}
	mask := AttrMask{
		Created:  !created.IsZero(),
		Accessed: !accessed.IsZero(),
		Modified: !written.IsZero(),
	}

	if err := open.Tree.Share.FS.SetAttr(open.Path, attr, mask); err != nil {
		logger.Debugf("SMB1 server: %s set times on %q, which failed: %v", conn.Remote, open.Path, err)
		return statusForFSError(err)
	}

	return conn.answer(w, commands.NewSetInformation2Response())
}

// dosDateTimeOf splits a time into the MS-DOS date and time pair the core-set
// commands carry.
//
// The zero time produces an all-zero pair, which is how "no timestamp" travels in
// a format that has no other way to say it.
func dosDateTimeOf(when time.Time) (types.SMB_DATE, types.SMB_TIME_DOS) {
	if when.IsZero() {
		return types.SMB_DATE{}, types.SMB_TIME_DOS{}
	}
	utc := when.UTC()
	date := types.NewSMB_DATEFromDate(utc.Year(), int(utc.Month()), utc.Day())
	clock := types.NewSMB_TIME_DOSFromTime(utc.Hour(), utc.Minute(), utc.Second())
	return *date, *clock
}

// timeFromDosDateTime reassembles a time from an MS-DOS date and time pair,
// returning the zero time for an all-zero pair.
//
// MS-DOS time has two-second resolution, so the seconds come back rounded down to
// an even value. That is a property of the format rather than of this conversion.
func timeFromDosDateTime(date types.SMB_DATE, clock types.SMB_TIME_DOS) time.Time {
	if date.Year == 0 && date.Month == 0 && date.Day == 0 &&
		clock.Hours == 0 && clock.Minutes == 0 && clock.TwoSeconds == 0 {
		return time.Time{}
	}

	// SMB_DATE counts years from 1980 and the accessors already return the
	// calendar values, so they are used directly.
	return time.Date(
		int(date.Year), time.Month(date.Month), int(date.Day),
		int(clock.Hours), int(clock.Minutes), int(clock.TwoSeconds)*2,
		0, time.UTC,
	)
}
