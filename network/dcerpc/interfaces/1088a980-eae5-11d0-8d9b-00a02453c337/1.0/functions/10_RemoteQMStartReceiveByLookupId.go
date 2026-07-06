package functions

// IDL source: [MS-MQQP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqqp/e3ad0b4f-51ab-4a7c-936b-c4f3e6f57b2d
// A fetched copy is kept at ms-mqqp.idl in the interface directory.

import (
	"fmt"

	qm2qm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1088a980-eae5-11d0-8d9b-00a02453c337/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqqp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqqp"
)

// remoteQMStartReceiveByLookupIdRequest carries the [in] parameters of RemoteQMStartReceiveByLookupId.
type remoteQMStartReceiveByLookupIdRequest struct {
	LookupId          uint64
	LpRemoteReadDesc2 msmqqp.REMOTEREADDESC2
}

func (*remoteQMStartReceiveByLookupIdRequest) Opnum() uint16 {
	return qm2qm.OpnumRemoteQMStartReceiveByLookupId
}

// remoteQMStartReceiveByLookupIdResponse carries the [out] parameters and return value of RemoteQMStartReceiveByLookupId.
type remoteQMStartReceiveByLookupIdResponse struct {
	PphContext        msmqqp.PCTX_REMOTEREAD_HANDLE_TYPE
	LpRemoteReadDesc2 msmqqp.REMOTEREADDESC2
	Status            ndr.DWORD `ndr:"retval"`
}

// RemoteQMStartReceiveByLookupId calls RemoteQMStartReceiveByLookupId (opnum 10) ([MS-MQQP] 3.1.4.11).
func RemoteQMStartReceiveByLookupId(rpc ndr.Invoker, lookupId uint64, lpRemoteReadDesc2 msmqqp.REMOTEREADDESC2) (PphContext msmqqp.PCTX_REMOTEREAD_HANDLE_TYPE, LpRemoteReadDesc2 msmqqp.REMOTEREADDESC2, err error) {
	req := &remoteQMStartReceiveByLookupIdRequest{
		LookupId:          lookupId,
		LpRemoteReadDesc2: lpRemoteReadDesc2,
	}
	var resp remoteQMStartReceiveByLookupIdResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RemoteQMStartReceiveByLookupId: %w", err)
		return
	}
	PphContext = resp.PphContext
	LpRemoteReadDesc2 = resp.LpRemoteReadDesc2
	if uint32(resp.Status) != qm2qm.StatusSuccess {
		err = fmt.Errorf("RemoteQMStartReceiveByLookupId failed: %s", qm2qm.StatusString(uint32(resp.Status)))
	}
	return
}
