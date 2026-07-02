package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_SendDocumentExRequest carries the [in] parameters of FAX_SendDocumentEx.
type fAX_SendDocumentExRequest struct {
	LpcwstrFileName  *ndr.WSTR `ndr:"unique"`
	LpcCoverPageInfo msfax.FAX_COVERPAGE_INFO_EXW
	LpcSenderProfile uint8
	DwNumRecipients  ndr.DWORD
	LpcRecipientList []*uint8 `ndr:"elem=unique,ref,size_is=DwNumRecipients"`
	LpJobParams      msfax.FAX_JOB_PARAM_EXW
	LpdwJobId        *ndr.DWORD `ndr:"unique"`
}

func (*fAX_SendDocumentExRequest) Opnum() uint16 { return fax.OpnumFAX_SendDocumentEx }

// fAX_SendDocumentExResponse carries the [out] parameters and return value of FAX_SendDocumentEx.
type fAX_SendDocumentExResponse struct {
	LpdwJobId                *ndr.DWORD `ndr:"unique"`
	LpdwlMessageId           uint64
	LpdwlRecipientMessageIds []uint64  `ndr:"ref,size_is=DwNumRecipients"`
	Status                   ndr.DWORD `ndr:"retval"`
}

// FAX_SendDocumentEx calls FAX_SendDocumentEx (opnum 27) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SendDocumentEx(rpc ndr.Invoker, lpcwstrFileName *ndr.WSTR, lpcCoverPageInfo msfax.FAX_COVERPAGE_INFO_EXW, lpcSenderProfile uint8, dwNumRecipients ndr.DWORD, lpcRecipientList []*uint8, lpJobParams msfax.FAX_JOB_PARAM_EXW, lpdwJobId *ndr.DWORD) (LpdwJobId *ndr.DWORD, LpdwlMessageId uint64, LpdwlRecipientMessageIds []uint64, err error) {
	req := &fAX_SendDocumentExRequest{
		LpcwstrFileName:  lpcwstrFileName,
		LpcCoverPageInfo: lpcCoverPageInfo,
		LpcSenderProfile: lpcSenderProfile,
		DwNumRecipients:  dwNumRecipients,
		LpcRecipientList: lpcRecipientList,
		LpJobParams:      lpJobParams,
		LpdwJobId:        lpdwJobId,
	}
	var resp fAX_SendDocumentExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SendDocumentEx: %w", err)
		return
	}
	LpdwJobId = resp.LpdwJobId
	LpdwlMessageId = resp.LpdwlMessageId
	LpdwlRecipientMessageIds = resp.LpdwlRecipientMessageIds
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SendDocumentEx failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
