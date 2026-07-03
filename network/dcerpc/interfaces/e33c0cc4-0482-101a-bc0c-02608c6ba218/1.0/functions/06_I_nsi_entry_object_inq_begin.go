package functions

import (
	"fmt"

	LocToLoc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e33c0cc4-0482-101a-bc0c-02608c6ba218/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrpcl "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rpcl"
)

// i_nsi_entry_object_inq_beginRequest carries the [in] parameters of I_nsi_entry_object_inq_begin.
type i_nsi_entry_object_inq_beginRequest struct {
	EntryNameSyntax ndr.DWORD
	EntryName       *ndr.WSTR `ndr:"unique"`
}

func (*i_nsi_entry_object_inq_beginRequest) Opnum() uint16 {
	return LocToLoc.OpnumI_nsi_entry_object_inq_begin
}

// i_nsi_entry_object_inq_beginResponse carries the [out] parameters of
// I_nsi_entry_object_inq_begin. The method returns void; its trailing
// [out] unsigned short *status is the NSI status.
type i_nsi_entry_object_inq_beginResponse struct {
	InqContext msrpcl.NSI_NS_HANDLE_T
	Status     uint16
}

// I_nsi_entry_object_inq_begin calls I_nsi_entry_object_inq_begin (opnum 6) ([MS-RPCL] 3.1.4.7).
func I_nsi_entry_object_inq_begin(rpc ndr.Invoker, entryNameSyntax ndr.DWORD, entryName *ndr.WSTR) (InqContext msrpcl.NSI_NS_HANDLE_T, Status uint16, err error) {
	req := &i_nsi_entry_object_inq_beginRequest{
		EntryNameSyntax: entryNameSyntax,
		EntryName:       entryName,
	}
	var resp i_nsi_entry_object_inq_beginResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("I_nsi_entry_object_inq_begin: %w", err)
		return
	}
	InqContext = resp.InqContext
	Status = resp.Status
	if uint32(resp.Status) != LocToLoc.StatusSuccess {
		err = fmt.Errorf("I_nsi_entry_object_inq_begin failed: %s", LocToLoc.StatusString(uint32(resp.Status)))
	}
	return
}
