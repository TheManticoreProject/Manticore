package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarEnumerateAccountsWithUserRightRequest is the [in] parameter set of
// LsarEnumerateAccountsWithUserRight: an open policy handle and a [unique] pointer to the
// user-right name to filter on (a NULL pointer enumerates all accounts with any right).
type lsarEnumerateAccountsWithUserRightRequest struct {
	PolicyHandle mslsad.LSAPR_HANDLE
	UserRight    *dtyp.RPC_UNICODE_STRING `ndr:"unique"`
}

func (*lsarEnumerateAccountsWithUserRightRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarEnumerateAccountsWithUserRight
}

// lsarEnumerateAccountsWithUserRightResponse is the reply: the [out] buffer of account SIDs
// holding the right (an inline [ref] struct) followed by the NTSTATUS return value.
type lsarEnumerateAccountsWithUserRightResponse struct {
	EnumerationBuffer mslsad.LSAPR_ACCOUNT_ENUM_BUFFER
	Status            ndr.DWORD `ndr:"retval"`
}

// LsarEnumerateAccountsWithUserRight calls LsarEnumerateAccountsWithUserRight (opnum 35),
// returning the SIDs of accounts that hold the named user right; a nil userRight enumerates
// accounts holding any right ([MS-LSAD] 3.1.4.5.10).
func LsarEnumerateAccountsWithUserRight(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, userRight *dtyp.RPC_UNICODE_STRING) (mslsad.LSAPR_ACCOUNT_ENUM_BUFFER, error) {
	req := &lsarEnumerateAccountsWithUserRightRequest{PolicyHandle: policyHandle, UserRight: userRight}
	var resp lsarEnumerateAccountsWithUserRightResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mslsad.LSAPR_ACCOUNT_ENUM_BUFFER{}, fmt.Errorf("LsarEnumerateAccountsWithUserRight: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.EnumerationBuffer, fmt.Errorf("LsarEnumerateAccountsWithUserRight failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.EnumerationBuffer, nil
}
