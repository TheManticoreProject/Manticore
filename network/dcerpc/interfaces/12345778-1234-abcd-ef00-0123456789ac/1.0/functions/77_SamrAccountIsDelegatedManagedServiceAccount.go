package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrAccountIsDelegatedManagedServiceAccountRequest carries the [in] server handle and the
// [ref] account name (inline) being tested.
type samrAccountIsDelegatedManagedServiceAccountRequest struct {
	ServerHandle mssamr.SAMPR_HANDLE
	AccountName  msdtyp.RPC_UNICODE_STRING
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
func SamrAccountIsDelegatedManagedServiceAccount(rpc ndr.Invoker, serverHandle mssamr.SAMPR_HANDLE, accountName msdtyp.RPC_UNICODE_STRING) (bool, bool, error) {
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
