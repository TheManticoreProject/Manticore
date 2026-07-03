package functions

import (
	"fmt"

	qm2qm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1088a980-eae5-11d0-8d9b-00a02453c337/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqqp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqqp"
)

// remoteQMStartReceive2Request carries the [in] parameters of RemoteQMStartReceive2.
type remoteQMStartReceive2Request struct {
	LpRemoteReadDesc2 msmqqp.REMOTEREADDESC2
}

func (*remoteQMStartReceive2Request) Opnum() uint16 { return qm2qm.OpnumRemoteQMStartReceive2 }

// remoteQMStartReceive2Response carries the [out] parameters and return value of RemoteQMStartReceive2.
type remoteQMStartReceive2Response struct {
	PphContext        msmqqp.PCTX_REMOTEREAD_HANDLE_TYPE
	LpRemoteReadDesc2 msmqqp.REMOTEREADDESC2
	Status            ndr.DWORD `ndr:"retval"`
}

// RemoteQMStartReceive2 calls RemoteQMStartReceive2 (opnum 9) ([MS-MQQP] 3.1.4.10).
func RemoteQMStartReceive2(rpc ndr.Invoker, lpRemoteReadDesc2 msmqqp.REMOTEREADDESC2) (PphContext msmqqp.PCTX_REMOTEREAD_HANDLE_TYPE, LpRemoteReadDesc2 msmqqp.REMOTEREADDESC2, err error) {
	req := &remoteQMStartReceive2Request{
		LpRemoteReadDesc2: lpRemoteReadDesc2,
	}
	var resp remoteQMStartReceive2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RemoteQMStartReceive2: %w", err)
		return
	}
	PphContext = resp.PphContext
	LpRemoteReadDesc2 = resp.LpRemoteReadDesc2
	if uint32(resp.Status) != qm2qm.StatusSuccess {
		err = fmt.Errorf("RemoteQMStartReceive2 failed: %s", qm2qm.StatusString(uint32(resp.Status)))
	}
	return
}
