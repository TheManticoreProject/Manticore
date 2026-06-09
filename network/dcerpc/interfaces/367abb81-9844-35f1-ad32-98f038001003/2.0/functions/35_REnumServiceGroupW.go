package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rEnumServiceGroupWRequest carries the [in] parameters of REnumServiceGroupW.
type rEnumServiceGroupWRequest struct {
	HSCManager     structures.SC_RPC_HANDLE
	DwServiceType  ndr.DWORD
	DwServiceState ndr.DWORD
	CbBufSize      ndr.DWORD
	LpResumeIndex  *ndr.DWORD `ndr:"unique"`
	PszGroupName   *ndr.WSTR  `ndr:"unique"`
}

func (*rEnumServiceGroupWRequest) Opnum() uint16 { return svcctl.OpnumREnumServiceGroupW }

// rEnumServiceGroupWResponse carries the [out] parameters and return value of REnumServiceGroupW.
type rEnumServiceGroupWResponse struct {
	LpBuffer           []uint8 `ndr:"ref,size_is=CbBufSize"`
	PcbBytesNeeded     structures.LPBOUNDED_DWORD_256K
	LpServicesReturned structures.LPBOUNDED_DWORD_256K
	LpResumeIndex      *ndr.DWORD `ndr:"unique"`
	Status             ndr.DWORD  `ndr:"retval"`
}

// REnumServiceGroupW calls REnumServiceGroupW (opnum 35) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func REnumServiceGroupW(rpc ndr.Invoker, hSCManager structures.SC_RPC_HANDLE, dwServiceType ndr.DWORD, dwServiceState ndr.DWORD, cbBufSize ndr.DWORD, lpResumeIndex *ndr.DWORD, pszGroupName *ndr.WSTR) (LpBuffer []uint8, PcbBytesNeeded structures.LPBOUNDED_DWORD_256K, LpServicesReturned structures.LPBOUNDED_DWORD_256K, LpResumeIndex *ndr.DWORD, err error) {
	req := &rEnumServiceGroupWRequest{
		HSCManager:     hSCManager,
		DwServiceType:  dwServiceType,
		DwServiceState: dwServiceState,
		CbBufSize:      cbBufSize,
		LpResumeIndex:  lpResumeIndex,
		PszGroupName:   pszGroupName,
	}
	var resp rEnumServiceGroupWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("REnumServiceGroupW: %w", err)
		return
	}
	LpBuffer = resp.LpBuffer
	PcbBytesNeeded = resp.PcbBytesNeeded
	LpServicesReturned = resp.LpServicesReturned
	LpResumeIndex = resp.LpResumeIndex
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("REnumServiceGroupW failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
