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

// fAX_CheckServerProtSeqRequest carries the [in] parameters of FAX_CheckServerProtSeq.
type fAX_CheckServerProtSeqRequest struct {
	LpdwProtSeq *ndr.DWORD `ndr:"unique"`
}

func (*fAX_CheckServerProtSeqRequest) Opnum() uint16 { return fax.OpnumFAX_CheckServerProtSeq }

// fAX_CheckServerProtSeqResponse carries the [out] parameters and return value of FAX_CheckServerProtSeq.
type fAX_CheckServerProtSeqResponse struct {
	LpdwProtSeq *ndr.DWORD `ndr:"unique"`
	Status      ndr.DWORD  `ndr:"retval"`
}

// FAX_CheckServerProtSeq calls FAX_CheckServerProtSeq (opnum 26) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_CheckServerProtSeq(rpc ndr.Invoker, lpdwProtSeq *ndr.DWORD) (LpdwProtSeq *ndr.DWORD, err error) {
	req := &fAX_CheckServerProtSeqRequest{
		LpdwProtSeq: lpdwProtSeq,
	}
	var resp fAX_CheckServerProtSeqResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_CheckServerProtSeq: %w", err)
		return
	}
	LpdwProtSeq = resp.LpdwProtSeq
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_CheckServerProtSeq failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
