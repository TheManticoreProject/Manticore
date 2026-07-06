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

// fAX_AddOutboundRuleRequest carries the [in] parameters of FAX_AddOutboundRule.
type fAX_AddOutboundRuleRequest struct {
	DwAreaCode      ndr.DWORD
	DwCountryCode   ndr.DWORD
	DwDeviceId      ndr.DWORD
	LpwstrGroupName *ndr.WSTR `ndr:"unique"`
	BUseGroup       ndr.BOOL
}

func (*fAX_AddOutboundRuleRequest) Opnum() uint16 { return fax.OpnumFAX_AddOutboundRule }

// fAX_AddOutboundRuleResponse carries the [out] parameters and return value of FAX_AddOutboundRule.
type fAX_AddOutboundRuleResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_AddOutboundRule calls FAX_AddOutboundRule (opnum 56) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_AddOutboundRule(rpc ndr.Invoker, dwAreaCode ndr.DWORD, dwCountryCode ndr.DWORD, dwDeviceId ndr.DWORD, lpwstrGroupName *ndr.WSTR, bUseGroup ndr.BOOL) (err error) {
	req := &fAX_AddOutboundRuleRequest{
		DwAreaCode:      dwAreaCode,
		DwCountryCode:   dwCountryCode,
		DwDeviceId:      dwDeviceId,
		LpwstrGroupName: lpwstrGroupName,
		BUseGroup:       bUseGroup,
	}
	var resp fAX_AddOutboundRuleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_AddOutboundRule: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_AddOutboundRule failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
