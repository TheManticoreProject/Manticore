package functions

// IDL source: [MS-EFSR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-efsr/4a25b8e1-fd90-41b6-9301-62ed71334436
// A fetched copy is kept at ms-efsr.idl in the interface directory.

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msefsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-efsr"
)

// efsRpcQueryProtectorsRequest carries the [in] parameters of EfsRpcQueryProtectors.
type efsRpcQueryProtectorsRequest struct {
	FileName ndr.WSTR
}

func (*efsRpcQueryProtectorsRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcQueryProtectors }

// efsRpcQueryProtectorsResponse carries the [out] parameters and return value of EfsRpcQueryProtectors.
type efsRpcQueryProtectorsResponse struct {
	PpProtectorList *msefsr.ENCRYPTION_PROTECTOR_LIST `ndr:"unique"`
	Status          ndr.DWORD                         `ndr:"retval"`
}

// EfsRpcQueryProtectors calls EfsRpcQueryProtectors (opnum 22) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcQueryProtectors(rpc ndr.Invoker, fileName ndr.WSTR) (PpProtectorList *msefsr.ENCRYPTION_PROTECTOR_LIST, err error) {
	req := &efsRpcQueryProtectorsRequest{
		FileName: fileName,
	}
	var resp efsRpcQueryProtectorsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcQueryProtectors: %w", err)
		return
	}
	PpProtectorList = resp.PpProtectorList
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcQueryProtectors failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
