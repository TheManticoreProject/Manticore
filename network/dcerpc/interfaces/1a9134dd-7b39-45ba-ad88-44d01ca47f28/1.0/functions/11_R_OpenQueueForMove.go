package functions

// IDL source: [MS-MQMQ] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqmq/56cc73e0-f57a-4bd9-aa45-861be5b85049
// A fetched copy is kept at ms-mqmq.idl in the interface directory.

import (
	"fmt"

	RemoteRead "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a9134dd-7b39-45ba-ad88-44d01ca47f28/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
	msmqrr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqrr"
)

// r_OpenQueueForMoveRequest carries the [in] parameters of R_OpenQueueForMove.
type r_OpenQueueForMoveRequest struct {
	PQueueFormat      msmqmq.QUEUE_FORMAT
	DwAccess          ndr.DWORD
	DwShareMode       ndr.DWORD
	PClientId         guid.GUID
	FNonRoutingServer int32
	Major             uint8
	Minor             uint8
	BuildNumber       uint16
	FWorkgroup        int32
}

func (*r_OpenQueueForMoveRequest) Opnum() uint16 { return RemoteRead.OpnumR_OpenQueueForMove }

// r_OpenQueueForMoveResponse carries the [out] parameters of R_OpenQueueForMove. The IDL
// method is declared `void` ([MS-MQRR] section 3.1.4.11), so there is NO return value on the
// wire — only the move-context cookie and the output context handle are marshaled back.
type r_OpenQueueForMoveResponse struct {
	PMoveContext uint64
	PphContext   msmqrr.QUEUE_CONTEXT_HANDLE_SERIALIZE
}

// R_OpenQueueForMove calls R_OpenQueueForMove (opnum 11) ([MS-MQRR] section 3.1.4.11). The
// method returns void; the only error path is the underlying RPC transport.
func R_OpenQueueForMove(rpc ndr.Invoker, pQueueFormat msmqmq.QUEUE_FORMAT, dwAccess ndr.DWORD, dwShareMode ndr.DWORD, pClientId guid.GUID, fNonRoutingServer int32, major uint8, minor uint8, buildNumber uint16, fWorkgroup int32) (PMoveContext uint64, PphContext msmqrr.QUEUE_CONTEXT_HANDLE_SERIALIZE, err error) {
	req := &r_OpenQueueForMoveRequest{
		PQueueFormat:      pQueueFormat,
		DwAccess:          dwAccess,
		DwShareMode:       dwShareMode,
		PClientId:         pClientId,
		FNonRoutingServer: fNonRoutingServer,
		Major:             major,
		Minor:             minor,
		BuildNumber:       buildNumber,
		FWorkgroup:        fWorkgroup,
	}
	var resp r_OpenQueueForMoveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_OpenQueueForMove: %w", err)
		return
	}
	PMoveContext = resp.PMoveContext
	PphContext = resp.PphContext
	return
}
