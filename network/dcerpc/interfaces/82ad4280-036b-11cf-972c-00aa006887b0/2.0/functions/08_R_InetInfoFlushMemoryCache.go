package functions

import (
	"fmt"

	inetinfo "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82ad4280-036b-11cf-972c-00aa006887b0/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_InetInfoFlushMemoryCacheRequest carries the [in] parameters of R_InetInfoFlushMemoryCache.
type r_InetInfoFlushMemoryCacheRequest struct {
	PszServer    *ndr.WSTR `ndr:"unique"`
	DwServerMask ndr.DWORD
}

func (*r_InetInfoFlushMemoryCacheRequest) Opnum() uint16 {
	return inetinfo.OpnumR_InetInfoFlushMemoryCache
}

// r_InetInfoFlushMemoryCacheResponse carries the [out] parameters and return value of R_InetInfoFlushMemoryCache.
type r_InetInfoFlushMemoryCacheResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_InetInfoFlushMemoryCache calls R_InetInfoFlushMemoryCache (opnum 8) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_InetInfoFlushMemoryCache(rpc ndr.Invoker, pszServer *ndr.WSTR, dwServerMask ndr.DWORD) (err error) {
	req := &r_InetInfoFlushMemoryCacheRequest{
		PszServer:    pszServer,
		DwServerMask: dwServerMask,
	}
	var resp r_InetInfoFlushMemoryCacheResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_InetInfoFlushMemoryCache: %w", err)
		return
	}
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_InetInfoFlushMemoryCache failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}
