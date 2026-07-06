package functions

// IDL source: [MS-RPCL] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpcl/17f647e6-54e2-4885-a31f-c585086f4783
// A fetched copy is kept at ms-rpcl.idl in the interface directory.

import (
	"fmt"

	LocToLoc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e33c0cc4-0482-101a-bc0c-02608c6ba218/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// i_nsi_ping_locatorRequest carries the [in] parameters of I_nsi_ping_locator.
type i_nsi_ping_locatorRequest struct {
}

func (*i_nsi_ping_locatorRequest) Opnum() uint16 { return LocToLoc.OpnumI_nsi_ping_locator }

// i_nsi_ping_locatorResponse carries the [out] parameter of I_nsi_ping_locator. The
// method returns void; its only [out] error_status_t *status is the RPC status
// (error_status_t is a 32-bit DWORD).
type i_nsi_ping_locatorResponse struct {
	Status ndr.DWORD
}

// I_nsi_ping_locator calls I_nsi_ping_locator (opnum 4) ([MS-RPCL] 3.1.4.5).
func I_nsi_ping_locator(rpc ndr.Invoker) (Status uint32, err error) {
	req := &i_nsi_ping_locatorRequest{}
	var resp i_nsi_ping_locatorResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("I_nsi_ping_locator: %w", err)
		return
	}
	Status = uint32(resp.Status)
	if uint32(resp.Status) != LocToLoc.StatusSuccess {
		err = fmt.Errorf("I_nsi_ping_locator failed: %s", LocToLoc.StatusString(uint32(resp.Status)))
	}
	return
}
