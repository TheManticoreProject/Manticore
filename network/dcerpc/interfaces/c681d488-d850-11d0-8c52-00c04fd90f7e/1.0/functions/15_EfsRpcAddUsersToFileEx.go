package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcAddUsersToFileExRequest carries the [in] parameters of EfsRpcAddUsersToFileEx.
type efsRpcAddUsersToFileExRequest struct {
	DwFlags                ndr.DWORD
	Reserved               *structures.EFS_RPC_BLOB `ndr:"unique"`
	FileName               *ndr.WSTR                `ndr:"unique"`
	EncryptionCertificates structures.ENCRYPTION_CERTIFICATE_LIST
}

func (*efsRpcAddUsersToFileExRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcAddUsersToFileEx }

// efsRpcAddUsersToFileExResponse carries the [out] parameters and return value of EfsRpcAddUsersToFileEx.
type efsRpcAddUsersToFileExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EfsRpcAddUsersToFileEx calls EfsRpcAddUsersToFileEx (opnum 15) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcAddUsersToFileEx(rpc ndr.Invoker, dwFlags ndr.DWORD, reserved *structures.EFS_RPC_BLOB, fileName *ndr.WSTR, encryptionCertificates structures.ENCRYPTION_CERTIFICATE_LIST) (err error) {
	req := &efsRpcAddUsersToFileExRequest{
		DwFlags:                dwFlags,
		Reserved:               reserved,
		FileName:               fileName,
		EncryptionCertificates: encryptionCertificates,
	}
	var resp efsRpcAddUsersToFileExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcAddUsersToFileEx: %w", err)
		return
	}
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcAddUsersToFileEx failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
