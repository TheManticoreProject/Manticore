package functions

// IDL source: [MS-FRS1] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-frs1/dd60a0d9-176a-46f4-9904-000172041b92
// A fetched copy is kept at ms-frs1.idl in the interface directory.

import (
	"fmt"

	frsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc59b4-4264-101a-8c59-08002b2f8426/1.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs1 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs1"
)

// frsRpcSendCommPktRequest carries the [in] parameters of FrsRpcSendCommPkt.
type frsRpcSendCommPktRequest struct {
	CommPkt msfrs1.COMM_PACKET
}

func (*frsRpcSendCommPktRequest) Opnum() uint16 { return frsrpc.OpnumFrsRpcSendCommPkt }

// FrsRpcSendCommPkt calls FrsRpcSendCommPkt (opnum 0) ([MS-FRS1] section 3.3.4.4).
func FrsRpcSendCommPkt(rpc ndr.Invoker, commPkt msfrs1.COMM_PACKET) (err error) {
	req := &frsRpcSendCommPktRequest{
		CommPkt: commPkt,
	}
	var resp statusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FrsRpcSendCommPkt: %w", err)
		return
	}
	if uint32(resp.Status) != frsrpc.StatusSuccess {
		err = fmt.Errorf("FrsRpcSendCommPkt failed: %s", frsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
