package functions

// IDL source: [MS-WCCE] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-wcce/df709596-4a70-4a26-ada9-76781250700c
// A fetched copy is kept at ms-wcce.idl in the interface directory.

import (
	"fmt"

	ICertPassage "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/91ae6020-9e3c-11cf-8d7c-00aa00c091be/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswcce "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wcce"
)

// certServerRequestRequest carries the [in]/[in,out] parameters of CertServerRequest.
// The [in] handle_t is the explicit binding handle and is not marshalled.
// pwszAuthority is a [unique] string, so a nil pointer is a valid (absent) value.
// pctbAttribs/pctbRequest are [in, ref] single pointers, transmitted inline.
type certServerRequestRequest struct {
	DwFlags       ndr.DWORD
	PwszAuthority *ndr.WSTR `ndr:"unique"`
	PdwRequestId  ndr.DWORD
	PctbAttribs   mswcce.CERTTRANSBLOB
	PctbRequest   mswcce.CERTTRANSBLOB
}

func (*certServerRequestRequest) Opnum() uint16 { return ICertPassage.OpnumCertServerRequest }

// certServerRequestResponse carries the [out]/[in,out] parameters and return value.
// PdwRequestId is [in, out] so it appears in both the request and the response.
type certServerRequestResponse struct {
	PdwRequestId           ndr.DWORD
	PdwDisposition         ndr.DWORD
	PctbCert               mswcce.CERTTRANSBLOB
	PctbEncodedCert        mswcce.CERTTRANSBLOB
	PctbDispositionMessage mswcce.CERTTRANSBLOB
	Status                 ndr.DWORD `ndr:"retval"`
}

// CertServerRequest calls CertServerRequest (opnum 0) ([MS-ICPR] 3.2.4.1). It submits
// a certificate enrollment request (pctbRequest, with optional pctbAttribs attributes)
// to the CA named by pwszAuthority and returns the CA disposition plus the issued
// certificate blobs. dwFlags carries the request flags; pdwRequestId is [in, out] (the
// pending request id on input, the assigned id on output).
//
// The return value is the CA disposition/HRESULT-style status; callers inspect
// pdwDisposition and the returned status together.
func CertServerRequest(rpc ndr.Invoker, dwFlags ndr.DWORD, pwszAuthority *ndr.WSTR, pdwRequestId ndr.DWORD, pctbAttribs mswcce.CERTTRANSBLOB, pctbRequest mswcce.CERTTRANSBLOB) (PdwRequestId ndr.DWORD, PdwDisposition ndr.DWORD, PctbCert mswcce.CERTTRANSBLOB, PctbEncodedCert mswcce.CERTTRANSBLOB, PctbDispositionMessage mswcce.CERTTRANSBLOB, err error) {
	req := &certServerRequestRequest{
		DwFlags:       dwFlags,
		PwszAuthority: pwszAuthority,
		PdwRequestId:  pdwRequestId,
		PctbAttribs:   pctbAttribs,
		PctbRequest:   pctbRequest,
	}
	var resp certServerRequestResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("CertServerRequest: %w", err)
		return
	}
	PdwRequestId = resp.PdwRequestId
	PdwDisposition = resp.PdwDisposition
	PctbCert = resp.PctbCert
	PctbEncodedCert = resp.PctbEncodedCert
	PctbDispositionMessage = resp.PctbDispositionMessage
	if uint32(resp.Status) != ICertPassage.StatusSuccess {
		err = fmt.Errorf("CertServerRequest failed: %s", ICertPassage.StatusString(uint32(resp.Status)))
	}
	return
}
