package functions

import (
	"fmt"

	sasec "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/378e52b0-c0a9-11cf-822d-00aa0051e40f/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// sAGetAccountInformationRequest carries the [in]/[in,out] parameters of
// SAGetAccountInformation. WszBuffer is the [in,out,size_is(ccBufferSize)] conformant
// buffer the client provides (zeroed) and the server fills; its maximum_count is the
// element count, so CcBufferSize must equal len(WszBuffer).
type sAGetAccountInformationRequest struct {
	Handle       *ndr.WSTR `ndr:"unique"`
	PwszJobName  ndr.WSTR
	CcBufferSize ndr.DWORD
	WszBuffer    []uint16 `ndr:"ref,size_is=CcBufferSize"`
}

func (*sAGetAccountInformationRequest) Opnum() uint16 { return sasec.OpnumSAGetAccountInformation }

// sAGetAccountInformationResponse carries the [in,out] buffer echoed back by the server
// and the HRESULT return value. The conformant array's maximum_count is read from the
// wire, so no size_is sibling is needed here.
type sAGetAccountInformationResponse struct {
	WszBuffer []uint16  `ndr:"ref"`
	Status    ndr.DWORD `ndr:"retval"`
}

// SAGetAccountInformation calls SAGetAccountInformation (opnum 3) ([MS-TSCH] 3.2.5.3.7).
// It retrieves the account name stored for the named .JOB task. The caller passes the
// task name and the buffer size in characters (ccBufferSize, at most MAX_BUFFER_SIZE);
// the stub allocates the [in,out] buffer of that size and returns the NUL-terminated
// account name the server writes into it.
func SAGetAccountInformation(rpc ndr.Invoker, handle *ndr.WSTR, pwszJobName ndr.WSTR, ccBufferSize uint32) (string, error) {
	req := &sAGetAccountInformationRequest{
		Handle:       handle,
		PwszJobName:  pwszJobName,
		CcBufferSize: ndr.DWORD(ccBufferSize),
		WszBuffer:    make([]uint16, ccBufferSize),
	}
	var resp sAGetAccountInformationResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return "", fmt.Errorf("SAGetAccountInformation: %w", err)
	}
	if !sasec.IsSuccess(uint32(resp.Status)) {
		return "", fmt.Errorf("SAGetAccountInformation failed: %s", sasec.StatusString(uint32(resp.Status)))
	}
	return decodeWideBuffer(resp.WszBuffer), nil
}
