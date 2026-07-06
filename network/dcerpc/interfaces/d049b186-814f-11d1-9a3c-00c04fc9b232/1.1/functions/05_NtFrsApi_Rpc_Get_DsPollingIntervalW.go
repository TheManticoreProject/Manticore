package functions

// IDL source: [MS-FRS1] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-frs1/dd60a0d9-176a-46f4-9904-000172041b92
// A fetched copy is kept at ms-frs1.idl in the interface directory.

import (
	"fmt"

	frsapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/d049b186-814f-11d1-9a3c-00c04fc9b232/1.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// ntFrsApi_Rpc_Get_DsPollingIntervalWRequest carries the [in] parameters of NtFrsApi_Rpc_Get_DsPollingIntervalW.
type ntFrsApi_Rpc_Get_DsPollingIntervalWRequest struct {
}

func (*ntFrsApi_Rpc_Get_DsPollingIntervalWRequest) Opnum() uint16 {
	return frsapi.OpnumNtFrsApi_Rpc_Get_DsPollingIntervalW
}

// ntFrsApi_Rpc_Get_DsPollingIntervalWResponse carries the [out] parameters and return value of NtFrsApi_Rpc_Get_DsPollingIntervalW.
type ntFrsApi_Rpc_Get_DsPollingIntervalWResponse struct {
	Interval      ndr.DWORD
	LongInterval  ndr.DWORD
	ShortInterval ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// NtFrsApi_Rpc_Get_DsPollingIntervalW calls NtFrsApi_Rpc_Get_DsPollingIntervalW (opnum 5) ([MS-FRS1] section 3.2.4.2).
func NtFrsApi_Rpc_Get_DsPollingIntervalW(rpc ndr.Invoker) (Interval ndr.DWORD, LongInterval ndr.DWORD, ShortInterval ndr.DWORD, err error) {
	req := &ntFrsApi_Rpc_Get_DsPollingIntervalWRequest{}
	var resp ntFrsApi_Rpc_Get_DsPollingIntervalWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NtFrsApi_Rpc_Get_DsPollingIntervalW: %w", err)
		return
	}
	Interval = resp.Interval
	LongInterval = resp.LongInterval
	ShortInterval = resp.ShortInterval
	if uint32(resp.Status) != frsapi.StatusSuccess {
		err = fmt.Errorf("NtFrsApi_Rpc_Get_DsPollingIntervalW failed: %s", frsapi.StatusString(uint32(resp.Status)))
	}
	return
}
