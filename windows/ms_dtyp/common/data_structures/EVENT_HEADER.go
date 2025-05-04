package data_structures

import (
	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_types"
)

// The EVENT_HEADER structure defines the main parameters of an event.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/fa4f7836-06ee-4ab6-8688-386a5a85f8c5
type EVENT_HEADER struct {
	Size            data_types.USHORT
	HeaderType      data_types.USHORT
	Flags           data_types.USHORT
	EventProperty   data_types.USHORT
	ThreadId        data_types.ULONG
	ProcessId       data_types.ULONG
	TimeStamp       LARGE_INTEGER
	ProviderId      GUID
	EventDescriptor EVENT_DESCRIPTOR
	KernelTime      data_types.ULONG
	UserTime        data_types.ULONG
	ProcessorTime   data_types.ULONG64
	ActivityId      GUID
}

type PEVENT_HEADER *EVENT_HEADER
