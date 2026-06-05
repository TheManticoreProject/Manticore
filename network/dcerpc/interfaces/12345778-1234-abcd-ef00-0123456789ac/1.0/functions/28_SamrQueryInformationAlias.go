package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// samrQueryInformationAliasRequest carries the [in] alias handle and the information class
// that selects the arm of the returned union.
type samrQueryInformationAliasRequest struct {
	AliasHandle           structures.SAMPR_HANDLE
	AliasInformationClass structures.ALIAS_INFORMATION_CLASS
}

func (*samrQueryInformationAliasRequest) Opnum() uint16 {
	return samr.OpnumSamrQueryInformationAlias
}

// samrQueryInformationAliasResponse is the reply: the [out, switch_is] alias info buffer
// (a [unique] pointer to the discriminated union) and the NTSTATUS.
type samrQueryInformationAliasResponse struct {
	Buffer *structures.SAMPR_ALIAS_INFO_BUFFER `ndr:"unique"`
	Status ndr.DWORD                           `ndr:"retval"`
}

// SamrQueryInformationAlias calls SamrQueryInformationAlias (opnum 28), retrieving
// attributes of an alias for the requested information class ([MS-SAMR] 3.1.5.5.4).
func SamrQueryInformationAlias(rpc ndr.Invoker, aliasHandle structures.SAMPR_HANDLE, aliasInformationClass structures.ALIAS_INFORMATION_CLASS) (*structures.SAMPR_ALIAS_INFO_BUFFER, error) {
	req := &samrQueryInformationAliasRequest{
		AliasHandle:           aliasHandle,
		AliasInformationClass: aliasInformationClass,
	}
	var resp samrQueryInformationAliasResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("SamrQueryInformationAlias: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Buffer, fmt.Errorf("SamrQueryInformationAlias failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Buffer, nil
}
