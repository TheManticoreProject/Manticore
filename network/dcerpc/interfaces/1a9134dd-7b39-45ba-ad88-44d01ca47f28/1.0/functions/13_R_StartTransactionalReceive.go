package functions

// IDL source: [MS-MQMQ] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqmq/56cc73e0-f57a-4bd9-aa45-861be5b85049
// A fetched copy is kept at ms-mqmq.idl in the interface directory.

import (
	"fmt"

	RemoteRead "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a9134dd-7b39-45ba-ad88-44d01ca47f28/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
	msmqrr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqrr"
)

// r_StartTransactionalReceiveRequest carries the [in] parameters of R_StartTransactionalReceive.
type r_StartTransactionalReceiveRequest struct {
	PhContext                msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE
	LookupId                 uint64
	HCursor                  ndr.DWORD
	UlAction                 ndr.DWORD
	UlTimeout                ndr.DWORD
	DwRequestId              ndr.DWORD
	DwMaxBodySize            ndr.DWORD
	DwMaxCompoundMessageSize ndr.DWORD
	PTransactionId           msmqmq.XACTUOW
}

func (*r_StartTransactionalReceiveRequest) Opnum() uint16 {
	return RemoteRead.OpnumR_StartTransactionalReceive
}

// r_StartTransactionalReceiveResponse carries the [out] parameters and return value of R_StartTransactionalReceive.
type r_StartTransactionalReceiveResponse struct {
	PdwArriveTime       ndr.DWORD
	PSequenceId         uint64
	PdwNumberOfSections ndr.DWORD
	// See R_StartReceive: [out, size_is(, *pdwNumberOfSections)] SectionBuffer** is a [unique]
	// pointer to a conformant array of SectionBuffer values -> []SectionBuffer, unique+size_is.
	PpPacketSections []msmqrr.SectionBuffer `ndr:"unique,size_is=PdwNumberOfSections"`
	Status           ndr.DWORD              `ndr:"retval"`
}

// R_StartTransactionalReceive calls R_StartTransactionalReceive (opnum 13) ([MS-MQRR] — verify the parameter
// modeling and status handling).
func R_StartTransactionalReceive(rpc ndr.Invoker, phContext msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE, lookupId uint64, hCursor ndr.DWORD, ulAction ndr.DWORD, ulTimeout ndr.DWORD, dwRequestId ndr.DWORD, dwMaxBodySize ndr.DWORD, dwMaxCompoundMessageSize ndr.DWORD, pTransactionId msmqmq.XACTUOW) (PdwArriveTime ndr.DWORD, PSequenceId uint64, PdwNumberOfSections ndr.DWORD, PpPacketSections []msmqrr.SectionBuffer, err error) {
	req := &r_StartTransactionalReceiveRequest{
		PhContext:                phContext,
		LookupId:                 lookupId,
		HCursor:                  hCursor,
		UlAction:                 ulAction,
		UlTimeout:                ulTimeout,
		DwRequestId:              dwRequestId,
		DwMaxBodySize:            dwMaxBodySize,
		DwMaxCompoundMessageSize: dwMaxCompoundMessageSize,
		PTransactionId:           pTransactionId,
	}
	var resp r_StartTransactionalReceiveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_StartTransactionalReceive: %w", err)
		return
	}
	PdwArriveTime = resp.PdwArriveTime
	PSequenceId = resp.PSequenceId
	PdwNumberOfSections = resp.PdwNumberOfSections
	PpPacketSections = resp.PpPacketSections
	if uint32(resp.Status) != RemoteRead.StatusSuccess {
		err = fmt.Errorf("R_StartTransactionalReceive failed: %s", RemoteRead.StatusString(uint32(resp.Status)))
	}
	return
}
