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
	if labelLen > 0 && fileFsVolumeInfoFixedSize+labelLen <= len(data) {
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
	if nameLen > 0 && fileFsAttributeInfoFixedSize+nameLen <= len(data) {
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
