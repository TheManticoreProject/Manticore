package functions

import (
	"fmt"

	browser "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-012892020162/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msbrwsa "github.com/TheManticoreProject/Manticore/windows/protocols/ms-brwsa"
)

// i_BrowserrQueryOtherDomainsRequest carries the [in] and [in,out] parameters of
// I_BrowserrQueryOtherDomains: the optional [unique] server name (ignored on receipt) and
// the [in,out] enumeration structure whose Level selects the info arm.
type i_BrowserrQueryOtherDomainsRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	InfoStruct msbrwsa.SERVER_ENUM_STRUCT
}

func (*i_BrowserrQueryOtherDomainsRequest) Opnum() uint16 {
	return browser.OpnumI_BrowserrQueryOtherDomains
}

// i_BrowserrQueryOtherDomainsResponse carries the updated [in,out] structure, the [out]
// total entry count, and the NET_API_STATUS return value.
type i_BrowserrQueryOtherDomainsResponse struct {
	InfoStruct   msbrwsa.SERVER_ENUM_STRUCT
	TotalEntries ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// I_BrowserrQueryOtherDomains calls I_BrowserrQueryOtherDomains (opnum 2), returning the
// list of other domains configured for the target computer ([MS-BRWSA] 3.1.4.1). The
// client SHOULD send this only to a primary domain controller acting as the domain master
// browser. serverName is ignored on receipt (pass "" for NULL). Level MUST be 100; any
// other value yields ERROR_INVALID_LEVEL. ERROR_MORE_DATA means not all entries were
// returned and is not treated as a failure.
func I_BrowserrQueryOtherDomains(rpc ndr.Invoker, serverName string, info msbrwsa.SERVER_ENUM_STRUCT) (msbrwsa.SERVER_ENUM_STRUCT, uint32, error) {
	req := &i_BrowserrQueryOtherDomainsRequest{
		ServerName: optWStr(serverName),
		InfoStruct: info,
	}
	var resp i_BrowserrQueryOtherDomainsResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return msbrwsa.SERVER_ENUM_STRUCT{}, 0, fmt.Errorf("I_BrowserrQueryOtherDomains: %w", err)
	}
	status := uint32(resp.Status)
	if status != browser.NERR_Success && status != browser.ERROR_MORE_DATA {
		return resp.InfoStruct, uint32(resp.TotalEntries), fmt.Errorf("I_BrowserrQueryOtherDomains failed: %s", browser.StatusString(status))
	}
	return resp.InfoStruct, uint32(resp.TotalEntries), nil
}
