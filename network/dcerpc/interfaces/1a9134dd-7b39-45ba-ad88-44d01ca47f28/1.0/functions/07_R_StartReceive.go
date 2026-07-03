package functions

import (
	"fmt"

	RemoteRead "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a9134dd-7b39-45ba-ad88-44d01ca47f28/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqrr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqrr"
)

// r_StartReceiveRequest carries the [in] parameters of R_StartReceive.
type r_StartReceiveRequest struct {
	PhContext                msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE
	LookupId                 uint64
	HCursor                  ndr.DWORD
	UlAction                 ndr.DWORD
	UlTimeout                ndr.DWORD
	DwRequestId              ndr.DWORD
	DwMaxBodySize            ndr.DWORD
	DwMaxCompoundMessageSize ndr.DWORD
}

func (*r_StartReceiveRequest) Opnum() uint16 { return RemoteRead.OpnumR_StartReceive }

// r_StartReceiveResponse carries the [out] parameters and return value of R_StartReceive.
type r_StartReceiveResponse struct {
	PdwArriveTime       ndr.DWORD
	PSequenceId         uint64
	PdwNumberOfSections ndr.DWORD
	// ppPacketSections is [out, size_is(, *pdwNumberOfSections)] SectionBuffer**: the [out]
	// top-level pointer is [ref] (transmitted in place), and it addresses a [unique] pointer
	// to a conformant array of SectionBuffer values sized by *pdwNumberOfSections. That is a
	// unique-pointer-to-conformant-array of value structs -> []SectionBuffer with unique+size_is.
	PpPacketSections []msmqrr.SectionBuffer `ndr:"unique,size_is=PdwNumberOfSections"`
	Status           ndr.DWORD              `ndr:"retval"`
}

// R_StartReceive calls R_StartReceive (opnum 7) ([MS-MQRR] — verify the parameter
// modeling and status handling).
func R_StartReceive(rpc ndr.Invoker, phContext msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE, lookupId uint64, hCursor ndr.DWORD, ulAction ndr.DWORD, ulTimeout ndr.DWORD, dwRequestId ndr.DWORD, dwMaxBodySize ndr.DWORD, dwMaxCompoundMessageSize ndr.DWORD) (PdwArriveTime ndr.DWORD, PSequenceId uint64, PdwNumberOfSections ndr.DWORD, PpPacketSections []msmqrr.SectionBuffer, err error) {
	req := &r_StartReceiveRequest{
		PhContext:                phContext,
		LookupId:                 lookupId,
		HCursor:                  hCursor,
		UlAction:                 ulAction,
		UlTimeout:                ulTimeout,
		DwRequestId:              dwRequestId,
		DwMaxBodySize:            dwMaxBodySize,
		DwMaxCompoundMessageSize: dwMaxCompoundMessageSize,
	}
	var resp r_StartReceiveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_StartReceive: %w", err)
		return
	}
	PdwArriveTime = resp.PdwArriveTime
	PSequenceId = resp.PSequenceId
	PdwNumberOfSections = resp.PdwNumberOfSections
	PpPacketSections = resp.PpPacketSections
	if uint32(resp.Status) != RemoteRead.StatusSuccess {
		err = fmt.Errorf("R_StartReceive failed: %s", RemoteRead.StatusString(uint32(resp.Status)))
	}
	return
}
