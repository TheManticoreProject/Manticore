package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarCreateSecretRequest is the [in] parameter set of LsarCreateSecret: an open policy
// handle, the secret name (a [ref] PRPC_UNICODE_STRING, modeled inline), and the desired
// access mask.
type lsarCreateSecretRequest struct {
	PolicyHandle  structures.LSAPR_HANDLE
	SecretName    dtyp.RPC_UNICODE_STRING
	DesiredAccess ndr.DWORD
}

func (*lsarCreateSecretRequest) Opnum() uint16 { return lsarpc.OpnumLsarCreateSecret }

// LsarCreateSecret calls LsarCreateSecret (opnum 16) and returns a handle to the newly
// created secret object ([MS-LSAD] 3.1.4.6.1).
func LsarCreateSecret(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, secretName string, desiredAccess uint32) (structures.LSAPR_HANDLE, error) {
	req := &lsarCreateSecretRequest{
		PolicyHandle:  policyHandle,
		SecretName:    dtyp.NewUnicodeString(secretName),
		DesiredAccess: ndr.DWORD(desiredAccess),
	}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.LSAPR_HANDLE{}, fmt.Errorf("LsarCreateSecret: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarCreateSecret failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
