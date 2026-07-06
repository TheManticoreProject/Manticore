package functions

// IDL source: [MS-RPCL] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpcl/17f647e6-54e2-4885-a31f-c585086f4783
// A fetched copy is kept at ms-rpcl.idl in the interface directory.

import (
	"fmt"

	LocToLoc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e33c0cc4-0482-101a-bc0c-02608c6ba218/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrpcl "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rpcl"
)

// i_nsi_entry_object_inq_doneRequest carries the [in] parameters of I_nsi_entry_object_inq_done.
type i_nsi_entry_object_inq_doneRequest struct {
	InqContext msrpcl.NSI_NS_HANDLE_T
}

func (*i_nsi_entry_object_inq_doneRequest) Opnum() uint16 {
	return LocToLoc.OpnumI_nsi_entry_object_inq_done
}

// i_nsi_entry_object_inq_doneResponse carries the [out]/[in,out] parameters of
// I_nsi_entry_object_inq_done. The method returns void; its trailing
// [out] unsigned short *status is the NSI status.
type i_nsi_entry_object_inq_doneResponse struct {
	InqContext msrpcl.NSI_NS_HANDLE_T
	Status     uint16
}

// I_nsi_entry_object_inq_done calls I_nsi_entry_object_inq_done (opnum 5) ([MS-RPCL] 3.1.4.6).
func I_nsi_entry_object_inq_done(rpc ndr.Invoker, inqContext msrpcl.NSI_NS_HANDLE_T) (InqContext msrpcl.NSI_NS_HANDLE_T, Status uint16, err error) {
	req := &i_nsi_entry_object_inq_doneRequest{
		InqContext: inqContext,
	}
	var resp i_nsi_entry_object_inq_doneResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("I_nsi_entry_object_inq_done: %w", err)
		return
	}
	InqContext = resp.InqContext
	Status = resp.Status
	if uint32(resp.Status) != LocToLoc.StatusSuccess {
		err = fmt.Errorf("I_nsi_entry_object_inq_done failed: %s", LocToLoc.StatusString(uint32(resp.Status)))
	}
	return
}
