package functions

import (
	"fmt"

	LocToLoc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e33c0cc4-0482-101a-bc0c-02608c6ba218/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrpcl "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rpcl"
)

// i_nsi_lookup_doneRequest carries the [in] parameters of I_nsi_lookup_done.
type i_nsi_lookup_doneRequest struct {
	Import_context msrpcl.NSI_NS_HANDLE_T
}

func (*i_nsi_lookup_doneRequest) Opnum() uint16 { return LocToLoc.OpnumI_nsi_lookup_done }

// i_nsi_lookup_doneResponse carries the [out]/[in,out] parameters of I_nsi_lookup_done.
// The method returns void; its trailing [out] unsigned short *status is the NSI status.
type i_nsi_lookup_doneResponse struct {
	Import_context msrpcl.NSI_NS_HANDLE_T
	Status         uint16
}

// I_nsi_lookup_done calls I_nsi_lookup_done (opnum 1) ([MS-RPCL] 3.1.4.2).
func I_nsi_lookup_done(rpc ndr.Invoker, import_context msrpcl.NSI_NS_HANDLE_T) (Import_context msrpcl.NSI_NS_HANDLE_T, Status uint16, err error) {
	req := &i_nsi_lookup_doneRequest{
		Import_context: import_context,
	}
	var resp i_nsi_lookup_doneResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("I_nsi_lookup_done: %w", err)
		return
	}
	Import_context = resp.Import_context
	Status = resp.Status
	if uint32(resp.Status) != LocToLoc.StatusSuccess {
		err = fmt.Errorf("I_nsi_lookup_done failed: %s", LocToLoc.StatusString(uint32(resp.Status)))
	}
	return
}
