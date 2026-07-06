package functions

// IDL source: [MS-MQQP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqqp/e3ad0b4f-51ab-4a7c-936b-c4f3e6f57b2d
// A fetched copy is kept at ms-mqqp.idl in the interface directory.

import (
	"fmt"

	qm2qm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1088a980-eae5-11d0-8d9b-00a02453c337/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// remoteQMCancelReceiveRequest carries the [in] parameters of RemoteQMCancelReceive.
type remoteQMCancelReceiveRequest struct {
	HQueue      ndr.DWORD
	PQueue      ndr.DWORD
	DwRequestID ndr.DWORD
}

func (*remoteQMCancelReceiveRequest) Opnum() uint16 { return qm2qm.OpnumRemoteQMCancelReceive }

// remoteQMCancelReceiveResponse carries the [out] parameters and return value of RemoteQMCancelReceive.
type remoteQMCancelReceiveResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RemoteQMCancelReceive calls RemoteQMCancelReceive (opnum 5) ([MS-MQQP] 3.1.4.6).
func RemoteQMCancelReceive(rpc ndr.Invoker, hQueue ndr.DWORD, pQueue ndr.DWORD, dwRequestID ndr.DWORD) (err error) {
	req := &remoteQMCancelReceiveRequest{
		HQueue:      hQueue,
		PQueue:      pQueue,
		DwRequestID: dwRequestID,
	}
	var resp remoteQMCancelReceiveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RemoteQMCancelReceive: %w", err)
		return
	}
	if uint32(resp.Status) != qm2qm.StatusSuccess {
		err = fmt.Errorf("RemoteQMCancelReceive failed: %s", qm2qm.StatusString(uint32(resp.Status)))
	}
	return
}
