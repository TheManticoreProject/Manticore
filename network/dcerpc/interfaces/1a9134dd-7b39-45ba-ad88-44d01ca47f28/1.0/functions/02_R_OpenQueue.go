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

// r_OpenQueueRequest carries the [in] parameters of R_OpenQueue.
type r_OpenQueueRequest struct {
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

func (*r_OpenQueueRequest) Opnum() uint16 { return RemoteRead.OpnumR_OpenQueue }

// r_OpenQueueResponse carries the [out] parameters of R_OpenQueue. The IDL method is declared
// `void` ([MS-MQRR] section 3.1.4.2), so there is NO return value on the wire — only the
// output context handle is marshaled back.
type r_OpenQueueResponse struct {
	PphContext msmqrr.QUEUE_CONTEXT_HANDLE_SERIALIZE
}

// R_OpenQueue calls R_OpenQueue (opnum 2) ([MS-MQRR] section 3.1.4.2). The method returns
// void; the only error path is the underlying RPC transport.
func R_OpenQueue(rpc ndr.Invoker, pQueueFormat msmqmq.QUEUE_FORMAT, dwAccess ndr.DWORD, dwShareMode ndr.DWORD, pClientId guid.GUID, fNonRoutingServer int32, major uint8, minor uint8, buildNumber uint16, fWorkgroup int32) (PphContext msmqrr.QUEUE_CONTEXT_HANDLE_SERIALIZE, err error) {
	req := &r_OpenQueueRequest{
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
	var resp r_OpenQueueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_OpenQueue: %w", err)
		return
	}
	PphContext = resp.PphContext
	return
}
