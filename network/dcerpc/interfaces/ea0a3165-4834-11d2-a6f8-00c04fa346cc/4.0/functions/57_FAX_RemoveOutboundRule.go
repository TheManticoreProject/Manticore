package functions

// IDL source: [MS-FAX] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fax/0907310f-0693-47e7-a6cb-3e599c89a1dd
// A fetched copy is kept at ms-fax.idl in the interface directory.

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_RemoveOutboundRuleRequest carries the [in] parameters of FAX_RemoveOutboundRule.
type fAX_RemoveOutboundRuleRequest struct {
	DwAreaCode    ndr.DWORD
	DwCountryCode ndr.DWORD
}

func (*fAX_RemoveOutboundRuleRequest) Opnum() uint16 { return fax.OpnumFAX_RemoveOutboundRule }

// fAX_RemoveOutboundRuleResponse carries the [out] parameters and return value of FAX_RemoveOutboundRule.
type fAX_RemoveOutboundRuleResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_RemoveOutboundRule calls FAX_RemoveOutboundRule (opnum 57) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_RemoveOutboundRule(rpc ndr.Invoker, dwAreaCode ndr.DWORD, dwCountryCode ndr.DWORD) (err error) {
	req := &fAX_RemoveOutboundRuleRequest{
		DwAreaCode:    dwAreaCode,
		DwCountryCode: dwCountryCode,
	}
	var resp fAX_RemoveOutboundRuleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_RemoveOutboundRule: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_RemoveOutboundRule failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
