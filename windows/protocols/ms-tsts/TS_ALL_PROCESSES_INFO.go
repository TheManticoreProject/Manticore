package mststs

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// TS_ALL_PROCESSES_INFO pairs one process' information with the raw bytes of its owner's
// SID ([MS-TSTS] 2.2.2.7.3, allproc.h _TS_ALL_PROCESSES_INFO).
type TS_ALL_PROCESSES_INFO struct {
	PTsProcessInfo *TS_SYS_PROCESS_INFORMATION `ndr:"unique"`
	SizeOfSid      ndr.DWORD
	PSid           []byte `ndr:"unique,size_is=SizeOfSid"`
}
