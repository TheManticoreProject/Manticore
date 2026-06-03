package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// DeleteFile deletes one or more files matching pattern on the current tree.
// pattern uses backslash separators relative to the share root (e.g. "\subdir\file.txt").
// Wildcards are permitted in the filename component.
//
// Wire: SMB_COM_DELETE request / response.
func (c *Client) DeleteFile(pattern string) error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	msg := c.newFileIOMessage(codes.SMB_COM_DELETE)

	cmd := commands.NewDeleteRequest()
	cmd.SearchAttributes = types.SMB_FILE_ATTRIBUTES{}
	if err := cmd.FileName.SetString(pattern); err != nil {
		return fmt.Errorf("failed to set file name: %v", err)
	}

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "Delete")
	if err != nil {
		return err
	}

	if response.Header.Status != 0x00000000 {
		return fmt.Errorf("Delete failed: 0x%08x", response.Header.Status)
	}

	return nil
}

// RenameFile renames (or moves) a file on the current tree.
// oldPath and newPath use backslash separators relative to the share root.
//
// Wire: SMB_COM_RENAME request / response.
func (c *Client) RenameFile(oldPath, newPath string) error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	msg := c.newFileIOMessage(codes.SMB_COM_RENAME)

	cmd := commands.NewRenameRequest()
	cmd.SearchAttributes = types.SMB_FILE_ATTRIBUTES{}
	if err := cmd.OldFileName.SetString(oldPath); err != nil {
		return fmt.Errorf("failed to set old file name: %v", err)
	}
	if err := cmd.NewFileName.SetString(newPath); err != nil {
		return fmt.Errorf("failed to set new file name: %v", err)
	}

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "Rename")
	if err != nil {
		return err
	}

	if response.Header.Status != 0x00000000 {
		return fmt.Errorf("Rename failed: 0x%08x", response.Header.Status)
	}

	return nil
}

// CreateDirectory creates a directory on the current tree.
// path uses backslash separators relative to the share root (e.g. "\newdir").
//
// Wire: SMB_COM_CREATE_DIRECTORY request / response.
func (c *Client) CreateDirectory(path string) error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	msg := c.newFileIOMessage(codes.SMB_COM_CREATE_DIRECTORY)

	cmd := commands.NewCreateDirectoryRequest()
	if err := cmd.DirectoryName.SetString(path); err != nil {
		return fmt.Errorf("failed to set directory name: %v", err)
	}

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "CreateDirectory")
	if err != nil {
		return err
	}

	if response.Header.Status != 0x00000000 {
		return fmt.Errorf("CreateDirectory failed: 0x%08x", response.Header.Status)
	}

	return nil
}

// DeleteDirectory removes an empty directory on the current tree.
// path uses backslash separators relative to the share root (e.g. "\olddir").
//
// Wire: SMB_COM_DELETE_DIRECTORY request / response.
func (c *Client) DeleteDirectory(path string) error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	msg := c.newFileIOMessage(codes.SMB_COM_DELETE_DIRECTORY)

	cmd := commands.NewDeleteDirectoryRequest()
	if err := cmd.DirectoryName.SetString(path); err != nil {
		return fmt.Errorf("failed to set directory name: %v", err)
	}

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "DeleteDirectory")
	if err != nil {
		return err
	}

	if response.Header.Status != 0x00000000 {
		return fmt.Errorf("DeleteDirectory failed: 0x%08x", response.Header.Status)
	}

	return nil
}

// CheckDirectory tests whether a path on the current tree is a valid directory.
// path uses backslash separators relative to the share root (e.g. "\subdir").
// Returns nil if the path is a directory, or an error otherwise.
//
// Wire: SMB_COM_CHECK_DIRECTORY request / response.
func (c *Client) CheckDirectory(path string) error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	msg := c.newFileIOMessage(codes.SMB_COM_CHECK_DIRECTORY)

	cmd := commands.NewCheckDirectoryRequest()
	if err := cmd.DirectoryName.SetString(path); err != nil {
		return fmt.Errorf("failed to set directory name: %v", err)
	}

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "CheckDirectory")
	if err != nil {
		return err
	}

	if response.Header.Status != 0x00000000 {
		return fmt.Errorf("CheckDirectory failed: 0x%08x", response.Header.Status)
	}

	return nil
}
