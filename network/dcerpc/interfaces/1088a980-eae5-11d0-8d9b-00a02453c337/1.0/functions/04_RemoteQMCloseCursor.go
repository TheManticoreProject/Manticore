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

// remoteQMCloseCursorRequest carries the [in] parameters of RemoteQMCloseCursor.
type remoteQMCloseCursorRequest struct {
	HQueue  ndr.DWORD
	HCursor ndr.DWORD
}

func (*remoteQMCloseCursorRequest) Opnum() uint16 { return qm2qm.OpnumRemoteQMCloseCursor }

// remoteQMCloseCursorResponse carries the [out] parameters and return value of RemoteQMCloseCursor.
type remoteQMCloseCursorResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RemoteQMCloseCursor calls RemoteQMCloseCursor (opnum 4) ([MS-MQQP] 3.1.4.5).
func RemoteQMCloseCursor(rpc ndr.Invoker, hQueue ndr.DWORD, hCursor ndr.DWORD) (err error) {
	req := &remoteQMCloseCursorRequest{
		HQueue:  hQueue,
		HCursor: hCursor,
	}
	var resp remoteQMCloseCursorResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RemoteQMCloseCursor: %w", err)
		return
	}
	if uint32(resp.Status) != qm2qm.StatusSuccess {
		err = fmt.Errorf("RemoteQMCloseCursor failed: %s", qm2qm.StatusString(uint32(resp.Status)))
	}
	return
}
