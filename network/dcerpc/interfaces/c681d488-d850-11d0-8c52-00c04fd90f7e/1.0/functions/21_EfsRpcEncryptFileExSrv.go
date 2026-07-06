package functions

// IDL source: [MS-EFSR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-efsr/4a25b8e1-fd90-41b6-9301-62ed71334436
// A fetched copy is kept at ms-efsr.idl in the interface directory.

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcEncryptFileExSrvRequest carries the [in] parameters of EfsRpcEncryptFileExSrv.
type efsRpcEncryptFileExSrvRequest struct {
	FileName            ndr.WSTR
	ProtectorDescriptor *ndr.WSTR `ndr:"unique"`
	Flags               ndr.DWORD
}

func (*efsRpcEncryptFileExSrvRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcEncryptFileExSrv }

// efsRpcEncryptFileExSrvResponse carries the [out] parameters and return value of EfsRpcEncryptFileExSrv.
type efsRpcEncryptFileExSrvResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EfsRpcEncryptFileExSrv calls EfsRpcEncryptFileExSrv (opnum 21) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcEncryptFileExSrv(rpc ndr.Invoker, fileName ndr.WSTR, protectorDescriptor *ndr.WSTR, flags ndr.DWORD) (err error) {
	req := &efsRpcEncryptFileExSrvRequest{
		FileName:            fileName,
		ProtectorDescriptor: protectorDescriptor,
		Flags:               flags,
	}
	var resp efsRpcEncryptFileExSrvResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcEncryptFileExSrv: %w", err)
		return
	}
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcEncryptFileExSrv failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
