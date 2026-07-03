package functions

import (
	"fmt"

	LocToLoc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e33c0cc4-0482-101a-bc0c-02608c6ba218/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrpcl "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rpcl"
)

// i_nsi_entry_object_inq_nextRequest carries the [in] parameters of I_nsi_entry_object_inq_next.
type i_nsi_entry_object_inq_nextRequest struct {
	InqContext msrpcl.NSI_NS_HANDLE_T
}

func (*i_nsi_entry_object_inq_nextRequest) Opnum() uint16 {
	return LocToLoc.OpnumI_nsi_entry_object_inq_next
}

// i_nsi_entry_object_inq_nextResponse carries the [out] parameters of
// I_nsi_entry_object_inq_next. The method returns void; its trailing
// [out] unsigned short *status is the NSI status. Uuid_vec is
// [out] NSI_UUID_VECTOR_P_T *uuid_vec: the pointee is the [unique]
// NSI_UUID_VECTOR_P_T pointer, hence the unique tag on the *T field.
type i_nsi_entry_object_inq_nextResponse struct {
	Uuid_vec msrpcl.NSI_UUID_VECTOR_P_T `ndr:"unique"`
	Status   uint16
}

// I_nsi_entry_object_inq_next calls I_nsi_entry_object_inq_next (opnum 3) ([MS-RPCL] 3.1.4.4).
func I_nsi_entry_object_inq_next(rpc ndr.Invoker, inqContext msrpcl.NSI_NS_HANDLE_T) (Uuid_vec msrpcl.NSI_UUID_VECTOR_P_T, Status uint16, err error) {
	req := &i_nsi_entry_object_inq_nextRequest{
		InqContext: inqContext,
	}
	var resp i_nsi_entry_object_inq_nextResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("I_nsi_entry_object_inq_next: %w", err)
		return
	}
	Uuid_vec = resp.Uuid_vec
	Status = resp.Status
	if uint32(resp.Status) != LocToLoc.StatusSuccess {
		err = fmt.Errorf("I_nsi_entry_object_inq_next failed: %s", LocToLoc.StatusString(uint32(resp.Status)))
	}
	return
}
