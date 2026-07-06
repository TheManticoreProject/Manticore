package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrAddMemberToAliasRequest carries the [in] alias handle and the [ref] SID of the member
// to add (transmitted inline).
type samrAddMemberToAliasRequest struct {
	AliasHandle mssamr.SAMPR_HANDLE
	MemberId    msdtyp.RPC_SID
}

func (*samrAddMemberToAliasRequest) Opnum() uint16 { return samr.OpnumSamrAddMemberToAlias }

// SamrAddMemberToAlias calls SamrAddMemberToAlias (opnum 31), adding a member to an alias
// ([MS-SAMR] 3.1.5.4.2).
func SamrAddMemberToAlias(rpc ndr.Invoker, aliasHandle mssamr.SAMPR_HANDLE, memberId msdtyp.RPC_SID) error {
	req := &samrAddMemberToAliasRequest{
		AliasHandle: aliasHandle,
		MemberId:    memberId,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrAddMemberToAlias: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrAddMemberToAlias failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}
