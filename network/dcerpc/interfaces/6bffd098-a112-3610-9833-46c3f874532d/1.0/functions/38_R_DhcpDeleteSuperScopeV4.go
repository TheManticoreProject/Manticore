package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpDeleteSuperScopeV4Request carries the [in] parameters of R_DhcpDeleteSuperScopeV4.
type r_DhcpDeleteSuperScopeV4Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SuperScopeName  ndr.WSTR
}

func (*r_DhcpDeleteSuperScopeV4Request) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpDeleteSuperScopeV4 }

// r_DhcpDeleteSuperScopeV4Response carries the [out] parameters and return value of R_DhcpDeleteSuperScopeV4.
type r_DhcpDeleteSuperScopeV4Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpDeleteSuperScopeV4 calls R_DhcpDeleteSuperScopeV4 (opnum 38) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpDeleteSuperScopeV4(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, superScopeName ndr.WSTR) (err error) {
	req := &r_DhcpDeleteSuperScopeV4Request{
		ServerIpAddress: serverIpAddress,
		SuperScopeName:  superScopeName,
	}
	var resp r_DhcpDeleteSuperScopeV4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpDeleteSuperScopeV4: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpDeleteSuperScopeV4 failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
