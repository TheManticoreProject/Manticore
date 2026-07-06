package functions

// IDL source: [MS-DHCPM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dhcpm/d1932d08-3249-44cb-90f1-8661f8fb5b90
// A fetched copy is kept at ms-dhcpm.idl in the interface directory.

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpSetSuperScopeV4Request carries the [in] parameters of R_DhcpSetSuperScopeV4.
type r_DhcpSetSuperScopeV4Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   ndr.DWORD
	SuperScopeName  *ndr.WSTR `ndr:"unique"`
	ChangeExisting  ndr.BOOL
}

func (*r_DhcpSetSuperScopeV4Request) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpSetSuperScopeV4 }

// r_DhcpSetSuperScopeV4Response carries the [out] parameters and return value of R_DhcpSetSuperScopeV4.
type r_DhcpSetSuperScopeV4Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetSuperScopeV4 calls R_DhcpSetSuperScopeV4 (opnum 36) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetSuperScopeV4(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, superScopeName *ndr.WSTR, changeExisting ndr.BOOL) (err error) {
	req := &r_DhcpSetSuperScopeV4Request{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
		SuperScopeName:  superScopeName,
		ChangeExisting:  changeExisting,
	}
	var resp r_DhcpSetSuperScopeV4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetSuperScopeV4: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetSuperScopeV4 failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
