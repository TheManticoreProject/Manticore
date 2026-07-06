package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrRemoveMemberFromAliasRequest carries the [in] alias handle and the [ref] SID of the
// member to remove (transmitted inline).
type samrRemoveMemberFromAliasRequest struct {
	AliasHandle mssamr.SAMPR_HANDLE
	MemberId    msdtyp.RPC_SID
}

func (*samrRemoveMemberFromAliasRequest) Opnum() uint16 {
	return samr.OpnumSamrRemoveMemberFromAlias
}

// SamrRemoveMemberFromAlias calls SamrRemoveMemberFromAlias (opnum 32), removing a member
// from an alias ([MS-SAMR] 3.1.5.5.3).
func SamrRemoveMemberFromAlias(rpc ndr.Invoker, aliasHandle mssamr.SAMPR_HANDLE, memberId msdtyp.RPC_SID) error {
	req := &samrRemoveMemberFromAliasRequest{
		AliasHandle: aliasHandle,
		MemberId:    memberId,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrRemoveMemberFromAlias: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrRemoveMemberFromAlias failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}
