package functions

// IDL source: [MS-FASP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fasp/1503b9d7-7fec-4793-9972-6ad58720c9db
// A fetched copy is kept at ms-fasp.idl in the interface directory.

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRPC_FWRestoreDefaultsRequest carries the [in] parameters of RRPC_FWRestoreDefaults.
type rRPC_FWRestoreDefaultsRequest struct {
}

func (*rRPC_FWRestoreDefaultsRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWRestoreDefaults }

// rRPC_FWRestoreDefaultsResponse carries the [out] parameters and return value of RRPC_FWRestoreDefaults.
type rRPC_FWRestoreDefaultsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWRestoreDefaults calls RRPC_FWRestoreDefaults (opnum 2) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWRestoreDefaults(rpc ndr.Invoker) (err error) {
	req := &rRPC_FWRestoreDefaultsRequest{}
	var resp rRPC_FWRestoreDefaultsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWRestoreDefaults: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWRestoreDefaults failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
