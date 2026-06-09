package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rEnumDependentServicesWRequest carries the [in] parameters of REnumDependentServicesW.
type rEnumDependentServicesWRequest struct {
	HService       structures.SC_RPC_HANDLE
	DwServiceState ndr.DWORD
	CbBufSize      ndr.DWORD
}

func (*rEnumDependentServicesWRequest) Opnum() uint16 { return svcctl.OpnumREnumDependentServicesW }

// rEnumDependentServicesWResponse carries the [out] parameters and return value of REnumDependentServicesW.
type rEnumDependentServicesWResponse struct {
	LpServices         []uint8 `ndr:"ref,size_is=CbBufSize"`
	PcbBytesNeeded     structures.LPBOUNDED_DWORD_256K
	LpServicesReturned structures.LPBOUNDED_DWORD_256K
	Status             ndr.DWORD `ndr:"retval"`
}

// REnumDependentServicesW calls REnumDependentServicesW (opnum 13) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func REnumDependentServicesW(rpc ndr.Invoker, hService structures.SC_RPC_HANDLE, dwServiceState ndr.DWORD, cbBufSize ndr.DWORD) (LpServices []uint8, PcbBytesNeeded structures.LPBOUNDED_DWORD_256K, LpServicesReturned structures.LPBOUNDED_DWORD_256K, err error) {
	req := &rEnumDependentServicesWRequest{
		HService:       hService,
		DwServiceState: dwServiceState,
		CbBufSize:      cbBufSize,
	}
	var resp rEnumDependentServicesWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("REnumDependentServicesW: %w", err)
		return
	}
	LpServices = resp.LpServices
	PcbBytesNeeded = resp.PcbBytesNeeded
	LpServicesReturned = resp.LpServicesReturned
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("REnumDependentServicesW failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
