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

// remoteQMStartReceiveRequest carries the [in] parameters of RemoteQMStartReceive.
type remoteQMStartReceiveRequest struct {
	LpRemoteReadDesc msmqqp.REMOTEREADDESC
}

func (*remoteQMStartReceiveRequest) Opnum() uint16 { return qm2qm.OpnumRemoteQMStartReceive }

// remoteQMStartReceiveResponse carries the [out] parameters and return value of RemoteQMStartReceive.
type remoteQMStartReceiveResponse struct {
	PphContext       msmqqp.PCTX_REMOTEREAD_HANDLE_TYPE
	LpRemoteReadDesc msmqqp.REMOTEREADDESC
	Status           ndr.DWORD `ndr:"retval"`
}

// RemoteQMStartReceive calls RemoteQMStartReceive (opnum 0) ([MS-MQQP] 3.1.4.1).
func RemoteQMStartReceive(rpc ndr.Invoker, lpRemoteReadDesc msmqqp.REMOTEREADDESC) (PphContext msmqqp.PCTX_REMOTEREAD_HANDLE_TYPE, LpRemoteReadDesc msmqqp.REMOTEREADDESC, err error) {
	req := &remoteQMStartReceiveRequest{
		LpRemoteReadDesc: lpRemoteReadDesc,
	}
	var resp remoteQMStartReceiveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RemoteQMStartReceive: %w", err)
		return
	}
	PphContext = resp.PphContext
	LpRemoteReadDesc = resp.LpRemoteReadDesc
	if uint32(resp.Status) != qm2qm.StatusSuccess {
		err = fmt.Errorf("RemoteQMStartReceive failed: %s", qm2qm.StatusString(uint32(resp.Status)))
	}
	return
}
