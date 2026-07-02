package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWEnumProductsRequest carries the [in] parameters of RRPC_FWEnumProducts.
type rRPC_FWEnumProductsRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
}

func (*rRPC_FWEnumProductsRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWEnumProducts }

// rRPC_FWEnumProductsResponse carries the [out] parameters and return value of RRPC_FWEnumProducts.
type rRPC_FWEnumProductsResponse struct {
	PdwNumProducts ndr.DWORD
	PpProducts     []*msfasp.FW_PRODUCT `ndr:"elem=unique,ref,conformant"`
	Status         ndr.DWORD            `ndr:"retval"`
}

// RRPC_FWEnumProducts calls RRPC_FWEnumProducts (opnum 31) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWEnumProducts(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE) (PdwNumProducts ndr.DWORD, PpProducts []*msfasp.FW_PRODUCT, err error) {
	req := &rRPC_FWEnumProductsRequest{
		HPolicyStore: hPolicyStore,
	}
	var resp rRPC_FWEnumProductsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWEnumProducts: %w", err)
		return
	}
	PdwNumProducts = resp.PdwNumProducts
	PpProducts = resp.PpProducts
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWEnumProducts failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
