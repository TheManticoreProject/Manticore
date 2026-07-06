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

// sASetAccountInformationRequest carries the [in] parameters of SASetAccountInformation.
type sASetAccountInformationRequest struct {
	Handle       *ndr.WSTR `ndr:"unique"`
	PwszJobName  ndr.WSTR
	PwszAccount  ndr.WSTR
	PwszPassword *ndr.WSTR `ndr:"unique"`
	DwJobFlags   ndr.DWORD
}

func (*sASetAccountInformationRequest) Opnum() uint16 { return sasec.OpnumSASetAccountInformation }

// sASetAccountInformationResponse carries the [out] parameters and return value of SASetAccountInformation.
type sASetAccountInformationResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// SASetAccountInformation calls SASetAccountInformation (opnum 0) ([MS-TSCH] section 3.2.5.3.4).
func SASetAccountInformation(rpc ndr.Invoker, handle *ndr.WSTR, pwszJobName ndr.WSTR, pwszAccount ndr.WSTR, pwszPassword *ndr.WSTR, dwJobFlags ndr.DWORD) (err error) {
	req := &sASetAccountInformationRequest{
		Handle:       handle,
		PwszJobName:  pwszJobName,
		PwszAccount:  pwszAccount,
		PwszPassword: pwszPassword,
		DwJobFlags:   dwJobFlags,
	}
	var resp sASetAccountInformationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SASetAccountInformation: %w", err)
		return
	}
	if !sasec.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SASetAccountInformation failed: %s", sasec.StatusString(uint32(resp.Status)))
	}
	return
}
