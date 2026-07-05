package mststs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TS_SYS_PROCESS_INFORMATION_NT6 is the Windows Vista+ ("NT6") form of
// TS_SYS_PROCESS_INFORMATION returned by RpcWinStationGetAllProcesses_NT6 ([MS-TSTS]
// 2.2.2.7.5, allproc.h _TS_SYS_PROCESS_INFORMATION_NT6). It differs from the legacy form
// only in the ImageName string type. See TS_SYS_PROCESS_INFORMATION for the SIZE_T caveat.
type TS_SYS_PROCESS_INFORMATION_NT6 struct {
	NextEntryOffset              ndr.DWORD
	NumberOfThreads              ndr.DWORD
	SpareLi1                     dtyp.LARGE_INTEGER
	SpareLi2                     dtyp.LARGE_INTEGER
	SpareLi3                     dtyp.LARGE_INTEGER
	CreateTime                   dtyp.LARGE_INTEGER
	UserTime                     dtyp.LARGE_INTEGER
	KernelTime                   dtyp.LARGE_INTEGER
	ImageName                    NT6_TS_UNICODE_STRING
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
