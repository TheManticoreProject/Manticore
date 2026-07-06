package functions

// IDL source: [MS-FRS2] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-frs2/39bd498b-2a94-41b7-914e-10921d543912
// A fetched copy is kept at ms-frs2.idl in the interface directory.

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// establishSessionRequest carries the [in] parameters of EstablishSession.
type establishSessionRequest struct {
	ConnectionId msfrs2.FRS_CONNECTION_ID
	ContentSetId msfrs2.FRS_CONTENT_SET_ID
}

func (*establishSessionRequest) Opnum() uint16 { return FrsTransport.OpnumEstablishSession }

// establishSessionResponse carries the [out] parameters and return value of EstablishSession.
type establishSessionResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EstablishSession calls EstablishSession (opnum 2) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func EstablishSession(rpc ndr.Invoker, connectionId msfrs2.FRS_CONNECTION_ID, contentSetId msfrs2.FRS_CONTENT_SET_ID) (err error) {
	req := &establishSessionRequest{
		ConnectionId: connectionId,
		ContentSetId: contentSetId,
	}
	var resp establishSessionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EstablishSession: %w", err)
		return
	}
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("EstablishSession failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}
