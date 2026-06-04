package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarOpenSecretRequest is the [in] parameter set of LsarOpenSecret: an open policy
// handle, the secret name (a [ref] PRPC_UNICODE_STRING, modeled inline), and the desired
// access mask.
type lsarOpenSecretRequest struct {
	PolicyHandle  structures.LSAPR_HANDLE
	SecretName    dtyp.RPC_UNICODE_STRING
	DesiredAccess ndr.DWORD
}

func (*lsarOpenSecretRequest) Opnum() uint16 { return lsarpc.OpnumLsarOpenSecret }

// LsarOpenSecret calls LsarOpenSecret (opnum 28) and returns a handle to the named secret
// object ([MS-LSAD] 3.1.4.6.2).
func LsarOpenSecret(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, secretName string, desiredAccess uint32) (structures.LSAPR_HANDLE, error) {
	req := &lsarOpenSecretRequest{
		PolicyHandle:  policyHandle,
		SecretName:    dtyp.NewUnicodeString(secretName),
		DesiredAccess: ndr.DWORD(desiredAccess),
	}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.LSAPR_HANDLE{}, fmt.Errorf("LsarOpenSecret: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarOpenSecret failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
