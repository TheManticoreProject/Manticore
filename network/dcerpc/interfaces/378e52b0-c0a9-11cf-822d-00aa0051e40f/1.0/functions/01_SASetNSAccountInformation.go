package functions

// IDL source: [MS-TSCH] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsch/6fc1f51a-26ec-43fa-a8bd-1c364657f110
// A fetched copy is kept at ms-tsch.idl in the interface directory.

import (
	"fmt"

	sasec "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/378e52b0-c0a9-11cf-822d-00aa0051e40f/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// sASetNSAccountInformationRequest carries the [in] parameters of SASetNSAccountInformation.
type sASetNSAccountInformationRequest struct {
	Handle       *ndr.WSTR `ndr:"unique"`
	PwszAccount  *ndr.WSTR `ndr:"unique"`
	PwszPassword *ndr.WSTR `ndr:"unique"`
}

func (*sASetNSAccountInformationRequest) Opnum() uint16 { return sasec.OpnumSASetNSAccountInformation }

// sASetNSAccountInformationResponse carries the [out] parameters and return value of SASetNSAccountInformation.
type sASetNSAccountInformationResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// SASetNSAccountInformation calls SASetNSAccountInformation (opnum 1) ([MS-TSCH] section 3.2.5.3.5).
func SASetNSAccountInformation(rpc ndr.Invoker, handle *ndr.WSTR, pwszAccount *ndr.WSTR, pwszPassword *ndr.WSTR) (err error) {
	req := &sASetNSAccountInformationRequest{
		Handle:       handle,
		PwszAccount:  pwszAccount,
		PwszPassword: pwszPassword,
	}
	var resp sASetNSAccountInformationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SASetNSAccountInformation: %w", err)
		return
	}
	if !sasec.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SASetNSAccountInformation failed: %s", sasec.StatusString(uint32(resp.Status)))
	}
	return
}
