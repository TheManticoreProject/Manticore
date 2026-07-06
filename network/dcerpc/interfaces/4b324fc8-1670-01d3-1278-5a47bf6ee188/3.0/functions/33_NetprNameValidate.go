package functions

// IDL source: [MS-SRVS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/77aacc74-f8f9-4b46-b2d8-bfe04a7d9c44
// A fetched copy is kept at ms-srvs.idl in the interface directory.

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netprNameValidateRequest is the [in] parameter set of NetprNameValidate: the [unique]
// server name, the (ref) name, the name type, and the validation flags.
type netprNameValidateRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Name       ndr.WSTR
	NameType   ndr.DWORD
	Flags      ndr.DWORD
}

func (*netprNameValidateRequest) Opnum() uint16 { return srvsvc.OpnumNetprNameValidate }

// NetprNameValidate calls NetprNameValidate (opnum 33), verifying that a name is valid for
// the given name type ([MS-SRVS] 3.1.4.30).
func NetprNameValidate(rpc ndr.Invoker, serverName, name string, nameType, flags uint32) error {
	req := &netprNameValidateRequest{
		ServerName: optWStr(serverName),
		Name:       ndr.WSTR(name),
		NameType:   ndr.DWORD(nameType),
		Flags:      ndr.DWORD(flags),
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetprNameValidate: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return fmt.Errorf("NetprNameValidate failed: %s", srvsvc.StatusString(status))
	}
	return nil
}
