package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrSetInformationAliasRequest carries the [in] alias handle, the information class, and
// the [in, switch_is] alias info buffer (the discriminated union, transmitted inline).
type samrSetInformationAliasRequest struct {
	AliasHandle           mssamr.SAMPR_HANDLE
	AliasInformationClass mssamr.ALIAS_INFORMATION_CLASS `ndr:"enum"`
	Buffer                mssamr.SAMPR_ALIAS_INFO_BUFFER
}

func (*samrSetInformationAliasRequest) Opnum() uint16 { return samr.OpnumSamrSetInformationAlias }

// SamrSetInformationAlias calls SamrSetInformationAlias (opnum 29), updating attributes of
// an alias for the supplied information class ([MS-SAMR] 3.1.5.6.4).
func SamrSetInformationAlias(rpc ndr.Invoker, aliasHandle mssamr.SAMPR_HANDLE, aliasInformationClass mssamr.ALIAS_INFORMATION_CLASS, buffer mssamr.SAMPR_ALIAS_INFO_BUFFER) error {
	req := &samrSetInformationAliasRequest{
		AliasHandle:           aliasHandle,
		AliasInformationClass: aliasInformationClass,
		Buffer:                buffer,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrSetInformationAlias: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrSetInformationAlias failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}
