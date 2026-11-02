package filesystem

import (
	"encoding/binary"
	"fmt"
)

// FileFsSizeInformation is FILE_FS_SIZE_INFORMATION (FsInformationClass 3): the
// total/available allocation units and the sector geometry of a volume. Fixed
// 24-byte wire layout.
//
// [MS-FSCC] 2.5.8 FileFsSizeInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/e13e068c-e3a7-4dd4-bff4-87d57c30ac0c
type FileFsSizeInformation struct {
	TotalAllocationUnits     int64
	AvailableAllocationUnits int64
	SectorsPerAllocationUnit uint32
	BytesPerSector           uint32
}

// FileFsSizeInformationSize is the fixed wire size of FILE_FS_SIZE_INFORMATION.
const FileFsSizeInformationSize = 24

// Marshal encodes the structure to its 24-byte wire form.
func (fi *FileFsSizeInformation) Marshal() ([]byte, error) {
	b := make([]byte, FileFsSizeInformationSize)
	binary.LittleEndian.PutUint64(b[0:8], uint64(fi.TotalAllocationUnits))
	binary.LittleEndian.PutUint64(b[8:16], uint64(fi.AvailableAllocationUnits))
	binary.LittleEndian.PutUint32(b[16:20], fi.SectorsPerAllocationUnit)
	binary.LittleEndian.PutUint32(b[20:24], fi.BytesPerSector)
	return b, nil
}

// Unmarshal decodes the structure from its wire form.
func (fi *FileFsSizeInformation) Unmarshal(data []byte) error {
	if len(data) < FileFsSizeInformationSize {
		return fmt.Errorf("FILE_FS_SIZE_INFORMATION: need %d bytes, got %d", FileFsSizeInformationSize, len(data))
	}
	fi.TotalAllocationUnits = int64(binary.LittleEndian.Uint64(data[0:8]))
	fi.AvailableAllocationUnits = int64(binary.LittleEndian.Uint64(data[8:16]))
	fi.SectorsPerAllocationUnit = binary.LittleEndian.Uint32(data[16:20])
	fi.BytesPerSector = binary.LittleEndian.Uint32(data[20:24])
	return nil
}

// TotalBytes returns the total capacity of the volume in bytes.
func (fi *FileFsSizeInformation) TotalBytes() uint64 {
	return uint64(fi.TotalAllocationUnits) * uint64(fi.SectorsPerAllocationUnit) * uint64(fi.BytesPerSector)
}

// AvailableBytes returns the free capacity of the volume in bytes.
func (fi *FileFsSizeInformation) AvailableBytes() uint64 {
	return uint64(fi.AvailableAllocationUnits) * uint64(fi.SectorsPerAllocationUnit) * uint64(fi.BytesPerSector)
}

// FileFsVolumeInformation is FILE_FS_VOLUME_INFORMATION (FsInformationClass 1):
// the volume creation time, serial number, label, and object-id support flag.
// VolumeCreationTime is a FILETIME (100ns ticks since 1601-01-01 UTC). The fixed
// 18-byte prefix is followed by the UTF-16LE VolumeLabel.
//
// [MS-FSCC] 2.5.9 FileFsVolumeInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/bf691378-c34e-4a13-976e-404ea1a87738
type FileFsVolumeInformation struct {
	VolumeCreationTime uint64
	VolumeSerialNumber uint32
	SupportsObjects    bool
	VolumeLabel        string
}

// fileFsVolumeInfoFixedSize is the fixed portion preceding VolumeLabel:
// VolumeCreationTime(8) VolumeSerialNumber(4) VolumeLabelLength(4)
// SupportsObjects(1) Reserved(1).
const fileFsVolumeInfoFixedSize = 18

// Marshal encodes the structure to its wire form (fixed prefix + VolumeLabel).
func (fi *FileFsVolumeInformation) Marshal() ([]byte, error) {
	label := encodeUTF16LE(fi.VolumeLabel)
	b := make([]byte, fileFsVolumeInfoFixedSize+len(label))
	binary.LittleEndian.PutUint64(b[0:8], fi.VolumeCreationTime)
	binary.LittleEndian.PutUint32(b[8:12], fi.VolumeSerialNumber)
	binary.LittleEndian.PutUint32(b[12:16], uint32(len(label)))
	if fi.SupportsObjects {
		b[16] = 1
	}
	// b[17] Reserved = 0
	copy(b[fileFsVolumeInfoFixedSize:], label)
	return b, nil
}

// Unmarshal decodes the structure from its wire form.
func (fi *FileFsVolumeInformation) Unmarshal(data []byte) error {
	if len(data) < fileFsVolumeInfoFixedSize {
		return fmt.Errorf("FILE_FS_VOLUME_INFORMATION: need %d bytes, got %d", fileFsVolumeInfoFixedSize, len(data))
	}
	fi.VolumeCreationTime = binary.LittleEndian.Uint64(data[0:8])
	fi.VolumeSerialNumber = binary.LittleEndian.Uint32(data[8:12])
	labelLen := int(binary.LittleEndian.Uint32(data[12:16]))
	fi.SupportsObjects = data[16] != 0
	if labelLen > 0 && labelLen <= len(data)-fileFsVolumeInfoFixedSize {
		fi.VolumeLabel = decodeUTF16LE(data[fileFsVolumeInfoFixedSize : fileFsVolumeInfoFixedSize+labelLen])
	}
	return nil
}

// FileFsFullSizeInformation is FILE_FS_FULL_SIZE_INFORMATION (FsInformationClass
// 7): like FileFsSizeInformation but distinguishing the caller's available
// allocation units (quota-limited) from the volume's actual free units. Fixed
// 32-byte wire layout.
//
// [MS-FSCC] 2.5.4 FileFsFullSizeInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/63768db7-9012-4209-8cca-00781e7322f5
type FileFsFullSizeInformation struct {
	TotalAllocationUnits           int64
	CallerAvailableAllocationUnits int64
	ActualAvailableAllocationUnits int64
	SectorsPerAllocationUnit       uint32
	BytesPerSector                 uint32
}

// FileFsFullSizeInformationSize is the fixed wire size of
// FILE_FS_FULL_SIZE_INFORMATION.
const FileFsFullSizeInformationSize = 32

// Marshal encodes the structure to its 32-byte wire form.
func (fi *FileFsFullSizeInformation) Marshal() ([]byte, error) {
	b := make([]byte, FileFsFullSizeInformationSize)
	binary.LittleEndian.PutUint64(b[0:8], uint64(fi.TotalAllocationUnits))
	binary.LittleEndian.PutUint64(b[8:16], uint64(fi.CallerAvailableAllocationUnits))
	binary.LittleEndian.PutUint64(b[16:24], uint64(fi.ActualAvailableAllocationUnits))
	binary.LittleEndian.PutUint32(b[24:28], fi.SectorsPerAllocationUnit)
	binary.LittleEndian.PutUint32(b[28:32], fi.BytesPerSector)
	return b, nil
}

// Unmarshal decodes the structure from its wire form.
func (fi *FileFsFullSizeInformation) Unmarshal(data []byte) error {
	if len(data) < FileFsFullSizeInformationSize {
		return fmt.Errorf("FILE_FS_FULL_SIZE_INFORMATION: need %d bytes, got %d", FileFsFullSizeInformationSize, len(data))
	}
	fi.TotalAllocationUnits = int64(binary.LittleEndian.Uint64(data[0:8]))
	fi.CallerAvailableAllocationUnits = int64(binary.LittleEndian.Uint64(data[8:16]))
	fi.ActualAvailableAllocationUnits = int64(binary.LittleEndian.Uint64(data[16:24]))
	fi.SectorsPerAllocationUnit = binary.LittleEndian.Uint32(data[24:28])
	fi.BytesPerSector = binary.LittleEndian.Uint32(data[28:32])
	return nil
}

// FileFsAttributeInformation is FILE_FS_ATTRIBUTE_INFORMATION (FsInformationClass
// 5): the file system's attribute flags, maximum component name length, and file
// system name. The fixed 12-byte prefix is followed by the UTF-16LE
// FileSystemName.
//
// [MS-FSCC] 2.5.1 FileFsAttributeInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/ebc7e6e5-4650-4e54-b17c-cf60f6fbeeaa
type FileFsAttributeInformation struct {
	FileSystemAttributes       uint32
	MaximumComponentNameLength int32
	FileSystemName             string
}

// fileFsAttributeInfoFixedSize is the fixed portion preceding FileSystemName.
const fileFsAttributeInfoFixedSize = 12

// Marshal encodes the structure to its wire form (fixed prefix + FileSystemName).
func (fi *FileFsAttributeInformation) Marshal() ([]byte, error) {
	name := encodeUTF16LE(fi.FileSystemName)
	b := make([]byte, fileFsAttributeInfoFixedSize+len(name))
	binary.LittleEndian.PutUint32(b[0:4], fi.FileSystemAttributes)
	binary.LittleEndian.PutUint32(b[4:8], uint32(fi.MaximumComponentNameLength))
	binary.LittleEndian.PutUint32(b[8:12], uint32(len(name)))
	copy(b[fileFsAttributeInfoFixedSize:], name)
	return b, nil
}

// Unmarshal decodes the structure from its wire form.
func (fi *FileFsAttributeInformation) Unmarshal(data []byte) error {
	if len(data) < fileFsAttributeInfoFixedSize {
		return fmt.Errorf("FILE_FS_ATTRIBUTE_INFORMATION: need %d bytes, got %d", fileFsAttributeInfoFixedSize, len(data))
	}
	fi.FileSystemAttributes = binary.LittleEndian.Uint32(data[0:4])
	fi.MaximumComponentNameLength = int32(binary.LittleEndian.Uint32(data[4:8]))
	nameLen := int(binary.LittleEndian.Uint32(data[8:12]))
	if nameLen > 0 && nameLen <= len(data)-fileFsAttributeInfoFixedSize {
		fi.FileSystemName = decodeUTF16LE(data[fileFsAttributeInfoFixedSize : fileFsAttributeInfoFixedSize+nameLen])
	}
	return nil
}

// FileFsDeviceInformation is FILE_FS_DEVICE_INFORMATION (FsInformationClass 4):
// the underlying device type and its characteristics. Fixed 8-byte wire layout.
//
// [MS-FSCC] 2.5.10 FileFsDeviceInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/616b66d5-b335-4e1c-8f87-b4a55e8d3e4a
type FileFsDeviceInformation struct {
	DeviceType      uint32
	Characteristics uint32
}

// FileFsDeviceInformationSize is the fixed wire size of FILE_FS_DEVICE_INFORMATION.
const FileFsDeviceInformationSize = 8

// Marshal encodes the structure to its 8-byte wire form.
func (fi *FileFsDeviceInformation) Marshal() ([]byte, error) {
	b := make([]byte, FileFsDeviceInformationSize)
	binary.LittleEndian.PutUint32(b[0:4], fi.DeviceType)
	binary.LittleEndian.PutUint32(b[4:8], fi.Characteristics)
	return b, nil
}

// Unmarshal decodes the structure from its wire form.
func (fi *FileFsDeviceInformation) Unmarshal(data []byte) error {
	if len(data) < FileFsDeviceInformationSize {
		return fmt.Errorf("FILE_FS_DEVICE_INFORMATION: need %d bytes, got %d", FileFsDeviceInformationSize, len(data))
	}
	fi.DeviceType = binary.LittleEndian.Uint32(data[0:4])
	fi.Characteristics = binary.LittleEndian.Uint32(data[4:8])
	return nil
}

// FILE_FS_SECTOR_SIZE_INFORMATION Flags values.
const (
	SSINFO_FLAGS_ALIGNED_DEVICE              = 0x00000001
	SSINFO_FLAGS_PARTITION_ALIGNED_ON_DEVICE = 0x00000002
	SSINFO_FLAGS_NO_SEEK_PENALTY             = 0x00000004
	SSINFO_FLAGS_TRIM_ENABLED                = 0x00000008
)

// SSINFO_OFFSET_UNKNOWN in ByteOffsetForSectorAlignment or
// ByteOffsetForPartitionAlignment means the offset could not be computed.
const SSINFO_OFFSET_UNKNOWN = 0xFFFFFFFF

// FileFsSectorSizeInformation is FILE_FS_SECTOR_SIZE_INFORMATION
// (FsInformationClass 11): the logical and physical sector sizes of the device
// backing a volume, and its alignment. Fixed 28-byte wire layout.
//
// [MS-FSCC] FileFsSectorSizeInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/3e75d97f-1d0b-4e47-b435-73c513837a57
type FileFsSectorSizeInformation struct {
	LogicalBytesPerSector                                 uint32
	PhysicalBytesPerSectorForAtomicity                    uint32
	PhysicalBytesPerSectorForPerformance                  uint32
	FileSystemEffectivePhysicalBytesPerSectorForAtomicity uint32
	Flags                                                 uint32
	ByteOffsetForSectorAlignment                          uint32
	ByteOffsetForPartitionAlignment                       uint32
}

// FileFsSectorSizeInformationSize is the fixed wire size of
// FILE_FS_SECTOR_SIZE_INFORMATION.
const FileFsSectorSizeInformationSize = 28

// Marshal encodes the structure to its 28-byte wire form.
func (fi *FileFsSectorSizeInformation) Marshal() ([]byte, error) {
	b := make([]byte, FileFsSectorSizeInformationSize)
	binary.LittleEndian.PutUint32(b[0:4], fi.LogicalBytesPerSector)
	binary.LittleEndian.PutUint32(b[4:8], fi.PhysicalBytesPerSectorForAtomicity)
	binary.LittleEndian.PutUint32(b[8:12], fi.PhysicalBytesPerSectorForPerformance)
	binary.LittleEndian.PutUint32(b[12:16], fi.FileSystemEffectivePhysicalBytesPerSectorForAtomicity)
	binary.LittleEndian.PutUint32(b[16:20], fi.Flags)
	binary.LittleEndian.PutUint32(b[20:24], fi.ByteOffsetForSectorAlignment)
	binary.LittleEndian.PutUint32(b[24:28], fi.ByteOffsetForPartitionAlignment)
	return b, nil
}

// Unmarshal decodes the structure from its wire form.
func (fi *FileFsSectorSizeInformation) Unmarshal(data []byte) error {
	if len(data) < FileFsSectorSizeInformationSize {
		return fmt.Errorf("FILE_FS_SECTOR_SIZE_INFORMATION: need %d bytes, got %d", FileFsSectorSizeInformationSize, len(data))
	}
	fi.LogicalBytesPerSector = binary.LittleEndian.Uint32(data[0:4])
	fi.PhysicalBytesPerSectorForAtomicity = binary.LittleEndian.Uint32(data[4:8])
	fi.PhysicalBytesPerSectorForPerformance = binary.LittleEndian.Uint32(data[8:12])
	fi.FileSystemEffectivePhysicalBytesPerSectorForAtomicity = binary.LittleEndian.Uint32(data[12:16])
	fi.Flags = binary.LittleEndian.Uint32(data[16:20])
	fi.ByteOffsetForSectorAlignment = binary.LittleEndian.Uint32(data[20:24])
	fi.ByteOffsetForPartitionAlignment = binary.LittleEndian.Uint32(data[24:28])
	return nil
}

// FILE_FS_CONTROL_INFORMATION FileSystemControlFlags values.
const (
	FILE_VC_QUOTA_TRACK            = 0x00000001
	FILE_VC_QUOTA_ENFORCE          = 0x00000002
	FILE_VC_CONTENT_INDEX_DISABLED = 0x00000008
	FILE_VC_LOG_QUOTA_THRESHOLD    = 0x00000010
	FILE_VC_LOG_QUOTA_LIMIT        = 0x00000020
	FILE_VC_LOG_VOLUME_THRESHOLD   = 0x00000040
	FILE_VC_LOG_VOLUME_LIMIT       = 0x00000080
	FILE_VC_QUOTAS_INCOMPLETE      = 0x00000100
	FILE_VC_QUOTAS_REBUILDING      = 0x00000200
)

// QuotaNoLimit in DefaultQuotaThreshold or DefaultQuotaLimit means no default
// per-user threshold or limit is set.
const QuotaNoLimit = 0xFFFFFFFFFFFFFFFF

// FileFsControlInformation is FILE_FS_CONTROL_INFORMATION (FsInformationClass
// 6): the quota and content-indexing control settings of a volume, used for
// both query and set. Fixed 48-byte wire layout, the last 4 bytes padding.
//
// [MS-FSCC] FileFsControlInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/e5a70738-7ee4-46d9-a5f7-6644daa49a51
type FileFsControlInformation struct {
	FreeSpaceStartFiltering int64
	FreeSpaceThreshold      int64
	FreeSpaceStopFiltering  int64
	DefaultQuotaThreshold   uint64
	DefaultQuotaLimit       uint64
	FileSystemControlFlags  uint32
}

// FileFsControlInformationSize is the fixed wire size of
// FILE_FS_CONTROL_INFORMATION, including its 4 bytes of padding.
const FileFsControlInformationSize = 48

// Marshal encodes the structure to its 48-byte wire form, the padding zero.
func (fi *FileFsControlInformation) Marshal() ([]byte, error) {
	b := make([]byte, FileFsControlInformationSize)
	binary.LittleEndian.PutUint64(b[0:8], uint64(fi.FreeSpaceStartFiltering))
	binary.LittleEndian.PutUint64(b[8:16], uint64(fi.FreeSpaceThreshold))
	binary.LittleEndian.PutUint64(b[16:24], uint64(fi.FreeSpaceStopFiltering))
	binary.LittleEndian.PutUint64(b[24:32], fi.DefaultQuotaThreshold)
	binary.LittleEndian.PutUint64(b[32:40], fi.DefaultQuotaLimit)
	binary.LittleEndian.PutUint32(b[40:44], fi.FileSystemControlFlags)
	return b, nil
}

// Unmarshal decodes the structure from its wire form. The trailing padding is
// ignored, and a 44-byte buffer without it is accepted.
func (fi *FileFsControlInformation) Unmarshal(data []byte) error {
	const withoutPadding = FileFsControlInformationSize - 4
	if len(data) < withoutPadding {
		return fmt.Errorf("FILE_FS_CONTROL_INFORMATION: need %d bytes, got %d", withoutPadding, len(data))
	}
	fi.FreeSpaceStartFiltering = int64(binary.LittleEndian.Uint64(data[0:8]))
	fi.FreeSpaceThreshold = int64(binary.LittleEndian.Uint64(data[8:16]))
	fi.FreeSpaceStopFiltering = int64(binary.LittleEndian.Uint64(data[16:24]))
	fi.DefaultQuotaThreshold = binary.LittleEndian.Uint64(data[24:32])
	fi.DefaultQuotaLimit = binary.LittleEndian.Uint64(data[32:40])
	fi.FileSystemControlFlags = binary.LittleEndian.Uint32(data[40:44])
	return nil
}
