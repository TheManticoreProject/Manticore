package ms_fscc

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
