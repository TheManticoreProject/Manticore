package mststs

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// TS_ALL_PROCESSES_INFO_NT6 is the Windows Vista+ ("NT6") form of TS_ALL_PROCESSES_INFO
// ([MS-TSTS] 2.2.2.7.6, allproc.h _TS_ALL_PROCESSES_INFO_NT6).
type TS_ALL_PROCESSES_INFO_NT6 struct {
	PTsProcessInfo *TS_SYS_PROCESS_INFORMATION_NT6 `ndr:"unique"`
	SizeOfSid      ndr.DWORD
	PSid           []byte `ndr:"unique,size_is=SizeOfSid"`
}
