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

// sAGetNSAccountInformationRequest carries the [in]/[in,out] parameters of
// SAGetNSAccountInformation. WszBuffer is the [in,out,size_is(ccBufferSize)] conformant
// buffer the client provides (zeroed) and the server fills; its maximum_count is the
// element count, so CcBufferSize must equal len(WszBuffer).
type sAGetNSAccountInformationRequest struct {
	Handle       *ndr.WSTR `ndr:"unique"`
	CcBufferSize ndr.DWORD
	WszBuffer    []uint16 `ndr:"ref,size_is=CcBufferSize"`
}

func (*sAGetNSAccountInformationRequest) Opnum() uint16 { return sasec.OpnumSAGetNSAccountInformation }

// sAGetNSAccountInformationResponse carries the [in,out] buffer echoed back by the server
// and the HRESULT return value. The conformant array's maximum_count is read from the
// wire, so no size_is sibling is needed here.
type sAGetNSAccountInformationResponse struct {
	WszBuffer []uint16  `ndr:"ref"`
	Status    ndr.DWORD `ndr:"retval"`
}

// SAGetNSAccountInformation calls SAGetNSAccountInformation (opnum 2) ([MS-TSCH]
// 3.2.5.3.6). It retrieves the account name stored for the .NET-services task store. The
// caller passes the buffer size in characters (ccBufferSize, at most MAX_BUFFER_SIZE);
// the stub allocates the [in,out] buffer of that size and returns the NUL-terminated
// account name the server writes into it.
func SAGetNSAccountInformation(rpc ndr.Invoker, handle *ndr.WSTR, ccBufferSize uint32) (string, error) {
	req := &sAGetNSAccountInformationRequest{
		Handle:       handle,
		CcBufferSize: ndr.DWORD(ccBufferSize),
		WszBuffer:    make([]uint16, ccBufferSize),
	}
	var resp sAGetNSAccountInformationResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return "", fmt.Errorf("SAGetNSAccountInformation: %w", err)
	}
	if !sasec.IsSuccess(uint32(resp.Status)) {
		return "", fmt.Errorf("SAGetNSAccountInformation failed: %s", sasec.StatusString(uint32(resp.Status)))
	}
	return decodeWideBuffer(resp.WszBuffer), nil
}
