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

// efsRpcGetEncryptedFileMetadataRequest carries the [in] parameters of EfsRpcGetEncryptedFileMetadata.
type efsRpcGetEncryptedFileMetadataRequest struct {
	FileName ndr.WSTR
}

func (*efsRpcGetEncryptedFileMetadataRequest) Opnum() uint16 {
	return efsrpc.OpnumEfsRpcGetEncryptedFileMetadata
}

// efsRpcGetEncryptedFileMetadataResponse carries the [out] parameters and return value of EfsRpcGetEncryptedFileMetadata.
type efsRpcGetEncryptedFileMetadataResponse struct {
	EfsStreamBlob *msefsr.EFS_RPC_BLOB `ndr:"unique"`
	Status        ndr.DWORD            `ndr:"retval"`
}

// EfsRpcGetEncryptedFileMetadata calls EfsRpcGetEncryptedFileMetadata (opnum 18) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcGetEncryptedFileMetadata(rpc ndr.Invoker, fileName ndr.WSTR) (EfsStreamBlob *msefsr.EFS_RPC_BLOB, err error) {
	req := &efsRpcGetEncryptedFileMetadataRequest{
		FileName: fileName,
	}
	var resp efsRpcGetEncryptedFileMetadataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcGetEncryptedFileMetadata: %w", err)
		return
	}
	EfsStreamBlob = resp.EfsStreamBlob
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcGetEncryptedFileMetadata failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
