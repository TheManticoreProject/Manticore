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

// i_nsi_lookup_beginRequest carries the [in] parameters of I_nsi_lookup_begin.
type i_nsi_lookup_beginRequest struct {
	Entry_name_syntax ndr.DWORD
	Entry_name        *ndr.WSTR                     `ndr:"unique"`
	Interfaceid       *msrpcl.RPC_SYNTAX_IDENTIFIER `ndr:"unique"`
	Xfersyntax        *msrpcl.RPC_SYNTAX_IDENTIFIER `ndr:"unique"`
	Obj_uuid          msrpcl.NSI_UUID_P_T           `ndr:"unique"`
	Binding_max_count ndr.DWORD
	MaxCacheAge       ndr.DWORD
}

func (*i_nsi_lookup_beginRequest) Opnum() uint16 { return LocToLoc.OpnumI_nsi_lookup_begin }

// i_nsi_lookup_beginResponse carries the [out] parameters of I_nsi_lookup_begin. The
// method returns void; its trailing [out] unsigned short *status is the NSI status.
type i_nsi_lookup_beginResponse struct {
	Import_context msrpcl.NSI_NS_HANDLE_T
	Status         uint16
}

// I_nsi_lookup_begin calls I_nsi_lookup_begin (opnum 0) ([MS-RPCL] 3.1.4.1).
func I_nsi_lookup_begin(rpc ndr.Invoker, entry_name_syntax ndr.DWORD, entry_name *ndr.WSTR, interfaceid *msrpcl.RPC_SYNTAX_IDENTIFIER, xfersyntax *msrpcl.RPC_SYNTAX_IDENTIFIER, obj_uuid msrpcl.NSI_UUID_P_T, binding_max_count ndr.DWORD, maxCacheAge ndr.DWORD) (Import_context msrpcl.NSI_NS_HANDLE_T, Status uint16, err error) {
	req := &i_nsi_lookup_beginRequest{
		Entry_name_syntax: entry_name_syntax,
		Entry_name:        entry_name,
		Interfaceid:       interfaceid,
		Xfersyntax:        xfersyntax,
		Obj_uuid:          obj_uuid,
		Binding_max_count: binding_max_count,
		MaxCacheAge:       maxCacheAge,
	}
	var resp i_nsi_lookup_beginResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("I_nsi_lookup_begin: %w", err)
		return
	}
	Import_context = resp.Import_context
	Status = resp.Status
	if uint32(resp.Status) != LocToLoc.StatusSuccess {
		err = fmt.Errorf("I_nsi_lookup_begin failed: %s", LocToLoc.StatusString(uint32(resp.Status)))
	}
	return
}
