package functions

// IDL source: [MS-DHCPM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dhcpm/d1932d08-3249-44cb-90f1-8661f8fb5b90
// A fetched copy is kept at ms-dhcpm.idl in the interface directory.

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpServerSetConfigVQRequest carries the [in] parameters of R_DhcpServerSetConfigVQ.
type r_DhcpServerSetConfigVQRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	FieldsToSet     ndr.DWORD
	ConfigInfo      msdhcpm.DHCP_SERVER_CONFIG_INFO_VQ
}

func (*r_DhcpServerSetConfigVQRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpServerSetConfigVQ }

// r_DhcpServerSetConfigVQResponse carries the [out] parameters and return value of R_DhcpServerSetConfigVQ.
type r_DhcpServerSetConfigVQResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpServerSetConfigVQ calls R_DhcpServerSetConfigVQ (opnum 41) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpServerSetConfigVQ(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, fieldsToSet ndr.DWORD, configInfo msdhcpm.DHCP_SERVER_CONFIG_INFO_VQ) (err error) {
	req := &r_DhcpServerSetConfigVQRequest{
		ServerIpAddress: serverIpAddress,
		FieldsToSet:     fieldsToSet,
		ConfigInfo:      configInfo,
	}
	var resp r_DhcpServerSetConfigVQResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpServerSetConfigVQ: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpServerSetConfigVQ failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
