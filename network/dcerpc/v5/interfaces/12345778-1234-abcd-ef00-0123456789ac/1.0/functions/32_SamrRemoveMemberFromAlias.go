package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrRemoveMemberFromAliasRequest carries the [in] alias handle and the [ref] SID of the
// member to remove (transmitted inline).
type samrRemoveMemberFromAliasRequest struct {
	AliasHandle structures.SAMPR_HANDLE
	MemberId    dtyp.RPC_SID
}

func (*samrRemoveMemberFromAliasRequest) Opnum() uint16 {
	return samr.OpnumSamrRemoveMemberFromAlias
}

// SamrRemoveMemberFromAlias calls SamrRemoveMemberFromAlias (opnum 32), removing a member
// from an alias ([MS-SAMR] 3.1.5.5.3).
func SamrRemoveMemberFromAlias(rpc *client.Client, aliasHandle structures.SAMPR_HANDLE, memberId dtyp.RPC_SID) error {
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
