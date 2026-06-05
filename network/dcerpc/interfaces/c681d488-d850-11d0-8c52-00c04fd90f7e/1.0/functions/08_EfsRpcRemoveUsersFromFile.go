package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcRemoveUsersFromFileRequest carries the [in] parameters of EfsRpcRemoveUsersFromFile.
type efsRpcRemoveUsersFromFileRequest struct {
	FileName *ndr.WSTR `ndr:"unique"`
	Users    structures.ENCRYPTION_CERTIFICATE_HASH_LIST
}

func (*efsRpcRemoveUsersFromFileRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcRemoveUsersFromFile }

// efsRpcRemoveUsersFromFileResponse carries the [out] parameters and return value of EfsRpcRemoveUsersFromFile.
type efsRpcRemoveUsersFromFileResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EfsRpcRemoveUsersFromFile calls EfsRpcRemoveUsersFromFile (opnum 8) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcRemoveUsersFromFile(rpc ndr.Invoker, fileName *ndr.WSTR, users structures.ENCRYPTION_CERTIFICATE_HASH_LIST) (err error) {
	req := &efsRpcRemoveUsersFromFileRequest{
		FileName: fileName,
		Users:    users,
	}
	var resp efsRpcRemoveUsersFromFileResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcRemoveUsersFromFile: %w", err)
		return
	}
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcRemoveUsersFromFile failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
