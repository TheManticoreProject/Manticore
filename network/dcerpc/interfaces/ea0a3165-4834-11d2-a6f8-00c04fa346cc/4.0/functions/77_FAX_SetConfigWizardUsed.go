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

// fAX_SetConfigWizardUsedRequest carries the [in] parameters of FAX_SetConfigWizardUsed.
type fAX_SetConfigWizardUsedRequest struct {
	BConfigWizardUsed ndr.BOOL
}

func (*fAX_SetConfigWizardUsedRequest) Opnum() uint16 { return fax.OpnumFAX_SetConfigWizardUsed }

// fAX_SetConfigWizardUsedResponse carries the [out] parameters and return value of FAX_SetConfigWizardUsed.
type fAX_SetConfigWizardUsedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetConfigWizardUsed calls FAX_SetConfigWizardUsed (opnum 77) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetConfigWizardUsed(rpc ndr.Invoker, bConfigWizardUsed ndr.BOOL) (err error) {
	req := &fAX_SetConfigWizardUsedRequest{
		BConfigWizardUsed: bConfigWizardUsed,
	}
	var resp fAX_SetConfigWizardUsedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetConfigWizardUsed: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetConfigWizardUsed failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
