package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// dsrAddressToSiteNamesExWRequest carries the [in] parameters of DsrAddressToSiteNamesExW.
type dsrAddressToSiteNamesExWRequest struct {
	ComputerName    *ndr.WSTR `ndr:"unique"`
	EntryCount      ndr.DWORD
	SocketAddresses []msnrpc.NL_SOCKET_ADDRESS `ndr:"ref,size_is=EntryCount"`
}

func (*dsrAddressToSiteNamesExWRequest) Opnum() uint16 { return logon.OpnumDsrAddressToSiteNamesExW }

// dsrAddressToSiteNamesExWResponse carries the [out] parameters and return value of DsrAddressToSiteNamesExW.
type dsrAddressToSiteNamesExWResponse struct {
	SiteNames *msnrpc.NL_SITE_NAME_EX_ARRAY `ndr:"unique"`
	Status    ndr.DWORD                     `ndr:"retval"`
}

// DsrAddressToSiteNamesExW calls DsrAddressToSiteNamesExW (opnum 37) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func DsrAddressToSiteNamesExW(rpc ndr.Invoker, computerName *ndr.WSTR, entryCount ndr.DWORD, socketAddresses []msnrpc.NL_SOCKET_ADDRESS) (SiteNames *msnrpc.NL_SITE_NAME_EX_ARRAY, err error) {
	req := &dsrAddressToSiteNamesExWRequest{
		ComputerName:    computerName,
		EntryCount:      entryCount,
		SocketAddresses: socketAddresses,
	}
	var resp dsrAddressToSiteNamesExWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("DsrAddressToSiteNamesExW: %w", err)
		return
	}
	SiteNames = resp.SiteNames
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("DsrAddressToSiteNamesExW failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
