package client

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/ms_fscc"
)

// DeleteFile removes a file on the current tree. It opens the file with DELETE
// access and FILE_DELETE_ON_CLOSE, so the file is removed when the handle closes.
func (c *Client) DeleteFile(path string) error {
	fileId, err := c.CreateFile(path,
		fileflags.DELETE,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE|fileflags.FILE_SHARE_DELETE,
		fileflags.FILE_OPEN,
		fileflags.FILE_NON_DIRECTORY_FILE|fileflags.FILE_DELETE_ON_CLOSE)
	if err != nil {
		return err
	}
	return c.CloseFile(fileId)
}

// CreateDirectory creates a directory on the current tree.
func (c *Client) CreateDirectory(path string) error {
	fileId, err := c.CreateFile(path,
		fileflags.FILE_LIST_DIRECTORY|fileflags.FILE_READ_ATTRIBUTES,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
		fileflags.FILE_CREATE,
		fileflags.FILE_DIRECTORY_FILE)
	if err != nil {
		return err
	}
	return c.CloseFile(fileId)
}

// DeleteDirectory removes an (empty) directory on the current tree.
func (c *Client) DeleteDirectory(path string) error {
	fileId, err := c.CreateFile(path,
		fileflags.DELETE,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE|fileflags.FILE_SHARE_DELETE,
		fileflags.FILE_OPEN,
		fileflags.FILE_DIRECTORY_FILE|fileflags.FILE_DELETE_ON_CLOSE)
	if err != nil {
		return err
	}
	return c.CloseFile(fileId)
}

// RenameFile renames (or moves) oldPath to newPath on the current tree. Both are
// share-relative. When replaceIfExists is false, the rename fails if newPath
// already exists.
func (c *Client) RenameFile(oldPath, newPath string, replaceIfExists bool) error {
	fileId, err := c.CreateFile(oldPath,
		fileflags.DELETE|fileflags.FILE_READ_ATTRIBUTES,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE|fileflags.FILE_SHARE_DELETE,
		fileflags.FILE_OPEN,
		fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		return err
	}
	defer c.CloseFile(fileId)
	return c.RenameByHandle(fileId, newPath, replaceIfExists)
}

// GetVolumeSizeInfo returns the size information of the volume backing the
// current tree. It opens the share root and issues a filesystem QUERY_INFO.
func (c *Client) GetVolumeSizeInfo() (*ms_fscc.FileFsSizeInformation, error) {
	root, err := c.CreateFile("",
		fileflags.FILE_READ_ATTRIBUTES,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
		fileflags.FILE_OPEN,
		fileflags.FILE_DIRECTORY_FILE)
	if err != nil {
		return nil, err
	}
	defer c.CloseFile(root)
	return c.QueryFsSizeInfo(root)
}

// openForQuery opens path for attribute queries (FILE_READ_ATTRIBUTES) and returns
// its handle; the caller closes it.
func (c *Client) openForQuery(path string) (types.SMB2_FILEID, error) {
	return c.CreateFile(path,
		fileflags.FILE_READ_ATTRIBUTES,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE|fileflags.FILE_SHARE_DELETE,
		fileflags.FILE_OPEN,
		0)
}

// Stat opens path, queries its basic and standard information, and closes it.
func (c *Client) Stat(path string) (*FileStat, error) {
	fileId, err := c.openForQuery(path)
	if err != nil {
		return nil, err
	}
	defer c.CloseFile(fileId)

	basic, err := c.QueryFileBasicInfo(fileId)
	if err != nil {
		return nil, err
	}
	std, err := c.QueryFileStandardInfo(fileId)
	if err != nil {
		return nil, err
	}
	return &FileStat{
		Size:           uint64(std.EndOfFile),
		AllocationSize: uint64(std.AllocationSize),
		FileAttributes: basic.FileAttributes,
		IsDirectory:    std.Directory,
		CreationTime:   basic.CreationTime,
		LastAccessTime: basic.LastAccessTime,
		LastWriteTime:  basic.LastWriteTime,
		ChangeTime:     basic.ChangeTime,
	}, nil
}

// FileStat is a convenience summary of a file's metadata, combining
// FILE_BASIC_INFORMATION and FILE_STANDARD_INFORMATION. FILETIME fields are raw
// 100ns ticks since 1601-01-01 UTC.
type FileStat struct {
	Size           uint64
	AllocationSize uint64
	FileAttributes uint32
	IsDirectory    bool
	CreationTime   uint64
	LastAccessTime uint64
	LastWriteTime  uint64
	ChangeTime     uint64
}
