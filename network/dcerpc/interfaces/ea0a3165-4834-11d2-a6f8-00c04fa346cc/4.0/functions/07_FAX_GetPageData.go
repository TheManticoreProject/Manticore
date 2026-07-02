package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_GetPageDataRequest carries the [in] parameters of FAX_GetPageData.
type fAX_GetPageDataRequest struct {
	JobId       ndr.DWORD
	ImageWidth  ndr.DWORD
	ImageHeight ndr.DWORD
}

func (*fAX_GetPageDataRequest) Opnum() uint16 { return fax.OpnumFAX_GetPageData }

// fAX_GetPageDataResponse carries the [out] parameters and return value of FAX_GetPageData.
type fAX_GetPageDataResponse struct {
	Buffer      []byte `ndr:"unique,conformant"`
	BufferSize  ndr.DWORD
	ImageWidth  ndr.DWORD
	ImageHeight ndr.DWORD
	Status      ndr.DWORD `ndr:"retval"`
}

// FAX_GetPageData calls FAX_GetPageData (opnum 7) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetPageData(rpc ndr.Invoker, jobId ndr.DWORD, imageWidth ndr.DWORD, imageHeight ndr.DWORD) (Buffer []byte, BufferSize ndr.DWORD, ImageWidth ndr.DWORD, ImageHeight ndr.DWORD, err error) {
	req := &fAX_GetPageDataRequest{
		JobId:       jobId,
		ImageWidth:  imageWidth,
		ImageHeight: imageHeight,
	}
	var resp fAX_GetPageDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetPageData: %w", err)
		return
	}
	Buffer = resp.Buffer
	BufferSize = resp.BufferSize
	ImageWidth = resp.ImageWidth
	ImageHeight = resp.ImageHeight
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetPageData failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
