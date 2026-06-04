package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrAccountIsDelegatedManagedServiceAccountRequest carries the [in] server handle and the
// [ref] account name (inline) being tested.
type samrAccountIsDelegatedManagedServiceAccountRequest struct {
	ServerHandle structures.SAMPR_HANDLE
	AccountName  dtyp.RPC_UNICODE_STRING
}

func (*samrAccountIsDelegatedManagedServiceAccountRequest) Opnum() uint16 {
	return samr.OpnumSamrAccountIsDelegatedManagedServiceAccount
}

// samrAccountIsDelegatedManagedServiceAccountResponse is the reply: the [out] BOOLEAN result
// (whether the account is a delegated MSA), the [out] BOOLEAN authorization flag, and the
// NTSTATUS.
type samrAccountIsDelegatedManagedServiceAccountResponse struct {
	Result     bool
	Authorized bool
	Status     ndr.DWORD `ndr:"retval"`
}

// SamrAccountIsDelegatedManagedServiceAccount calls
// SamrAccountIsDelegatedManagedServiceAccount (opnum 77), reporting whether the named account
// is a delegated managed service account and whether the caller is authorized for it
// ([MS-SAMR] 3.1.5.13.6).
func SamrAccountIsDelegatedManagedServiceAccount(rpc *client.Client, serverHandle structures.SAMPR_HANDLE, accountName dtyp.RPC_UNICODE_STRING) (bool, bool, error) {
	req := &samrAccountIsDelegatedManagedServiceAccountRequest{
		ServerHandle: serverHandle,
		AccountName:  accountName,
	}
	var resp samrAccountIsDelegatedManagedServiceAccountResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return false, false, fmt.Errorf("SamrAccountIsDelegatedManagedServiceAccount: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Result, resp.Authorized, fmt.Errorf("SamrAccountIsDelegatedManagedServiceAccount failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Result, resp.Authorized, nil
}
