package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCloseNetInterfaceRequest carries the [in] parameters of ApiCloseNetInterface.
type apiCloseNetInterfaceRequest struct {
	NetInterface mscmrp.HNETINTERFACE_RPC
}

func (*apiCloseNetInterfaceRequest) Opnum() uint16 { return clusapi.OpnumApiCloseNetInterface }

// apiCloseNetInterfaceResponse carries the [out] parameters and return value of ApiCloseNetInterface.
type apiCloseNetInterfaceResponse struct {
	NetInterface mscmrp.HNETINTERFACE_RPC
	Status       ndr.DWORD `ndr:"retval"`
}

// ApiCloseNetInterface calls ApiCloseNetInterface (opnum 93) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCloseNetInterface(rpc ndr.Invoker, netInterface mscmrp.HNETINTERFACE_RPC) (NetInterface mscmrp.HNETINTERFACE_RPC, err error) {
	req := &apiCloseNetInterfaceRequest{
		NetInterface: netInterface,
	}
	var resp apiCloseNetInterfaceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCloseNetInterface: %w", err)
		return
	}
	NetInterface = resp.NetInterface
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCloseNetInterface failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
