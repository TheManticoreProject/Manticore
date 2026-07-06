package msdtyp

// The EVENT_DESCRIPTOR structure specifies the metadata that defines an event.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/a6110d36-28c1-4290-b79e-26aa95a0b1a0
type EVENT_DESCRIPTOR struct {
	Id      USHORT
	Version UCHAR
	Channel UCHAR
	Level   UCHAR
	Opcode  UCHAR
	Task    USHORT
	Keyword ULONGLONG
}
