package functions

import (
	"fmt"

	tsgu "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/44e265dd-7daf-42cd-8560-3cdb6e7a2729/1.3"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstsgu "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsgu"
)

// tsProxyMakeTunnelCallRequest carries the [in] parameters of TsProxyMakeTunnelCall.
type tsProxyMakeTunnelCallRequest struct {
	TunnelContext mstsgu.PTUNNEL_CONTEXT_HANDLE_NOSERIALIZE
	ProcId        ndr.DWORD
	TSGPacket     mstsgu.TSG_PACKET
}

func (*tsProxyMakeTunnelCallRequest) Opnum() uint16 {
	return tsgu.OpnumTsProxyMakeTunnelCall
}

// tsProxyMakeTunnelCallResponse carries the [out] parameters and return value of TsProxyMakeTunnelCall.
type tsProxyMakeTunnelCallResponse struct {
	TSGPacketResponse *mstsgu.TSG_PACKET `ndr:"unique"`
	Status            ndr.DWORD          `ndr:"retval"`
}

// TsProxyMakeTunnelCall calls TsProxyMakeTunnelCall (opnum 3) ([MS-TSGU] — verify the parameter
// modeling and status handling).
func TsProxyMakeTunnelCall(rpc ndr.Invoker, tunnelContext mstsgu.PTUNNEL_CONTEXT_HANDLE_NOSERIALIZE, procId ndr.DWORD, tSGPacket mstsgu.TSG_PACKET) (TSGPacketResponse *mstsgu.TSG_PACKET, err error) {
	req := &tsProxyMakeTunnelCallRequest{
		TunnelContext: tunnelContext,
		ProcId:        procId,
		TSGPacket:     tSGPacket,
	}
	var resp tsProxyMakeTunnelCallResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("TsProxyMakeTunnelCall: %w", err)
		return
	}
	TSGPacketResponse = resp.TSGPacketResponse
	if uint32(resp.Status) != tsgu.StatusSuccess {
		err = fmt.Errorf("TsProxyMakeTunnelCall failed: %s", tsgu.StatusString(uint32(resp.Status)))
	}
	return
}
