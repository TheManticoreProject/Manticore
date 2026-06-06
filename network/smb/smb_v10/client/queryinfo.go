package client

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/informationlevels"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
)

// TRANS2 information level codes (MS-CIFS 2.2.2.3). QUERY file levels are used with
// QueryPathInformation/QueryFileInformation; SET file levels with
// SetPathInformation/SetFileInformation; FS levels with QueryFsInformation.
const (
	InfoLevelQueryFileBasic       uint16 = 0x0101
	InfoLevelQueryFileStandard    uint16 = 0x0102
	InfoLevelQueryFileEA          uint16 = 0x0103
	InfoLevelQueryFileName        uint16 = 0x0104
	InfoLevelQueryFileAll         uint16 = 0x0107
	InfoLevelQueryFileAltName     uint16 = 0x0108
	InfoLevelQueryFileStream      uint16 = 0x0109
	InfoLevelQueryFileCompression uint16 = 0x010B

	InfoLevelSetFileBasic       uint16 = 0x0101
	InfoLevelSetFileDisposition uint16 = 0x0102
	InfoLevelSetFileAllocation  uint16 = 0x0103
	InfoLevelSetFileEndOfFile   uint16 = 0x0104

	InfoLevelQueryFsVolume    uint16 = 0x0102
	InfoLevelQueryFsSize      uint16 = 0x0103
	InfoLevelQueryFsDevice    uint16 = 0x0104
	InfoLevelQueryFsAttribute uint16 = 0x0105
)

// normalizeSMBPath returns path with a single leading backslash, as expected for a
// share-relative SMB path. An empty path becomes "\".
func normalizeSMBPath(path string) string {
	if len(path) == 0 {
		return "\\"
	}
	if path[0] != '\\' {
		return "\\" + path
	}
	return path
}

// QueryPathInformation issues TRANS2_QUERY_PATH_INFORMATION for a share-relative
// path at the given information level and returns the raw response data bytes
// (the information-level structure).
func (c *Client) QueryPathInformation(path string, infoLevel uint16) ([]byte, error) {
	if c.Session == nil {
		return nil, fmt.Errorf("no session established")
	}

	// Parameters: InformationLevel(2) + Reserved(4) + FileName (OEM, null-terminated).
	params := []byte{}
	params = binary.LittleEndian.AppendUint16(params, infoLevel)
	params = binary.LittleEndian.AppendUint32(params, 0) // Reserved
	params = append(params, []byte(normalizeSMBPath(path))...)
	params = append(params, 0x00)

	_, data, err := c.trans2(uint16(subcommands.TRANS2_QUERY_PATH_INFORMATION), params, nil)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// QueryFileInformation issues TRANS2_QUERY_FILE_INFORMATION for an open FID at the
// given information level and returns the raw response data bytes.
func (c *Client) QueryFileInformation(fid FID, infoLevel uint16) ([]byte, error) {
	if c.Session == nil {
		return nil, fmt.Errorf("no session established")
	}

	// Parameters: FID(2) + InformationLevel(2).
	params := []byte{}
	params = binary.LittleEndian.AppendUint16(params, uint16(fid))
	params = binary.LittleEndian.AppendUint16(params, infoLevel)

	_, data, err := c.trans2(uint16(subcommands.TRANS2_QUERY_FILE_INFORMATION), params, nil)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// QueryFsInformation issues TRANS2_QUERY_FS_INFORMATION at the given information
// level and returns the raw response data bytes.
func (c *Client) QueryFsInformation(infoLevel uint16) ([]byte, error) {
	if c.Session == nil {
		return nil, fmt.Errorf("no session established")
	}

	// Parameters: InformationLevel(2).
	params := binary.LittleEndian.AppendUint16([]byte{}, infoLevel)

	_, data, err := c.trans2(uint16(subcommands.TRANS2_QUERY_FS_INFORMATION), params, nil)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SetFileInformation issues TRANS2_SET_FILE_INFORMATION for an open FID at the
// given information level, sending data as the information-level payload.
func (c *Client) SetFileInformation(fid FID, infoLevel uint16, data []byte) error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	// Parameters: FID(2) + InformationLevel(2) + Reserved(2).
	params := []byte{}
	params = binary.LittleEndian.AppendUint16(params, uint16(fid))
	params = binary.LittleEndian.AppendUint16(params, infoLevel)
	params = binary.LittleEndian.AppendUint16(params, 0) // Reserved

	_, _, err := c.trans2(uint16(subcommands.TRANS2_SET_FILE_INFORMATION), params, data)
	return err
}

// SetPathInformation issues TRANS2_SET_PATH_INFORMATION for a share-relative path
// at the given information level, sending data as the information-level payload.
func (c *Client) SetPathInformation(path string, infoLevel uint16, data []byte) error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	// Parameters: InformationLevel(2) + Reserved(4) + FileName (OEM, null-terminated).
	params := []byte{}
	params = binary.LittleEndian.AppendUint16(params, infoLevel)
	params = binary.LittleEndian.AppendUint32(params, 0) // Reserved
	params = append(params, []byte(normalizeSMBPath(path))...)
	params = append(params, 0x00)

	_, _, err := c.trans2(uint16(subcommands.TRANS2_SET_PATH_INFORMATION), params, data)
	return err
}

// --- Typed convenience wrappers -------------------------------------------------

// GetFileBasicInfo queries the 64-bit timestamps and extended attributes of a
// share-relative path.
func (c *Client) GetFileBasicInfo(path string) (*informationlevels.SMB_QUERY_FILE_BASIC_INFO, error) {
	data, err := c.QueryPathInformation(path, InfoLevelQueryFileBasic)
	if err != nil {
		return nil, err
	}
	info := &informationlevels.SMB_QUERY_FILE_BASIC_INFO{}
	if _, err := info.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("failed to decode SMB_QUERY_FILE_BASIC_INFO: %v", err)
	}
	return info, nil
}

// GetFileStandardInfo queries the size, link count, delete-pending and directory
// flags of a share-relative path.
func (c *Client) GetFileStandardInfo(path string) (*informationlevels.SMB_QUERY_FILE_STANDARD_INFO, error) {
	data, err := c.QueryPathInformation(path, InfoLevelQueryFileStandard)
	if err != nil {
		return nil, err
	}
	info := &informationlevels.SMB_QUERY_FILE_STANDARD_INFO{}
	if _, err := info.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("failed to decode SMB_QUERY_FILE_STANDARD_INFO: %v", err)
	}
	return info, nil
}

// SetFileEndOfFile sets the end-of-file position (logical size) of an open file.
func (c *Client) SetFileEndOfFile(fid FID, endOfFile uint64) error {
	info := &informationlevels.SMB_SET_FILE_END_OF_FILE_INFO{}
	info.Endoffile.QuadPart = endOfFile
	data, err := info.Marshal()
	if err != nil {
		return err
	}
	return c.SetFileInformation(fid, InfoLevelSetFileEndOfFile, data)
}

// SetFileDeleteOnClose marks (or unmarks) an open file for deletion when its last
// handle is closed.
func (c *Client) SetFileDeleteOnClose(fid FID, deletePending bool) error {
	info := &informationlevels.SMB_SET_FILE_DISPOSITION_INFO{}
	if deletePending {
		info.Deletepending = 0x01
	}
	data, err := info.Marshal()
	if err != nil {
		return err
	}
	return c.SetFileInformation(fid, InfoLevelSetFileDisposition, data)
}

// GetFsSizeInfo queries the volume's total/free allocation units and sector sizes.
func (c *Client) GetFsSizeInfo() (*informationlevels.SMB_QUERY_FS_SIZE_INFO, error) {
	data, err := c.QueryFsInformation(InfoLevelQueryFsSize)
	if err != nil {
		return nil, err
	}
	info := &informationlevels.SMB_QUERY_FS_SIZE_INFO{}
	if _, err := info.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("failed to decode SMB_QUERY_FS_SIZE_INFO: %v", err)
	}
	return info, nil
}

// GetFsAttributeInfo queries the filesystem attribute flags and name.
func (c *Client) GetFsAttributeInfo() (*informationlevels.SMB_QUERY_FS_ATTRIBUTE_INFO, error) {
	data, err := c.QueryFsInformation(InfoLevelQueryFsAttribute)
	if err != nil {
		return nil, err
	}
	info := &informationlevels.SMB_QUERY_FS_ATTRIBUTE_INFO{}
	if _, err := info.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("failed to decode SMB_QUERY_FS_ATTRIBUTE_INFO: %v", err)
	}
	return info, nil
}
