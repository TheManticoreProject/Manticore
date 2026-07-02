package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarAddPrivilegesToAccountRequest is the [in] parameter set of
// LsarAddPrivilegesToAccount: an open account handle and the inline set of privileges to
// add (a top-level [ref] struct).
type lsarAddPrivilegesToAccountRequest struct {
	AccountHandle mslsad.LSAPR_HANDLE
	Privileges    mslsad.LSAPR_PRIVILEGE_SET
}

func (*lsarAddPrivilegesToAccountRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarAddPrivilegesToAccount
}

// LsarAddPrivilegesToAccount calls LsarAddPrivilegesToAccount (opnum 19), adding the given
// privileges to the account ([MS-LSAD] 3.1.4.5.5).
func LsarAddPrivilegesToAccount(rpc ndr.Invoker, accountHandle mslsad.LSAPR_HANDLE, privileges mslsad.LSAPR_PRIVILEGE_SET) error {
	req := &lsarAddPrivilegesToAccountRequest{AccountHandle: accountHandle, Privileges: privileges}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarAddPrivilegesToAccount: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarAddPrivilegesToAccount failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
