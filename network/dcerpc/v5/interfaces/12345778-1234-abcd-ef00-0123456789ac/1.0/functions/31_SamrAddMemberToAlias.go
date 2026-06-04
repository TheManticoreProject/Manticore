package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrAddMemberToAliasRequest carries the [in] alias handle and the [ref] SID of the member
// to add (transmitted inline).
type samrAddMemberToAliasRequest struct {
	AliasHandle structures.SAMPR_HANDLE
	MemberId    dtyp.RPC_SID
}

func (*samrAddMemberToAliasRequest) Opnum() uint16 { return samr.OpnumSamrAddMemberToAlias }

// SamrAddMemberToAlias calls SamrAddMemberToAlias (opnum 31), adding a member to an alias
// ([MS-SAMR] 3.1.5.4.2).
func SamrAddMemberToAlias(rpc *client.Client, aliasHandle structures.SAMPR_HANDLE, memberId dtyp.RPC_SID) error {
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
