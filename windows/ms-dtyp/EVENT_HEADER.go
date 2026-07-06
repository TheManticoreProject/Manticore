package msdtyp

import ()

// The EVENT_HEADER structure defines the main parameters of an event.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/fa4f7836-06ee-4ab6-8688-386a5a85f8c5
type EVENT_HEADER struct {
	Size            USHORT
	HeaderType      USHORT
	Flags           USHORT
	EventProperty   USHORT
	ThreadId        ULONG
	ProcessId       ULONG
	TimeStamp       LARGE_INTEGER
	ProviderId      GUID
	EventDescriptor EVENT_DESCRIPTOR
	KernelTime      ULONG
	UserTime        ULONG
	ProcessorTime   ULONG64
	ActivityId      GUID
}

type PEVENT_HEADER *EVENT_HEADER
