package functions

// IDL source: [MS-SRVS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/77aacc74-f8f9-4b46-b2d8-bfe04a7d9c44
// A fetched copy is kept at ms-srvs.idl in the interface directory.

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// netrpSetFileSecurityRequest is the [in] parameter set of NetrpSetFileSecurity: the
// [unique] server name, the [unique] share name, the (ref) file name, the
// SECURITY_INFORMATION bits, and the inline ADT_SECURITY_DESCRIPTOR (a single [in] ref
// pointer in the IDL).
type netrpSetFileSecurityRequest struct {
	ServerName          *ndr.WSTR `ndr:"unique"`
	ShareName           *ndr.WSTR `ndr:"unique"`
	LpFileName          ndr.WSTR
	SecurityInformation ndr.DWORD
	SecurityDescriptor  mssrvs.ADT_SECURITY_DESCRIPTOR
}

func (*netrpSetFileSecurityRequest) Opnum() uint16 { return srvsvc.OpnumNetrpSetFileSecurity }

// NetrpSetFileSecurity calls NetrpSetFileSecurity (opnum 40), setting the security
// descriptor of a file or directory on a share ([MS-SRVS] 3.1.4.41).
func NetrpSetFileSecurity(rpc ndr.Invoker, serverName, shareName, lpFileName string, securityInformation uint32, securityDescriptor mssrvs.ADT_SECURITY_DESCRIPTOR) error {
	req := &netrpSetFileSecurityRequest{
		ServerName:          optWStr(serverName),
		ShareName:           optWStr(shareName),
		LpFileName:          ndr.WSTR(lpFileName),
		SecurityInformation: ndr.DWORD(securityInformation),
		SecurityDescriptor:  securityDescriptor,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrpSetFileSecurity: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return fmt.Errorf("NetrpSetFileSecurity failed: %s", srvsvc.StatusString(status))
	}
	return nil
}
