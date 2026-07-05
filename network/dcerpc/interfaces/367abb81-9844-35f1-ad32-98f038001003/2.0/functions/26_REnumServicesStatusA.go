package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// rEnumServicesStatusARequest carries the [in] parameters of REnumServicesStatusA.
type rEnumServicesStatusARequest struct {
	HSCManager     msscmr.SC_RPC_HANDLE
	DwServiceType  ndr.DWORD
	DwServiceState ndr.DWORD
	CbBufSize      ndr.DWORD
	LpResumeIndex  *ndr.DWORD `ndr:"unique"`
}

func (*rEnumServicesStatusARequest) Opnum() uint16 { return svcctl.OpnumREnumServicesStatusA }

// rEnumServicesStatusAResponse carries the [out] parameters and return value of REnumServicesStatusA.
type rEnumServicesStatusAResponse struct {
	LpBuffer           []uint8 `ndr:"ref,size_is=CbBufSize"`
	PcbBytesNeeded     msscmr.LPBOUNDED_DWORD_256K
	LpServicesReturned msscmr.LPBOUNDED_DWORD_256K
	LpResumeIndex      *ndr.DWORD `ndr:"unique"`
	Status             ndr.DWORD  `ndr:"retval"`
}

// REnumServicesStatusA calls REnumServicesStatusA (opnum 26) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func REnumServicesStatusA(rpc ndr.Invoker, hSCManager msscmr.SC_RPC_HANDLE, dwServiceType ndr.DWORD, dwServiceState ndr.DWORD, cbBufSize ndr.DWORD, lpResumeIndex *ndr.DWORD) (LpBuffer []uint8, PcbBytesNeeded msscmr.LPBOUNDED_DWORD_256K, LpServicesReturned msscmr.LPBOUNDED_DWORD_256K, LpResumeIndex *ndr.DWORD, err error) {
	req := &rEnumServicesStatusARequest{
		HSCManager:     hSCManager,
		DwServiceType:  dwServiceType,
		DwServiceState: dwServiceState,
		CbBufSize:      cbBufSize,
		LpResumeIndex:  lpResumeIndex,
	}
	var resp rEnumServicesStatusAResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("REnumServicesStatusA: %w", err)
		return
	}
	LpBuffer = resp.LpBuffer
	PcbBytesNeeded = resp.PcbBytesNeeded
	LpServicesReturned = resp.LpServicesReturned
	LpResumeIndex = resp.LpResumeIndex
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("REnumServicesStatusA failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
