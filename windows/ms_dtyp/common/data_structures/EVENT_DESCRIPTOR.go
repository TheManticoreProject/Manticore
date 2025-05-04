package data_structures

import "github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_types"

// The EVENT_DESCRIPTOR structure specifies the metadata that defines an event.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/a6110d36-28c1-4290-b79e-26aa95a0b1a0
type EVENT_DESCRIPTOR struct {
	Id      data_types.USHORT
	Version data_types.UCHAR
	Channel data_types.UCHAR
	Level   data_types.UCHAR
	Opcode  data_types.UCHAR
	Task    data_types.USHORT
	Keyword data_types.ULONGLONG
}
