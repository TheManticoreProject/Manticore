package functions

import (
	"fmt"

	LocToLoc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e33c0cc4-0482-101a-bc0c-02608c6ba218/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrpcl "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rpcl"
)

// i_nsi_lookup_nextRequest carries the [in] parameters of I_nsi_lookup_next.
type i_nsi_lookup_nextRequest struct {
	Import_context msrpcl.NSI_NS_HANDLE_T
}

func (*i_nsi_lookup_nextRequest) Opnum() uint16 { return LocToLoc.OpnumI_nsi_lookup_next }

// i_nsi_lookup_nextResponse carries the [out] parameters of I_nsi_lookup_next. The
// method returns void; its trailing [out] unsigned short *status is the NSI status.
// Binding_vector is [out] NSI_BINDING_VECTOR_P_T *binding_vector: the pointee is the
// [unique] NSI_BINDING_VECTOR_P_T pointer, hence the unique tag on the *T field.
type i_nsi_lookup_nextResponse struct {
	Binding_vector msrpcl.NSI_BINDING_VECTOR_P_T `ndr:"unique"`
	Status         uint16
}

// I_nsi_lookup_next calls I_nsi_lookup_next (opnum 2) ([MS-RPCL] 3.1.4.3).
func I_nsi_lookup_next(rpc ndr.Invoker, import_context msrpcl.NSI_NS_HANDLE_T) (Binding_vector msrpcl.NSI_BINDING_VECTOR_P_T, Status uint16, err error) {
	req := &i_nsi_lookup_nextRequest{
		Import_context: import_context,
	}
	var resp i_nsi_lookup_nextResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("I_nsi_lookup_next: %w", err)
		return
	}
	Binding_vector = resp.Binding_vector
	Status = resp.Status
	// NSI_S_NO_MORE_BINDINGS (0x1) is the normal terminal status of the lookup
	// enumeration (empty binding_vector), not a failure ([MS-RPCL] 3.1.4.3): callers
	// loop on I_nsi_lookup_next until they see it, so only other nonzero values are errors.
	switch uint32(resp.Status) {
	case LocToLoc.NSI_S_OK, LocToLoc.NSI_S_NO_MORE_BINDINGS:
	default:
		err = fmt.Errorf("I_nsi_lookup_next failed: %s", LocToLoc.StatusString(uint32(resp.Status)))
	}
	return
}
