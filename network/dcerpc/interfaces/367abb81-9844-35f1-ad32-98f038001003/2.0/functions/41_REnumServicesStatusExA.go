package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// rEnumServicesStatusExARequest carries the [in] parameters of REnumServicesStatusExA.
type rEnumServicesStatusExARequest struct {
	HSCManager     msscmr.SC_RPC_HANDLE
	InfoLevel      msscmr.SC_ENUM_TYPE
	DwServiceType  ndr.DWORD
	DwServiceState ndr.DWORD
	CbBufSize      ndr.DWORD
	LpResumeIndex  *ndr.DWORD `ndr:"unique"`
	PszGroupName   *ndr.STR   `ndr:"unique"`
}

func (*rEnumServicesStatusExARequest) Opnum() uint16 { return svcctl.OpnumREnumServicesStatusExA }

// rEnumServicesStatusExAResponse carries the [out] parameters and return value of REnumServicesStatusExA.
type rEnumServicesStatusExAResponse struct {
	LpBuffer           []uint8 `ndr:"ref,size_is=CbBufSize"`
	PcbBytesNeeded     msscmr.LPBOUNDED_DWORD_256K
	LpServicesReturned msscmr.LPBOUNDED_DWORD_256K
	LpResumeIndex      *ndr.DWORD `ndr:"unique"`
	Status             ndr.DWORD  `ndr:"retval"`
}

// REnumServicesStatusExA calls REnumServicesStatusExA (opnum 41) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func REnumServicesStatusExA(rpc ndr.Invoker, hSCManager msscmr.SC_RPC_HANDLE, infoLevel msscmr.SC_ENUM_TYPE, dwServiceType ndr.DWORD, dwServiceState ndr.DWORD, cbBufSize ndr.DWORD, lpResumeIndex *ndr.DWORD, pszGroupName *ndr.STR) (LpBuffer []uint8, PcbBytesNeeded msscmr.LPBOUNDED_DWORD_256K, LpServicesReturned msscmr.LPBOUNDED_DWORD_256K, LpResumeIndex *ndr.DWORD, err error) {
	req := &rEnumServicesStatusExARequest{
		HSCManager:     hSCManager,
		InfoLevel:      infoLevel,
		DwServiceType:  dwServiceType,
		DwServiceState: dwServiceState,
		CbBufSize:      cbBufSize,
		LpResumeIndex:  lpResumeIndex,
		PszGroupName:   pszGroupName,
	}
	var resp rEnumServicesStatusExAResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("REnumServicesStatusExA: %w", err)
		return
	}
	LpBuffer = resp.LpBuffer
	PcbBytesNeeded = resp.PcbBytesNeeded
	LpServicesReturned = resp.LpServicesReturned
	LpResumeIndex = resp.LpResumeIndex
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("REnumServicesStatusExA failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
