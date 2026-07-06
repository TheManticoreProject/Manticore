package functions

// IDL source: [MS-TSGU] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsgu/ea0ac9e8-2d53-477e-ba57-b1ad01e38039
// A fetched copy is kept at ms-tsgu.idl in the interface directory.

import (
	"fmt"

	tsgu "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/44e265dd-7daf-42cd-8560-3cdb6e7a2729/1.3"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstsgu "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsgu"
)

// tsProxyCloseTunnelRequest carries the [in] parameters of TsProxyCloseTunnel.
type tsProxyCloseTunnelRequest struct {
	Context mstsgu.PTUNNEL_CONTEXT_HANDLE_SERIALIZE
}

func (*tsProxyCloseTunnelRequest) Opnum() uint16 { return tsgu.OpnumTsProxyCloseTunnel }

// tsProxyCloseTunnelResponse carries the [out] parameters and return value of TsProxyCloseTunnel.
type tsProxyCloseTunnelResponse struct {
	Context mstsgu.PTUNNEL_CONTEXT_HANDLE_SERIALIZE
	Status  ndr.DWORD `ndr:"retval"`
}

// TsProxyCloseTunnel calls TsProxyCloseTunnel (opnum 7) ([MS-TSGU] — verify the parameter
// modeling and status handling).
func TsProxyCloseTunnel(rpc ndr.Invoker, context mstsgu.PTUNNEL_CONTEXT_HANDLE_SERIALIZE) (Context mstsgu.PTUNNEL_CONTEXT_HANDLE_SERIALIZE, err error) {
	req := &tsProxyCloseTunnelRequest{
		Context: context,
	}
	var resp tsProxyCloseTunnelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("TsProxyCloseTunnel: %w", err)
		return
	}
	Context = resp.Context
	if uint32(resp.Status) != tsgu.StatusSuccess {
		err = fmt.Errorf("TsProxyCloseTunnel failed: %s", tsgu.StatusString(uint32(resp.Status)))
	}
	return
}
