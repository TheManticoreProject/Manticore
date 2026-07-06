package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	TermSrvSession "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/484809d6-4239-471b-b5bc-61df8c23ac48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcShowMessageBoxRequest carries the [in] parameters of RpcShowMessageBox.
type rpcShowMessageBoxRequest struct {
	HSession   mststs.SESSION_HANDLE
	SzTitle    ndr.WSTR
	SzMessage  ndr.WSTR
	UlStyle    ndr.DWORD
	UlTimeout  ndr.DWORD
	BDoNotWait ndr.BOOL
}

func (*rpcShowMessageBoxRequest) Opnum() uint16 { return TermSrvSession.OpnumRpcShowMessageBox }

// rpcShowMessageBoxResponse carries the [out] parameters and return value of RpcShowMessageBox.
type rpcShowMessageBoxResponse struct {
	PulResponse ndr.DWORD
	Status      ndr.DWORD `ndr:"retval"`
}

// RpcShowMessageBox calls RpcShowMessageBox (opnum 9) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcShowMessageBox(rpc ndr.Invoker, hSession mststs.SESSION_HANDLE, szTitle ndr.WSTR, szMessage ndr.WSTR, ulStyle ndr.DWORD, ulTimeout ndr.DWORD, bDoNotWait ndr.BOOL) (PulResponse ndr.DWORD, err error) {
	req := &rpcShowMessageBoxRequest{
		HSession:   hSession,
		SzTitle:    szTitle,
		SzMessage:  szMessage,
		UlStyle:    ulStyle,
		UlTimeout:  ulTimeout,
		BDoNotWait: bDoNotWait,
	}
	var resp rpcShowMessageBoxResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcShowMessageBox: %w", err)
		return
	}
	PulResponse = resp.PulResponse
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcShowMessageBox failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
