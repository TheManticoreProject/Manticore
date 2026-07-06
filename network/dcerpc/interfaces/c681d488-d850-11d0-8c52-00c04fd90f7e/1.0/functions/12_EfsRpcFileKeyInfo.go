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

// efsRpcFileKeyInfoRequest carries the [in] parameters of EfsRpcFileKeyInfo.
type efsRpcFileKeyInfoRequest struct {
	FileName  ndr.WSTR
	InfoClass ndr.DWORD
}

func (*efsRpcFileKeyInfoRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcFileKeyInfo }

// efsRpcFileKeyInfoResponse carries the [out] parameters and return value of EfsRpcFileKeyInfo.
type efsRpcFileKeyInfoResponse struct {
	KeyInfo *msefsr.EFS_RPC_BLOB `ndr:"unique"`
	Status  ndr.DWORD            `ndr:"retval"`
}

// EfsRpcFileKeyInfo calls EfsRpcFileKeyInfo (opnum 12) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcFileKeyInfo(rpc ndr.Invoker, fileName ndr.WSTR, infoClass ndr.DWORD) (KeyInfo *msefsr.EFS_RPC_BLOB, err error) {
	req := &efsRpcFileKeyInfoRequest{
		FileName:  fileName,
		InfoClass: infoClass,
	}
	var resp efsRpcFileKeyInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcFileKeyInfo: %w", err)
		return
	}
	KeyInfo = resp.KeyInfo
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcFileKeyInfo failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
