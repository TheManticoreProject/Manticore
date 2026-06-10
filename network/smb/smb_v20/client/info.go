package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
	"github.com/TheManticoreProject/Manticore/windows/filesystem/infoclass"
)

// QueryInfo issues an SMB2 QUERY_INFO on the open identified by fileId and returns
// the raw output buffer (an MS-FSCC structure, a security descriptor, …). infoType
// is one of commands.SMB2_0_INFO_* and fileInfoClass the class within it;
// additionalInformation carries the security-info bits for security queries and is
// 0 otherwise. Wire: SMB2 QUERY_INFO.
func (c *Client) QueryInfo(fileId types.SMB2_FILEID, infoType, fileInfoClass uint8, additionalInformation uint32) ([]byte, error) {
	if c.Session == nil || c.Session.TreeId == 0 {
		return nil, fmt.Errorf("no tree connect established")
	}

	req := commands.NewQueryInfoRequest()
	req.InfoType = types.UCHAR(infoType)
	req.FileInfoClass = types.UCHAR(fileInfoClass)
	req.AdditionalInformation = types.ULONG(additionalInformation)
	req.OutputBufferLength = types.ULONG(c.Connection.Server.MaxTransactSize)
	req.FileId = fileId

	response, err := c.sendReceive(c.newRequest(req), "QueryInfo")
	if err != nil {
		return nil, err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return nil, fmt.Errorf("query info failed: %s", formatNTStatus(status))
	}
	resp, ok := response.Command.(*commands.QueryInfoResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected query info response command: %T", response.Command)
	}
	return resp.OutputBuffer, nil
}

// SetInfo issues an SMB2 SET_INFO on fileId, sending buffer as the information to
// set. Wire: SMB2 SET_INFO.
func (c *Client) SetInfo(fileId types.SMB2_FILEID, infoType, fileInfoClass uint8, additionalInformation uint32, buffer []byte) error {
	if c.Session == nil || c.Session.TreeId == 0 {
		return fmt.Errorf("no tree connect established")
	}

	req := commands.NewSetInfoRequest()
	req.InfoType = types.UCHAR(infoType)
	req.FileInfoClass = types.UCHAR(fileInfoClass)
	req.AdditionalInformation = types.ULONG(additionalInformation)
	req.FileId = fileId
	req.Buffer = buffer

	response, err := c.sendReceive(c.newRequest(req), "SetInfo")
	if err != nil {
		return err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return fmt.Errorf("set info failed: %s", formatNTStatus(status))
	}
	return nil
}

// QueryFileBasicInfo returns the FILE_BASIC_INFORMATION (timestamps, attributes)
// of an open file.
func (c *Client) QueryFileBasicInfo(fileId types.SMB2_FILEID) (*filesystem.FileBasicInformation, error) {
	raw, err := c.QueryInfo(fileId, commands.SMB2_0_INFO_FILE, uint8(infoclass.FileBasicInformation), 0)
	if err != nil {
		return nil, err
	}
	fi := &filesystem.FileBasicInformation{}
	if err := fi.Unmarshal(raw); err != nil {
		return nil, err
	}
	return fi, nil
}

// QueryFileStandardInfo returns the FILE_STANDARD_INFORMATION (size, link count,
// delete/directory flags) of an open file.
func (c *Client) QueryFileStandardInfo(fileId types.SMB2_FILEID) (*filesystem.FileStandardInformation, error) {
	raw, err := c.QueryInfo(fileId, commands.SMB2_0_INFO_FILE, uint8(infoclass.FileStandardInformation), 0)
	if err != nil {
		return nil, err
	}
	fi := &filesystem.FileStandardInformation{}
	if err := fi.Unmarshal(raw); err != nil {
		return nil, err
	}
	return fi, nil
}

// QueryFsSizeInfo returns the FILE_FS_SIZE_INFORMATION of the volume backing the
// open identified by fileId (any open on the tree, e.g. the share root).
func (c *Client) QueryFsSizeInfo(fileId types.SMB2_FILEID) (*filesystem.FileFsSizeInformation, error) {
	raw, err := c.QueryInfo(fileId, commands.SMB2_0_INFO_FILESYSTEM, uint8(infoclass.FileFsSizeInformation), 0)
	if err != nil {
		return nil, err
	}
	fi := &filesystem.FileFsSizeInformation{}
	if err := fi.Unmarshal(raw); err != nil {
		return nil, err
	}
	return fi, nil
}

// QueryFileAllInfo returns the FILE_ALL_INFORMATION (the aggregate basic,
// standard, internal, EA, access, position, mode, alignment, and name data) of an
// open file.
func (c *Client) QueryFileAllInfo(fileId types.SMB2_FILEID) (*filesystem.FileAllInformation, error) {
	raw, err := c.QueryInfo(fileId, commands.SMB2_0_INFO_FILE, uint8(infoclass.FileAllInformation), 0)
	if err != nil {
		return nil, err
	}
	fi := &filesystem.FileAllInformation{}
	if err := fi.Unmarshal(raw); err != nil {
		return nil, err
	}
	return fi, nil
}

// QueryFileNetworkOpenInfo returns the FILE_NETWORK_OPEN_INFORMATION (timestamps,
// sizes, attributes) of an open file.
func (c *Client) QueryFileNetworkOpenInfo(fileId types.SMB2_FILEID) (*filesystem.FileNetworkOpenInformation, error) {
	raw, err := c.QueryInfo(fileId, commands.SMB2_0_INFO_FILE, uint8(infoclass.FileNetworkOpenInformation), 0)
	if err != nil {
		return nil, err
	}
	fi := &filesystem.FileNetworkOpenInformation{}
	if err := fi.Unmarshal(raw); err != nil {
		return nil, err
	}
	return fi, nil
}

// QueryFsVolumeInfo returns the FILE_FS_VOLUME_INFORMATION (creation time, serial
// number, label) of the volume backing the open identified by fileId.
func (c *Client) QueryFsVolumeInfo(fileId types.SMB2_FILEID) (*filesystem.FileFsVolumeInformation, error) {
	raw, err := c.QueryInfo(fileId, commands.SMB2_0_INFO_FILESYSTEM, uint8(infoclass.FileFsVolumeInformation), 0)
	if err != nil {
		return nil, err
	}
	fi := &filesystem.FileFsVolumeInformation{}
	if err := fi.Unmarshal(raw); err != nil {
		return nil, err
	}
	return fi, nil
}

// QueryFsFullSizeInfo returns the FILE_FS_FULL_SIZE_INFORMATION (total and
// caller/actual available allocation units) of the volume backing fileId.
func (c *Client) QueryFsFullSizeInfo(fileId types.SMB2_FILEID) (*filesystem.FileFsFullSizeInformation, error) {
	raw, err := c.QueryInfo(fileId, commands.SMB2_0_INFO_FILESYSTEM, uint8(infoclass.FileFsFullSizeInformation), 0)
	if err != nil {
		return nil, err
	}
	fi := &filesystem.FileFsFullSizeInformation{}
	if err := fi.Unmarshal(raw); err != nil {
		return nil, err
	}
	return fi, nil
}

// QueryFsAttributeInfo returns the FILE_FS_ATTRIBUTE_INFORMATION (attribute flags,
// max component name length, file system name) of the volume backing fileId.
func (c *Client) QueryFsAttributeInfo(fileId types.SMB2_FILEID) (*filesystem.FileFsAttributeInformation, error) {
	raw, err := c.QueryInfo(fileId, commands.SMB2_0_INFO_FILESYSTEM, uint8(infoclass.FileFsAttributeInformation), 0)
	if err != nil {
		return nil, err
	}
	fi := &filesystem.FileFsAttributeInformation{}
	if err := fi.Unmarshal(raw); err != nil {
		return nil, err
	}
	return fi, nil
}

// QueryFsDeviceInfo returns the FILE_FS_DEVICE_INFORMATION (device type and
// characteristics) of the volume backing fileId.
func (c *Client) QueryFsDeviceInfo(fileId types.SMB2_FILEID) (*filesystem.FileFsDeviceInformation, error) {
	raw, err := c.QueryInfo(fileId, commands.SMB2_0_INFO_FILESYSTEM, uint8(infoclass.FileFsDeviceInformation), 0)
	if err != nil {
		return nil, err
	}
	fi := &filesystem.FileFsDeviceInformation{}
	if err := fi.Unmarshal(raw); err != nil {
		return nil, err
	}
	return fi, nil
}

// SetEndOfFile sets the logical size of an open file (truncate or extend).
func (c *Client) SetEndOfFile(fileId types.SMB2_FILEID, size int64) error {
	buf, err := (&filesystem.FileEndOfFileInformation{EndOfFile: size}).Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal FILE_END_OF_FILE_INFORMATION: %w", err)
	}
	return c.SetInfo(fileId, commands.SMB2_0_INFO_FILE, uint8(infoclass.FileEndOfFileInformation), 0, buf)
}

// SetDeleteOnClose marks (or unmarks) an open file for deletion when its last
// handle is closed.
func (c *Client) SetDeleteOnClose(fileId types.SMB2_FILEID, deletePending bool) error {
	buf, err := (&filesystem.FileDispositionInformation{DeletePending: deletePending}).Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal FILE_DISPOSITION_INFORMATION: %w", err)
	}
	return c.SetInfo(fileId, commands.SMB2_0_INFO_FILE, uint8(infoclass.FileDispositionInformation), 0, buf)
}

// RenameByHandle renames the open file to newName (share-relative). The handle
// must have been opened with DELETE access.
func (c *Client) RenameByHandle(fileId types.SMB2_FILEID, newName string, replaceIfExists bool) error {
	buf, err := (&filesystem.FileRenameInformation{ReplaceIfExists: replaceIfExists, FileName: newName}).Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal FILE_RENAME_INFORMATION: %w", err)
	}
	return c.SetInfo(fileId, commands.SMB2_0_INFO_FILE, uint8(infoclass.FileRenameInformation), 0, buf)
}
