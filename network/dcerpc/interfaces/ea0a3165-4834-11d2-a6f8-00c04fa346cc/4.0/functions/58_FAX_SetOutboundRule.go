package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_SetOutboundRuleRequest carries the [in] parameters of FAX_SetOutboundRule.
type fAX_SetOutboundRuleRequest struct {
	PRule msfax.RPC_FAX_OUTBOUND_ROUTING_RULEW
}

func (*fAX_SetOutboundRuleRequest) Opnum() uint16 { return fax.OpnumFAX_SetOutboundRule }

// fAX_SetOutboundRuleResponse carries the [out] parameters and return value of FAX_SetOutboundRule.
type fAX_SetOutboundRuleResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetOutboundRule calls FAX_SetOutboundRule (opnum 58) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetOutboundRule(rpc ndr.Invoker, pRule msfax.RPC_FAX_OUTBOUND_ROUTING_RULEW) (err error) {
	req := &fAX_SetOutboundRuleRequest{
		PRule: pRule,
	}
	var resp fAX_SetOutboundRuleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetOutboundRule: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetOutboundRule failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
