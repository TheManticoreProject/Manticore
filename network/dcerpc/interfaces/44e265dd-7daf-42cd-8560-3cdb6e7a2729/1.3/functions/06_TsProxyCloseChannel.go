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

// tsProxyCloseChannelRequest carries the [in] parameters of TsProxyCloseChannel.
type tsProxyCloseChannelRequest struct {
	Context mstsgu.PCHANNEL_CONTEXT_HANDLE_NOSERIALIZE
}

func (*tsProxyCloseChannelRequest) Opnum() uint16 {
	return tsgu.OpnumTsProxyCloseChannel
}

// tsProxyCloseChannelResponse carries the [out] parameters and return value of TsProxyCloseChannel.
type tsProxyCloseChannelResponse struct {
	Context mstsgu.PCHANNEL_CONTEXT_HANDLE_NOSERIALIZE
	Status  ndr.DWORD `ndr:"retval"`
}

// TsProxyCloseChannel calls TsProxyCloseChannel (opnum 6) ([MS-TSGU] — verify the parameter
// modeling and status handling).
func TsProxyCloseChannel(rpc ndr.Invoker, context mstsgu.PCHANNEL_CONTEXT_HANDLE_NOSERIALIZE) (Context mstsgu.PCHANNEL_CONTEXT_HANDLE_NOSERIALIZE, err error) {
	req := &tsProxyCloseChannelRequest{
		Context: context,
	}
	var resp tsProxyCloseChannelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("TsProxyCloseChannel: %w", err)
		return
	}
	Context = resp.Context
	if uint32(resp.Status) != tsgu.StatusSuccess {
		err = fmt.Errorf("TsProxyCloseChannel failed: %s", tsgu.StatusString(uint32(resp.Status)))
	}
	return
}
