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

// efsRpcFileKeyInfoExRequest carries the [in] parameters of EfsRpcFileKeyInfoEx.
type efsRpcFileKeyInfoExRequest struct {
	DwFileKeyInfoFlags ndr.DWORD
	Reserved           *msefsr.EFS_RPC_BLOB `ndr:"unique"`
	FileName           ndr.WSTR
	InfoClass          ndr.DWORD
}

func (*efsRpcFileKeyInfoExRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcFileKeyInfoEx }

// efsRpcFileKeyInfoExResponse carries the [out] parameters and return value of EfsRpcFileKeyInfoEx.
type efsRpcFileKeyInfoExResponse struct {
	KeyInfo *msefsr.EFS_RPC_BLOB `ndr:"unique"`
	Status  ndr.DWORD            `ndr:"retval"`
}

// EfsRpcFileKeyInfoEx calls EfsRpcFileKeyInfoEx (opnum 16) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcFileKeyInfoEx(rpc ndr.Invoker, dwFileKeyInfoFlags ndr.DWORD, reserved *msefsr.EFS_RPC_BLOB, fileName ndr.WSTR, infoClass ndr.DWORD) (KeyInfo *msefsr.EFS_RPC_BLOB, err error) {
	req := &efsRpcFileKeyInfoExRequest{
		DwFileKeyInfoFlags: dwFileKeyInfoFlags,
		Reserved:           reserved,
		FileName:           fileName,
		InfoClass:          infoClass,
	}
	var resp efsRpcFileKeyInfoExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcFileKeyInfoEx: %w", err)
		return
	}
	KeyInfo = resp.KeyInfo
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcFileKeyInfoEx failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
