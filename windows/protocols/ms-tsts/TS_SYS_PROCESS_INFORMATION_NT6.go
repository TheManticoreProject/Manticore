package mststs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// TS_SYS_PROCESS_INFORMATION_NT6 is the Windows Vista+ ("NT6") form of
// TS_SYS_PROCESS_INFORMATION returned by RpcWinStationGetAllProcesses_NT6 ([MS-TSTS]
// 2.2.2.7.5, allproc.h _TS_SYS_PROCESS_INFORMATION_NT6). It differs from the legacy form
// only in the ImageName string type. See TS_SYS_PROCESS_INFORMATION for the SIZE_T caveat.
type TS_SYS_PROCESS_INFORMATION_NT6 struct {
	NextEntryOffset              ndr.DWORD
	NumberOfThreads              ndr.DWORD
	SpareLi1                     msdtyp.LARGE_INTEGER
	SpareLi2                     msdtyp.LARGE_INTEGER
	SpareLi3                     msdtyp.LARGE_INTEGER
	CreateTime                   msdtyp.LARGE_INTEGER
	UserTime                     msdtyp.LARGE_INTEGER
	KernelTime                   msdtyp.LARGE_INTEGER
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
