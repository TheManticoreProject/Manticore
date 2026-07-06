package mststs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// TS_SYS_PROCESS_INFORMATION describes a single process returned by
// RpcWinStationGetAllProcesses ([MS-TSTS] 2.2.2.7.2, allproc.h _TS_SYS_PROCESS_INFORMATION).
//
// The SIZE_T fields are modeled as uint64: their on-the-wire width follows the server's
// pointer size, and modern terminal servers are 64-bit. This field width is UNVERIFIED
// against a live 32-bit server.
type TS_SYS_PROCESS_INFORMATION struct {
	NextEntryOffset              ndr.DWORD
	NumberOfThreads              ndr.DWORD
	SpareLi1                     msdtyp.LARGE_INTEGER
	SpareLi2                     msdtyp.LARGE_INTEGER
	SpareLi3                     msdtyp.LARGE_INTEGER
	CreateTime                   msdtyp.LARGE_INTEGER
	UserTime                     msdtyp.LARGE_INTEGER
	KernelTime                   msdtyp.LARGE_INTEGER
	ImageName                    TS_UNICODE_STRING
	BasePriority                 int32
	UniqueProcessId              ndr.DWORD
	InheritedFromUniqueProcessId ndr.DWORD
	HandleCount                  ndr.DWORD
	SessionId                    ndr.DWORD
	SpareUl3                     ndr.DWORD
	PeakVirtualSize              uint64
	VirtualSize                  uint64
	PageFaultCount               ndr.DWORD
	PeakWorkingSetSize           ndr.DWORD
	WorkingSetSize               ndr.DWORD
	QuotaPeakPagedPoolUsage      uint64
	QuotaPagedPoolUsage          uint64
	QuotaPeakNonPagedPoolUsage   uint64
	QuotaNonPagedPoolUsage       uint64
	PagefileUsage                uint64
	PeakPagefileUsage            uint64
	PrivatePageCount             uint64
}
