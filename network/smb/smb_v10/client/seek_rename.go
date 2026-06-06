package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// Seek mode values for Seek (MS-CIFS 2.2.4.43.1).
const (
	SeekModeStart   uint16 = 0 // offset is relative to the start of the file
	SeekModeCurrent uint16 = 1 // offset is relative to the current file pointer
	SeekModeEnd     uint16 = 2 // offset is relative to the end of the file
)

// NT_RENAME information levels (MS-CIFS 2.2.4.65.1).
const (
	infoNtRenameSetLinkInfo uint16 = 0x0103 // create a hard link
	infoNtRenameRenameFile  uint16 = 0x0104 // in-place rename
)

// Seek sets the file pointer of the open file referenced by fid and returns the
// resulting absolute file position (bytes from the start of the file). mode is one
// of SeekModeStart/SeekModeCurrent/SeekModeEnd and offset is a signed displacement.
//
// Wire: SMB_COM_SEEK request / response.
func (c *Client) Seek(fid FID, mode uint16, offset int32) (uint32, error) {
	if c.Session == nil {
		return 0, fmt.Errorf("no session established")
	}

	msg := c.newFileIOMessage(codes.SMB_COM_SEEK)

	cmd := commands.NewSeekRequest()
	cmd.FID = types.USHORT(fid)
	cmd.Mode = types.USHORT(mode)
	cmd.Offset = types.LONG(offset)

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "Seek")
	if err != nil {
		return 0, err
	}
	if response.Header.Status != 0x00000000 {
		return 0, fmt.Errorf("Seek failed: 0x%08x", response.Header.Status)
	}

	seekResponse, ok := response.Command.(*commands.SeekResponse)
	if !ok {
		return 0, fmt.Errorf("unexpected response command: 0x%02x", response.Header.Command)
	}

	return uint32(seekResponse.Offset), nil
}

// ntRename issues SMB_COM_NT_RENAME for the given information level. oldPath and
// newPath are share-relative paths using backslash separators.
func (c *Client) ntRename(oldPath, newPath string, informationLevel uint16) error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	msg := c.newFileIOMessage(codes.SMB_COM_NT_RENAME)

	cmd := commands.NewNtRenameRequest()
	cmd.SearchAttributes = types.SMB_FILE_ATTRIBUTES{}
	cmd.InformationLevel = types.USHORT(informationLevel)
	if err := cmd.OldFileName.SetString(oldPath); err != nil {
		return fmt.Errorf("failed to set old file name: %v", err)
	}
	if err := cmd.NewFileName.SetString(newPath); err != nil {
		return fmt.Errorf("failed to set new file name: %v", err)
	}

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "NtRename")
	if err != nil {
		return err
	}
	if response.Header.Status != 0x00000000 {
		return fmt.Errorf("NtRename failed: 0x%08x", response.Header.Status)
	}
	return nil
}

// NtRename performs an in-place rename of oldPath to newPath using
// SMB_COM_NT_RENAME (SMB_NT_RENAME_RENAME_FILE). Unlike the legacy RenameFile,
// wildcards are not supported.
func (c *Client) NtRename(oldPath, newPath string) error {
	return c.ntRename(oldPath, newPath, infoNtRenameRenameFile)
}

// CreateHardLink creates a hard link at linkPath that refers to the existing file
// existingPath, using SMB_COM_NT_RENAME (SMB_NT_RENAME_SET_LINK_INFO).
func (c *Client) CreateHardLink(existingPath, linkPath string) error {
	return c.ntRename(existingPath, linkPath, infoNtRenameSetLinkInfo)
}
