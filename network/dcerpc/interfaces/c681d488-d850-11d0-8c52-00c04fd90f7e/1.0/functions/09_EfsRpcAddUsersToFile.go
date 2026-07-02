package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msefsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-efsr"
)

// efsRpcAddUsersToFileRequest carries the [in] parameters of EfsRpcAddUsersToFile.
type efsRpcAddUsersToFileRequest struct {
	FileName               ndr.WSTR
	EncryptionCertificates msefsr.ENCRYPTION_CERTIFICATE_LIST
}

func (*efsRpcAddUsersToFileRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcAddUsersToFile }

// efsRpcAddUsersToFileResponse carries the [out] parameters and return value of EfsRpcAddUsersToFile.
type efsRpcAddUsersToFileResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EfsRpcAddUsersToFile calls EfsRpcAddUsersToFile (opnum 9) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcAddUsersToFile(rpc ndr.Invoker, fileName ndr.WSTR, encryptionCertificates msefsr.ENCRYPTION_CERTIFICATE_LIST) (err error) {
	req := &efsRpcAddUsersToFileRequest{
		FileName:               fileName,
		EncryptionCertificates: encryptionCertificates,
	}
	var resp efsRpcAddUsersToFileResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcAddUsersToFile: %w", err)
		return
	}
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcAddUsersToFile failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
